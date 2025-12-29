package meter_test

import (
	"log/slog"
	"testing"

	"github.com/cloudfoundry/go-cfclient/v3/resource"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/usage/meter"
)

type mockClient struct {
	Spaces           *MockSpaceClient
	ServiceInstances *MockServiceInstanceClient
	ServicePlans     *MockServicePlanClient
	ServiceOfferings *MockServiceOfferingClient
}

func TestCFServiceMeter_ReadUsage(t *testing.T) {
	// arrange
	client := mockClient{
		Spaces:           NewMockSpaceClient(),
		ServiceInstances: NewMockServiceInstanceClient(),
		ServicePlans:     NewMockServicePlanClient(),
		ServiceOfferings: NewMockServiceOfferingClient(),
	}

	offeringID := newUUID()
	planID := newUUID()
	orgID := newUUID()
	spaceID := newUUID()
	instanceID := newUUID()

	client.ServiceOfferings.Data = append(client.ServiceOfferings.Data, &resource.ServiceOffering{
		Resource: resource.Resource{
			GUID: offeringID,
		},
	})

	client.ServicePlans.Data = append(client.ServicePlans.Data, &resource.ServicePlan{
		Resource: resource.Resource{
			GUID: planID,
		},
		Relationships: resource.ServicePlanRelationship{
			ServiceOffering: resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: offeringID,
				},
			},
		},
	})

	client.Spaces.Data = append(client.Spaces.Data, &resource.Space{
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

	client.ServiceInstances.Data = append(client.ServiceInstances.Data, &resource.ServiceInstance{
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

	customers := []db.Customer{{}}

	sut := meter.NewCFServiceMeter(slog.Default(), customers, client)

	// act
	readings, _, err := sut.ReadUsage(t.Context())
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
