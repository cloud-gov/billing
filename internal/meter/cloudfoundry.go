package meter

import (
	"context"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
)

// Abstract to an interface so we can create a mock client for testing.
type CFServiceInstanceClient interface {
	ListAll(context.Context, *client.ServiceInstanceListOptions) ([]*resource.ServiceInstance, error)
}

type CFSpaceClient interface {
	GetIncludeOrganization(context.Context, string) (*resource.Space, *resource.Organization, error)
}

// Reading is a single point-in-time snapshot of the utilization of a billable resource. Reading only includes information gleaned directly from the target system -- not the database.
type Reading struct {
	OrgID      string
	PlanID     string
	InstanceID string
	Value      int
	Time       time.Time
	Errs       []error
}

// ReadUsage function gets usage across all paid resources.
// First draft: Just do top-level services. TODO: Applications and sub-resources.
// Returns a non-nil error if there was an error during the overall process of reading usage information from the target system. If individual readings had errors, their errs fields should be set.
func ReadUsage(ctx context.Context, services CFServiceInstanceClient, spaces CFSpaceClient) ([]Reading, error) {
	si, err := services.ListAll(ctx, &client.ServiceInstanceListOptions{
		Type: "managed", // Ignore user-provided services, which we do not bill for. IMPORTANT: If this is not set, user-provided services will be included. Some response fields that we assume are non-nil, like .Relationships, will be nil on user-provided services. The code below does not guard against this and will panic.
	})
	if err != nil {
		return nil, err
	}
	usage := make([]Reading, len(si))
	now := time.Now().UTC()

	for i, instance := range si {
		_, org, err := spaces.GetIncludeOrganization(ctx, instance.Relationships.Space.Data.GUID)
		orgID := ""
		if err != nil {
			orgID = org.GUID
		}
		usage[i] = Reading{
			OrgID:      orgID,
			PlanID:     instance.Relationships.ServicePlan.Data.GUID,
			InstanceID: instance.GUID,
			Value:      1, // For this type of service, 1 indicates it is present at time of reading
			Time:       now,
			Errs:       []error{err},
		}
	}
	return usage, nil
}
