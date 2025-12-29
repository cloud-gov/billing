package meter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"

	"github.com/cloud-gov/billing/internal/usage/reader"
)

// Abstract to interfaces so we can create a mock client for testing.
type CFSpaceClient interface {
	ListAll(context.Context, *client.SpaceListOptions) ([]*resource.Space, error)
}
type CFServiceInstanceClient interface {
	ListAll(context.Context, *client.ServiceInstanceListOptions) ([]*resource.ServiceInstance, error)
}
type CFServicePlanClient interface {
	ListAll(context.Context, *client.ServicePlanListOptions) ([]*resource.ServicePlan, error)
}
type CFServiceOfferingClient interface {
	ListAll(context.Context, *client.ServiceOfferingListOptions) ([]*resource.ServiceOffering, error)
}

// CFServiceMeter reads usage from Cloud Foundry service instances.
type CFServiceMeter struct {
	logger *slog.Logger

	spaces CFSpaceClient

	services  CFServiceInstanceClient
	plans     CFServicePlanClient
	offerings CFServiceOfferingClient
}

func NewCFServiceMeter(logger *slog.Logger, services CFServiceInstanceClient, spaces CFSpaceClient) *CFServiceMeter {
	return &CFServiceMeter{
		logger:   logger.WithGroup("CFServiceMeter"),
		services: services,
		spaces:   spaces,
	}
}

func (m *CFServiceMeter) Name() string {
	return "cfservices"
}

// ReadUsage returns the point-in-time usage of services in Cloud Foundry.
// Returns a non-nil error if there was an error during the overall process of reading usage information from the target system. If individual readings had errors, their errs fields should be set.
func (m *CFServiceMeter) ReadUsage(ctx context.Context) ([]reader.Measurement, []reader.Node, error) {
	m.logger.DebugContext(ctx, "service meter: listing services")
	opts := client.NewServiceInstanceListOptions()
	// Ignore user-provided services, which we do not bill for. IMPORTANT: If this is not set, user-provided services will be included. Some response fields that we assume are non-nil, like .Relationships, will be nil on user-provided services. The code below does not guard against this and will panic.
	opts.Type = "managed"
	si, err := m.services.ListAll(ctx, opts)
	if err != nil {
		return nil, nil, err
	}

	m.logger.DebugContext(ctx, "service meter: getting service plans")
	sp, err := m.plans.ListAll(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	spMap := make(map[string]*resource.ServicePlan, len(sp))
	for _, p := range sp {
		spMap[p.GUID] = p
	}

	m.logger.DebugContext(ctx, "service meter: getting service offerings")
	so, err := m.offerings.ListAll(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	soMap := make(map[string]*resource.ServiceOffering, len(so))
	for _, o := range so {
		soMap[o.GUID] = o
	}

	m.logger.DebugContext(ctx, "service meter: listing spaces")
	spaceopts := client.NewSpaceListOptions()
	spaces, err := m.spaces.ListAll(ctx, spaceopts)
	if err != nil {
		return nil, nil, err
	}

	spacesToOrgs := make(map[string]string, len(spaces))
	for _, space := range spaces {
		spacesToOrgs[space.GUID] = space.Relationships.Organization.Data.GUID
	}

	usage := make([]reader.Measurement, len(si))
	nodes := make([]reader.Node, len(si)*3) // TODO: return nodes along with usage

	m.logger.DebugContext(ctx, "service meter: aggregating services")
	for i, instance := range si {
		plan := spMap[instance.Relationships.ServicePlan.Data.GUID]
		offr := soMap[plan.GUID]
		slug := fmt.Sprintf("svc_%v_%v_%v", offr.Name, plan.Name, instance.GUID)
		svcInNode := reader.Node{Slug: slug, ResourceNaturalID: instance.GUID}

		// TODO: Get the customer ID here
		space := instance.Relationships.Space.Data.GUID
		cfOrg := spacesToOrgs[space]

		spaceNode := reader.Node{Slug: fmt.Sprintf("space_%v", space), ResourceNaturalID: space}
		cfOrgNode := reader.Node{Slug: fmt.Sprintf("cforg_%v", cfOrg), ResourceNaturalID: cfOrg}

		cfOrgNode.Path = fmt.Sprintf("%v.%v", "xxx.apps.usage", cfOrgNode.Slug)
		spaceNode.Path = fmt.Sprintf("%v.%v", cfOrgNode.Path, spaceNode.Slug)
		svcInNode.Path = fmt.Sprintf("%v.%v", spaceNode.Path, svcInNode.Slug)

		nodes = append(nodes, []reader.Node{cfOrgNode, spaceNode, svcInNode}...)

		usage[i] = reader.Measurement{
			Meter:                 m.Name(),
			OrgID:                 cfOrg,
			ResourceKindNaturalID: instance.Relationships.ServicePlan.Data.GUID,
			ResourceNaturalID:     instance.GUID,
			Value:                 1, // For this type of service, 1 indicates it is present at time of reading
			Errs:                  nil,
		}
	}

	return usage, nodes, nil
}
