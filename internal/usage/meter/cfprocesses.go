package meter

import (
	"context"
	"errors"
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

type CFProcessMeter struct {
	name      string
	apps      CFAppClient
	processes CFProcessClient
}

func NewCFProcessMeter(apps CFAppClient, processes CFProcessClient) *CFProcessMeter {
	return &CFProcessMeter{
		name:      "cfapps",
		apps:      apps,
		processes: processes,
	}
}

func (m *CFProcessMeter) ReadUsage(ctx context.Context) ([]reader.Measurement, error) {
	procs, err := m.processes.ListAll(ctx, client.NewProcessOptions())
	if err != nil {
		return []reader.Measurement{}, err
	}
	apps, spaces, err := m.apps.ListIncludeSpacesAll(ctx, client.NewAppListOptions())
	if err != nil {
		return []reader.Measurement{}, err
	}

	var readings = make([]reader.Measurement, len(apps))

	// Aggregate process usage info by app.
	appUsage := make(map[string]int, len(apps))
	for _, proc := range procs {
		usage := proc.Instances * proc.MemoryInMB
		appUsage[proc.Relationships.App.Data.GUID] += usage
	}

	for i, app := range apps {
		if app.State != appStateStarted {
			// Only STARTED apps consume resources. Skip the rest.
			continue
		}
		m := reader.Measurement{
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

		readings[i] = m
	}

	return readings, nil
}
