package meter_test

import (
	"log/slog"
	"testing"

	"github.com/cloudfoundry/go-cfclient/v3/resource"

	"github.com/cloud-gov/billing/internal/usage/meter"
)

func TestCFServiceMeter_ReadUsage(t *testing.T) {
	// arrange
	cfProvider := NewMockServiceMeterCfProvider()

	offeringID := newUUID()
	planID := newUUID()
	orgID := newUUID()
	spaceID := newUUID()
	instanceID := newUUID()

	cfProvider.Offerings = append(cfProvider.Offerings, &resource.ServiceOffering{
		Resource: resource.Resource{
			GUID: offeringID,
		},
	})

	cfProvider.Plans = append(cfProvider.Plans, &resource.ServicePlan{
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

	cfProvider.Spaces = append(cfProvider.Spaces, &resource.Space{
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

	cfProvider.Instances = append(cfProvider.Instances, &resource.ServiceInstance{
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

	sut := meter.NewCFServiceMeter(slog.Default(), cfProvider, &StubDbQ{})

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
