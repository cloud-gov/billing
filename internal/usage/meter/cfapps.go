package meter

import (
	"context"
	"errors"
	"log/slog"
	"slices"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/dbx"
	"github.com/cloud-gov/billing/internal/usage/node"
	"github.com/cloud-gov/billing/internal/usage/reader"
)

const appStateStarted = "STARTED"

var (
	ErrAppNotFound   = errors.New("CF processes meter: application not found")
	ErrSpaceNotFound = errors.New("CF processes meter: space not found")
)

type AppMeterDB interface {
	GetCFOrg(ctx context.Context, id pgtype.UUID) (db.CFOrg, error)
}

type CFAppMeter struct {
	logger *slog.Logger
	client AppClient
	dbq    AppMeterDB
}

func NewCFAppMeter(
	logger *slog.Logger, client AppClient, dbq AppMeterDB,
) *CFAppMeter {
	return &CFAppMeter{
		logger: logger.WithGroup("CFAppMeter"),
		client: client,
		dbq:    dbq,
	}
}

func (m *CFAppMeter) Name() string {
	return "cfapps"
}

func (m *CFAppMeter) ReadUsage(ctx context.Context) ([]reader.Measurement, []*node.Node, error) {
	m.logger.DebugContext(ctx, "app meter: listing processes")
	procs, err := m.client.ProcessesList(ctx, client.NewProcessOptions())
	if err != nil {
		return nil, nil, err
	}

	m.logger.DebugContext(ctx, "app meter: listing apps")
	appOpts := client.NewAppListOptions() // For fast troubleshooting, set .GUIDs to an app GUID.
	apps, spaces, err := m.client.AppsListWithSpaces(ctx, appOpts)
	if err != nil {
		return nil, nil, err
	}

	measurements := []reader.Measurement{}
	nodes := []*node.Node{}

	// Aggregate process usage info by app.
	m.logger.DebugContext(ctx, "app meter: aggregating process usage")
	appUsage := map[string]int{}
	for _, proc := range procs {
		usage := proc.Instances * proc.MemoryInMB
		appUsage[proc.Relationships.App.Data.GUID] += usage
	}

	m.logger.DebugContext(ctx, "app meter: aggregating app usage")
	for _, app := range apps {
		if app.State != appStateStarted {
			// Only STARTED apps consume resources. Skip the rest.
			continue
		}

		msrmt := reader.Measurement{
			Meter:             m.Name(),
			ResourceNaturalID: app.GUID,
			Value:             appUsage[app.GUID], // In MB. TODO: make sure units align.
		}

		spaceGUID := app.Relationships.Space.Data.GUID

		sidx := slices.IndexFunc(spaces, func(s *resource.Space) bool {
			return s.GUID == spaceGUID
		})
		if sidx < 0 {
			msrmt.Errs = errors.Join(msrmt.Errs, ErrSpaceNotFound)
			appNode, err := node.New(
				nil,
				app.GUID,
				node.WithSlugAuto("app", app.Name),
				node.WithPathAuto("orphan"),
			)
			if err != nil {
				return nil, nil, err
			}

			nodes = append(nodes, []*node.Node{appNode}...)
		} else {
			cfOrgGUIDString := spaces[sidx].Relationships.Organization.Data.GUID
			cfOrgGUID := dbx.UtilUUID(cfOrgGUIDString)

			org, err := m.dbq.GetCFOrg(ctx, cfOrgGUID)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, nil, err
			}

			customerID := org.CustomerID
			msrmt.CustomerID = customerID
			msrmt.OrgID = cfOrgGUIDString

			cfOrgNode, err := node.New(
				customerID,
				cfOrgGUIDString,
				node.WithSlugAuto("cforg", org.Name.String),
				node.WithPathAuto("apps.usage"),
			)
			if err != nil {
				return nil, nil, err
			}

			spaceNode, err := node.New(
				customerID,
				spaceGUID,
				node.WithSlugAuto("space", spaces[sidx].Name),
				node.WithPathByParent(cfOrgNode),
			)
			if err != nil {
				return nil, nil, err
			}

			appNode, err := node.New(
				customerID,
				app.GUID,
				node.WithSlugAuto("app", app.Name),
				node.WithPathByParent(spaceNode),
			)
			if err != nil {
				return nil, nil, err
			}

			nodes = append(nodes, []*node.Node{cfOrgNode, spaceNode, appNode}...)
		}

		measurements = append(measurements, msrmt)
	}

	return measurements, nodes, nil
}
