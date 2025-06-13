// Package reader reads usage information from systems that contain billable resources, like Cloud Foundry and AWS.
package reader

import (
	"context"
	"errors"
	"time"
)

// Reading is a point in time at which measurements of billable resources were taken.
type Reading struct {
	Time         time.Time
	Measurements []Measurement
}

// Measurement is a single point-in-time snapshot of the utilization of a billable resource. Measurement only includes information gleaned directly from the target system -- not the database.
type Measurement struct {
	OrgID string
	// Meter is the name of the meter which produced this [Measurement].
	Meter string
	// ResourceKindNaturalID is the "natural" ID of the Kind of billable resource being measured. The ID is maintained by the external system. For example, the plan ID of a Cloud Foundry service instance. Not all ResourceKinds have a natural ID, so this field may be nil.
	ResourceKindNaturalID *string
	// ResourceNaturalID is the "natural" ID of the billable Resource being measured. The ID is maintained by the external system. For example, the service instance GUID of a Cloud Foundry service instance, or the process ID of a Cloud Foundry process.
	ResourceNaturalID string
	Value             int
	// Errs contains any errors that occurred while gathering information about this particular measurement. It exists so we can preserve as much data as possible about the measurement. For instance, if we record a resource but fail to get its corresponding organization, a Measurement should be returned with a blank OrgID field and an Errs field including the error. Use [errors.Join] to add new errors.
	Errs error
}

// Meter defines a type that can read usage information from a system containing billable resources, akin to a utility meter.
type Meter interface {
	ReadUsage(context.Context) ([]Measurement, error)
	Name() string
}

// Reader reads usage information from all configured meters and returns it in aggregate.
type Reader struct {
	meters []Meter
}

// TODO: Meter registration needs to make sure meters all have entries in the database.
func New(meters []Meter) *Reader {
	return &Reader{
		meters: meters,
	}
}

func (m *Reader) Read(ctx context.Context) (Reading, error) {
	reading := Reading{
		Time:         time.Now().UTC(),
		Measurements: make([]Measurement, 0),
	}
	var reterr error

	for _, p := range m.meters {
		r, err := p.ReadUsage(ctx)
		if err != nil {
			reterr = errors.Join(reterr, err)
		}
		reading.Measurements = append(reading.Measurements, r...)
	}

	return reading, reterr
}

// next step: POST a ReadMeter job or something. Starts a job which finishes when services are read and result is stored in the database.
