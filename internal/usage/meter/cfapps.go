package meter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"

	"github.com/cloud-gov/billing/internal/usage/reader"
)

const appStateStarted = "STARTED"

var (
	ErrAppNotFound   = errors.New("CF processes meter: application not found")
	ErrSpaceNotFound = errors.New("CF processes meter: space not found")
)

type CFProcessClient interface {
	ListAll(context.Context, *client.ProcessListOptions) ([]*resource.Process, error)
}

type CFAppClient interface {
	ListIncludeSpacesAll(context.Context, *client.AppListOptions) ([]*resource.App, []*resource.Space, error)
}

type CFAppMeter struct {
	logger *slog.Logger

	apps      CFAppClient
	processes CFProcessClient
}

func NewCFAppMeter(logger *slog.Logger, apps CFAppClient, processes CFProcessClient) *CFAppMeter {
	return &CFAppMeter{
		logger:    logger.WithGroup("CFAppMeter"),
		apps:      apps,
		processes: processes,
	}
}

func (m *CFAppMeter) Name() string {
	return "cfapps"
}

func (m *CFAppMeter) ReadUsage(ctx context.Context) ([]reader.Measurement, []reader.Node, error) {
	m.logger.DebugContext(ctx, "app meter: listing processes")
	procs, err := m.processes.ListAll(ctx, client.NewProcessOptions())
	if err != nil {
		return nil, nil, err
	}
	m.logger.DebugContext(ctx, "app meter: listing apps")
	appopts := client.NewAppListOptions() // For fast troubleshooting, set .GUIDs to an app GUID.
	apps, spaces, err := m.apps.ListIncludeSpacesAll(ctx, appopts)
	if err != nil {
		return nil, nil, err
	}

	measurements := []reader.Measurement{}
	nodes := []reader.Node{}

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

		m := reader.Measurement{
			Meter:             m.Name(),
			ResourceNaturalID: app.GUID,
			Value:             appUsage[app.GUID], // In MB. TODO: make sure units align.
		}

		cfOrgGUID := ""
		spaceGUID := app.Relationships.Space.Data.GUID

		sidx := slices.IndexFunc(spaces, func(s *resource.Space) bool {
			return s.GUID == spaceGUID
		})
		if sidx < 0 {
			m.Errs = errors.Join(m.Errs, ErrSpaceNotFound)
		} else {
			cfOrgGUID = spaces[sidx].Relationships.Organization.Data.GUID
			m.OrgID = cfOrgGUID
		}

		measurements = append(measurements, m)

		// TODO: Get the customer ID here
		cfOrgNode := reader.Node{Slug: fmt.Sprintf("cforg_%v", cfOrgGUID), ResourceNaturalID: cfOrgGUID}
		spaceNode := reader.Node{Slug: fmt.Sprintf("space_%v", spaceGUID), ResourceNaturalID: spaceGUID}
		appNode := reader.Node{Slug: fmt.Sprintf("app_%v", app.GUID), ResourceNaturalID: app.GUID}

		cfOrgNode.Path = fmt.Sprintf("%v.%v", "xxx.apps.usage", cfOrgNode.Slug)
		spaceNode.Path = fmt.Sprintf("%v.%v", cfOrgNode.Path, spaceNode.Slug)
		appNode.Path = fmt.Sprintf("%v.%v", spaceNode.Path, appNode.Slug)

		nodes = append(nodes, []reader.Node{cfOrgNode, spaceNode, appNode}...)
	}

	return measurements, nodes, nil
}
