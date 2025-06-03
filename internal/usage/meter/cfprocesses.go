package meter

import (
	"context"
	"errors"
	"slices"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"

	"github.com/cloud-gov/billing/internal/usage/reader"
)

var (
	ErrAppNotFound = errors.New("processes meter: application not found")
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

	// Convert each Process to a Measurement.
	for i, proc := range procs {
		val := proc.MemoryInMB * proc.Instances
		appGUID := proc.Relationships.App.Data.GUID
		appIdx := slices.IndexFunc(apps, func(app *resource.App) bool {
			return app.GUID == appGUID
		})

		m := reader.Measurement{
			ResourceNaturalID: proc.GUID,
			Value:             val, // In MB. TODO: make sure units align.
			Errs:              []error{},
		}

		// Since procs and apps came from different requests, there is a chance their data will not match and appIdx = -1.
		if appIdx < 0 {
			m.Errs = append(m.Errs, ErrAppNotFound)
		} else {
			app := apps[appIdx]
			spaceGUID := app.Relationships.Space.Data.GUID
			sidx := slices.IndexFunc(spaces, func(s *resource.Space) bool {
				return s.GUID == spaceGUID
			})
			orgGUID := spaces[sidx].Relationships.Organization.Data.GUID

			m.OrgID = orgGUID
		}

		readings[i] = m
	}

	return readings, nil
}
