package meter_test

import (
	"testing"

	"github.com/cloudfoundry/go-cfclient/v3/resource"

	"github.com/cloud-gov/billing/internal/usage/meter"
)

func Test_CFApps_ReadUsage(t *testing.T) {
	// arrange
	apps := NewMockAppClient()
	procs := NewMockProcessClient()

	procID := newUUID()
	appID := newUUID()
	spaceID := newUUID()
	orgID := newUUID()

	procs.Processes = append(procs.Processes, &resource.Process{
		Resource: resource.Resource{
			GUID: procID,
		},
		Relationships: resource.ProcessRelationships{
			App: resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: appID,
				},
			},
		},
		Instances:  2,
		MemoryInMB: 1024,
	})
	apps.Apps = append(apps.Apps, &resource.App{
		Resource: resource.Resource{
			GUID: appID,
		},
		Relationships: resource.AppRelationships{
			Space: resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: spaceID,
				},
			},
		},
	})
	apps.Spaces = append(apps.Spaces, &resource.Space{
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

	sut := meter.NewCFProcessMeter(apps, procs)

	// act
	measurements, err := sut.ReadUsage(t.Context())

	// assert
	if err != nil {
		t.Fatal("error was not expected when reading usage", err)
	}

	m := measurements[0]
	if m.ResourceNaturalID != procID {
		t.Fatalf("expected process ID %v, got %v", procID, m.ResourceNaturalID)
	}
	if m.ResourceKindNaturalID != "" {
		t.Fatalf("expected plan ID %v, got %v", "", m.ResourceKindNaturalID)
	}
	if m.OrgID != orgID {
		t.Fatalf("expected org ID %v, got %v", orgID, m.OrgID)
	}
	if m.Value != 2048 {
		t.Fatalf("expected reading value %v, got %v", 2048, m.Value)
	}
}
