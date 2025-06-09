package meter

import (
	"context"
	"log/slog"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"

	"github.com/cloud-gov/billing/internal/usage/reader"
)

// Abstract to an interface so we can create a mock client for testing.
type CFServiceInstanceClient interface {
	ListAll(context.Context, *client.ServiceInstanceListOptions) ([]*resource.ServiceInstance, error)
}

type CFSpaceClient interface {
	ListAll(context.Context, *client.SpaceListOptions) ([]*resource.Space, error)
}

// CFServiceMeter reads usage from Cloud Foundry service instances.
type CFServiceMeter struct {
	logger *slog.Logger

	services CFServiceInstanceClient
	spaces   CFSpaceClient
}

func NewCFServiceMeter(logger *slog.Logger, services CFServiceInstanceClient, spaces CFSpaceClient) *CFServiceMeter {
	return &CFServiceMeter{
		logger:   logger.WithGroup("CFServiceMeter"),
		services: services,
		spaces:   spaces,
	}
}

// ReadUsage returns the point-in-time usage of services in Cloud Foundry.
// Returns a non-nil error if there was an error during the overall process of reading usage information from the target system. If individual readings had errors, their errs fields should be set.
func (m *CFServiceMeter) ReadUsage(ctx context.Context) ([]reader.Measurement, error) {
	m.logger.DebugContext(ctx, "service meter: listing services")
	opts := client.NewServiceInstanceListOptions()
	// Ignore user-provided services, which we do not bill for. IMPORTANT: If this is not set, user-provided services will be included. Some response fields that we assume are non-nil, like .Relationships, will be nil on user-provided services. The code below does not guard against this and will panic.
	opts.Type = "managed"
	si, err := m.services.ListAll(ctx, opts)
	if err != nil {
		return nil, err
	}

	m.logger.DebugContext(ctx, "service meter: listing spaces")
	spaceopts := client.NewSpaceListOptions()
	spaces, err := m.spaces.ListAll(ctx, spaceopts)
	if err != nil {
		return nil, err
	}
	spacesToOrgs := make(map[string]string, len(spaces))
	for _, space := range spaces {
		spacesToOrgs[space.GUID] = space.Relationships.Organization.Data.GUID
	}

	usage := make([]reader.Measurement, len(si))

	m.logger.DebugContext(ctx, "service meter: aggregating services")
	for i, instance := range si {
		usage[i] = reader.Measurement{
			OrgID:                 spacesToOrgs[instance.Relationships.Space.Data.GUID],
			ResourceKindNaturalID: instance.Relationships.ServicePlan.Data.GUID,
			ResourceNaturalID:     instance.GUID,
			Value:                 1, // For this type of service, 1 indicates it is present at time of reading
			Errs:                  nil,
		}
	}
	return usage, nil
}
