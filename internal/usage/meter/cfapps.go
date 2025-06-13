package meter

import (
	"context"
	"errors"
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

func (m *CFAppMeter) ReadUsage(ctx context.Context) ([]reader.Measurement, error) {
	m.logger.DebugContext(ctx, "app meter: listing processes")
	procs, err := m.processes.ListAll(ctx, client.NewProcessOptions())
	if err != nil {
		return []reader.Measurement{}, err
	}
	m.logger.DebugContext(ctx, "app meter: listing apps")
	apps, spaces, err := m.apps.ListIncludeSpacesAll(ctx, client.NewAppListOptions())
	if err != nil {
		return []reader.Measurement{}, err
	}

	var measurements = make([]reader.Measurement, len(apps))

	// Aggregate process usage info by app.
	m.logger.DebugContext(ctx, "app meter: aggregating process usage")
	appUsage := make(map[string]int, len(apps))
	for _, proc := range procs {
		usage := proc.Instances * proc.MemoryInMB
		appUsage[proc.Relationships.App.Data.GUID] += usage
	}

	m.logger.DebugContext(ctx, "app meter: aggregating app usage")
	for i, app := range apps {
		if app.State != appStateStarted {
			// Only STARTED apps consume resources. Skip the rest.
			continue
		}
		m := reader.Measurement{
			Meter:             m.Name(),
			ResourceNaturalID: app.GUID,
			Value:             appUsage[app.GUID], // In MB. TODO: make sure units align.
		}
		spaceGUID := app.Relationships.Space.Data.GUID
		sidx := slices.IndexFunc(spaces, func(s *resource.Space) bool {
			return s.GUID == spaceGUID
		})
		if sidx < 0 {
			m.Errs = errors.Join(m.Errs, ErrSpaceNotFound)
		} else {
			orgGUID := spaces[sidx].Relationships.Organization.Data.GUID
			m.OrgID = orgGUID
		}

		measurements[i] = m
	}

	return measurements, nil
}
