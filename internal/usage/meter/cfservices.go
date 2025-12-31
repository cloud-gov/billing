package meter

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/usage/node"
	"github.com/cloud-gov/billing/internal/usage/reader"
)

type ServiceMeterDB interface {
	GetCFOrg(ctx context.Context, id pgtype.UUID) (db.CFOrg, error)
}

// CFServiceMeter reads usage from Cloud Foundry service instances.
type CFServiceMeter struct {
	logger *slog.Logger
	client ServiceClient
	dbq    ServiceMeterDB
}

func NewCFServiceMeter(
	logger *slog.Logger, client ServiceClient, dbq ServiceMeterDB,
) *CFServiceMeter {
	return &CFServiceMeter{
		logger: logger.WithGroup("CFServiceMeter"),
		client: client,
		dbq:    dbq,
	}
}

func (m *CFServiceMeter) Name() string {
	return "cfservices"
}

// ReadUsage returns the point-in-time usage of services in Cloud Foundry.
// Returns a non-nil error if there was an error during the overall process of reading usage information from the target system. If individual readings had errors, their errs fields should be set.
func (m *CFServiceMeter) ReadUsage(ctx context.Context) ([]reader.Measurement, []*node.Node, error) {
	m.logger.DebugContext(ctx, "service meter: listing services")
	opts := client.NewServiceInstanceListOptions()
	// Ignore user-provided services, which we do not bill for. IMPORTANT: If this is not set, user-provided services will be included. Some response fields that we assume are non-nil, like .Relationships, will be nil on user-provided services. The code below does not guard against this and will panic.
	opts.Type = "managed"
	si, err := m.client.ServiceInstancesList(ctx, opts)
	if err != nil {
		return nil, nil, err
	}

	m.logger.DebugContext(ctx, "service meter: getting service plans and offerings")
	sp, so, err := m.client.ServicePlansOfferingsList(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	planMap := make(map[string]*resource.ServicePlan, len(sp))
	for _, p := range sp {
		planMap[p.GUID] = p
	}

	offerMap := make(map[string]*resource.ServiceOffering, len(so))
	for _, o := range so {
		offerMap[o.GUID] = o
	}

	m.logger.DebugContext(ctx, "service meter: listing spaces")
	spaceopts := client.NewSpaceListOptions()
	spaces, err := m.client.SpacesList(ctx, spaceopts)
	if err != nil {
		return nil, nil, err
	}
	// TODO: should maybe just use an indexer?
	spaceMap := make(map[string]*resource.Space, len(spaces))
	for _, s := range spaces {
		spaceMap[s.GUID] = s
	}

	spacesToOrgs := make(map[string]*db.CFOrg, len(spaces))
	for _, space := range spaces {
		orgID := pgtype.UUID{}
		if err := orgID.Scan(space.Relationships.Organization.Data.GUID); err != nil {
			return nil, nil, err
		}
		org, err := m.dbq.GetCFOrg(ctx, orgID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, nil, err
			}
			org.ID = orgID // still get an ID if not in our DB
		}
		spacesToOrgs[space.GUID] = &org
	}

	usage := make([]reader.Measurement, len(si))
	nodes := make([]*node.Node, 0, len(si)*3)

	m.logger.DebugContext(ctx, "service meter: aggregating services")
	for i, instance := range si {
		planID := planMap[instance.Relationships.ServicePlan.Data.GUID]
		offrID := offerMap[planID.Relationships.ServiceOffering.Data.GUID]
		spaceID := instance.Relationships.Space.Data.GUID

		org := spacesToOrgs[spaceID]
		orgID := org.ID.String()
		customerID := org.CustomerID

		cfOrgNode, err := node.New(
			customerID,
			orgID,
			node.WithSlugAuto("cforg", org.Name.String),
			node.WithPathAuto("apps.usage"),
		)
		if err != nil {
			return nil, nil, err
		}

		spaceNode, err := node.New(
			customerID,
			spaceID,
			node.WithSlugAuto("space", spaceMap[spaceID].Name),
			node.WithPathByParent(cfOrgNode),
		)
		if err != nil {
			return nil, nil, err
		}

		svcNode, err := node.New(
			customerID,
			instance.GUID,
			node.WithSlugAuto("svc", offrID.Name, planID.Name, instance.Name),
			node.WithPathByParent(spaceNode),
		)
		if err != nil {
			return nil, nil, err
		}

		nodes = append(nodes, []*node.Node{cfOrgNode, spaceNode, svcNode}...)

		usage[i] = reader.Measurement{
			Meter:                 m.Name(),
			CustomerID:            customerID,
			OrgID:                 orgID,
			ResourceKindNaturalID: instance.Relationships.ServicePlan.Data.GUID,
			ResourceNaturalID:     instance.GUID,
			Value:                 1, // For this type of service, 1 indicates it is present at time of reading
			Errs:                  nil,
		}
	}

	return usage, nodes, nil
}
