// Package reader reads usage information from systems that contain billable resources, like Cloud Foundry and AWS.
package reader

import (
	"context"
	"errors"
	"time"
)

// Reading is a single point-in-time snapshot of the utilization of a billable resource. Reading only includes information gleaned directly from the target system -- not the database.
type Reading struct {
	OrgID      string
	PlanID     string // BillableClassID
	InstanceID string // ResourceID
	Value      int
	Time       time.Time
	Errs       []error
}

// Meter defines a type that can read usage information from a system containing billable resources, akin to a utility meter. It is defined here instead of the meter package to avoid a circular dependency, as that package must reference the [Reading] type.
type Meter interface {
	ReadUsage(context.Context) ([]Reading, error)
}

// Reader reads usage information from all configured meters and returns it in aggregate.
type Reader struct {
	meters []Meter
}

func New(meters []Meter) *Reader {
	return &Reader{
		meters: meters,
	}
}

func (m *Reader) Read(ctx context.Context) ([]Reading, error) {
	var readings = make([]Reading, 0)
	var reterr error
	for _, p := range m.meters {
		r, err := p.ReadUsage(ctx)
		if err != nil {
			reterr = errors.Join(reterr, err)
		}
		readings = append(readings, r...)
	}
	return readings, reterr
}

// next step: POST a ReadMeter job or something. Starts a job which finishes when services are read and result is stored in the database.
