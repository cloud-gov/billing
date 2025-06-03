package meter

import (
	"context"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"

	"github.com/cloud-gov/billing/internal/usage/reader"
)

// Abstract to an interface so we can create a mock client for testing.
type CFServiceInstanceClient interface {
	ListAll(context.Context, *client.ServiceInstanceListOptions) ([]*resource.ServiceInstance, error)
}

type CFSpaceClient interface {
	GetIncludeOrganization(context.Context, string) (*resource.Space, *resource.Organization, error)
}

// CFServiceMeter reads usage from Cloud Foundry service instances.
type CFServiceMeter struct {
	services CFServiceInstanceClient
	spaces   CFSpaceClient
}

func NewCFServiceMeter(services CFServiceInstanceClient, spaces CFSpaceClient) *CFServiceMeter {
	return &CFServiceMeter{
		services: services,
		spaces:   spaces,
	}
}

// ReadUsage returns the point-in-time usage of services in Cloud Foundry.
// Returns a non-nil error if there was an error during the overall process of reading usage information from the target system. If individual readings had errors, their errs fields should be set.
func (m *CFServiceMeter) ReadUsage(ctx context.Context) ([]reader.Measurement, error) {
	si, err := m.services.ListAll(ctx, &client.ServiceInstanceListOptions{
		Type: "managed", // Ignore user-provided services, which we do not bill for. IMPORTANT: If this is not set, user-provided services will be included. Some response fields that we assume are non-nil, like .Relationships, will be nil on user-provided services. The code below does not guard against this and will panic.
	})
	if err != nil {
		return nil, err
	}
	usage := make([]reader.Measurement, len(si))

	for i, instance := range si {
		_, org, errr := m.spaces.GetIncludeOrganization(ctx, instance.Relationships.Space.Data.GUID)
		orgID := ""
		if errr != nil {
			orgID = org.GUID
		}
		usage[i] = reader.Measurement{
			OrgID:      orgID,
			PlanID:     instance.Relationships.ServicePlan.Data.GUID,
			InstanceID: instance.GUID,
			Value:      1, // For this type of service, 1 indicates it is present at time of reading
			Errs:       []error{errr},
		}
	}
	return usage, nil
}
