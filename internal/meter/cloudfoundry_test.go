package meter_test

import (
	"testing"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/google/uuid"

	"github.com/cloud-gov/billing/internal/cf"
	"github.com/cloud-gov/billing/internal/meter"
)

func TestReadUsage(t *testing.T) {
	// arrange
	services := cf.NewMockServiceInstanceClient()
	spaces := cf.NewMockSpaceClient()

	instanceID := uuid.New().String()
	planID := uuid.New().String()
	spaceID := uuid.New().String()

	services.ServiceInstances = append(services.ServiceInstances, &resource.ServiceInstance{
		Resource: resource.Resource{
			GUID: instanceID,
		},
		Relationships: resource.ServiceInstanceRelationships{
			ServicePlan: &resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: planID,
				},
			},
			Space: &resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: spaceID,
				},
			},
		},
	})
	spaces.OrgsForSpaces[spaceID] = &resource.Organization{
		Resource: resource.Resource{
			GUID: uuid.New().String(),
		},
	}
	spaces.Spaces = append(spaces.Spaces, &resource.Space{
		Resource: resource.Resource{
			GUID: spaceID,
		},
	})

	// act
	readings, err := meter.ReadUsage(t.Context(), &services, &spaces)

	// assert
	if err != nil {
		t.Fatal("error was not expected when reading usage", err)
	}
	r := readings[0]
	if r.InstanceID != instanceID {
		t.Fatal("instance ID did not match")
	}
	if r.PlanID != planID {
		t.Fatal("plan ID did not match")
	}
}
