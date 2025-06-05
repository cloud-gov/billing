package meter_test

import (
	"log/slog"
	"testing"

	"github.com/cloudfoundry/go-cfclient/v3/resource"

	"github.com/cloud-gov/billing/internal/usage/meter"
)

func TestCFServiceMeter_ReadUsage(t *testing.T) {
	// arrange
	services := NewMockServiceInstanceClient()
	spaces := NewMockSpaceClient()

	instanceID := newUUID()
	planID := newUUID()
	spaceID := newUUID()
	orgID := newUUID()

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
	spaces.Spaces = append(spaces.Spaces, &resource.Space{
		Resource: resource.Resource{
			GUID: spaceID,
		},
		Relationships: &resource.SpaceRelationships{
			Organization: &resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: orgID,
				},
			},
		},
	})

	sut := meter.NewCFServiceMeter(slog.Default(), services, spaces)

	// act
	readings, err := sut.ReadUsage(t.Context())

	// assert
	if err != nil {
		t.Fatal("error was not expected when reading usage", err)
	}
	r := readings[0]
	if r.ResourceNaturalID != instanceID {
		t.Fatal("instance ID did not match")
	}
	if r.ResourceKindNaturalID != planID {
		t.Fatal("plan ID did not match")
	}
}
