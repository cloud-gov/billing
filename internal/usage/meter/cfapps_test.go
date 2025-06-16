package meter_test

import (
	"errors"
	"log/slog"
	"reflect"
	"testing"

	"github.com/cloudfoundry/go-cfclient/v3/resource"

	"github.com/cloud-gov/billing/internal/usage/meter"
	"github.com/cloud-gov/billing/internal/usage/reader"
)

const (
	appStateStarted = "STARTED"
	appStateStopped = "STOPPED"
)

func mkProc(appGUID string, instances, mb int) *resource.Process {
	return &resource.Process{
		Relationships: resource.ProcessRelationships{
			App: resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: appGUID,
				},
			},
		},
		Instances:  instances,
		MemoryInMB: mb,
	}
}

func mkApp(guid, spaceGUID, state string) *resource.App {
	return &resource.App{
		Resource: resource.Resource{
			GUID: guid,
		},
		Relationships: resource.AppRelationships{
			Space: resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: spaceGUID,
				},
			},
		},
		State: state,
	}
}

func mkSpace(guid, orgGUID string) *resource.Space {
	return &resource.Space{
		Resource: resource.Resource{
			GUID: guid,
		},
		Relationships: &resource.SpaceRelationships{
			Organization: &resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: orgGUID,
				},
			},
		},
	}
}

// measurementsToMap converts the slice the function returns into a
// map[appGUID]usage, ignoring zero‑value entries.
func measurementsToMap(ms []reader.Measurement) map[string]int {
	out := map[string]int{}
	for _, m := range ms {
		if m.ResourceNaturalID != "" {
			out[m.ResourceNaturalID] = m.Value
		}
	}
	return out
}

func measurementErrsToMap(ms []reader.Measurement) map[string]error {
	out := map[string]error{}
	for _, m := range ms {
		if m.ResourceNaturalID != "" {
			out[m.ResourceNaturalID] = m.Errs
		}
	}
	return out
}

func TestCFAppMeter_ReadUsage(t *testing.T) {
	const (
		app1 = "app‑1"
		app2 = "app‑2"
		sp   = "space‑1"
		org  = "org‑1"
	)

	hugeInstances := 1024
	hugeMemory := 128 * 1024 // 128GB
	hugeValue := 134_217_728 // instances * memory

	tests := []struct {
		name               string
		procs              []*resource.Process
		apps               []*resource.App
		spaces             []*resource.Space
		procErr            error
		appErr             error
		want               map[string]int // expected aggregated usage by app GUID
		wantMeasurementErr map[string]error
		wantErr            bool
	}{
		{
			name:    "process call error",
			procErr: errors.New("CF process API error"),
			wantErr: true,
		},
		{
			name:    "app call error",
			appErr:  errors.New("CF app API error"),
			wantErr: true,
		},
		{
			name:   "no apps",
			apps:   nil,
			spaces: nil,
			want:   map[string]int{},
		},
		{
			name:   "one app, no processes",
			apps:   []*resource.App{mkApp(app1, sp, appStateStarted)},
			spaces: []*resource.Space{mkSpace(sp, org)},
			want:   map[string]int{app1: 0},
		},
		{
			name: "aggregate multiple procs",
			procs: []*resource.Process{
				mkProc(app1, 2, 512), // 1024
				mkProc(app1, 1, 256), // +256 = 1280
			},
			apps:   []*resource.App{mkApp(app1, sp, appStateStarted)},
			spaces: []*resource.Space{mkSpace(sp, org)},
			want:   map[string]int{app1: 1280},
		},
		{
			name: "process for unknown app is ignored",
			procs: []*resource.Process{
				mkProc("orphan‑app", 1, 512),
			},
			apps:   []*resource.App{mkApp(app1, sp, appStateStarted)},
			spaces: []*resource.Space{mkSpace(sp, org)},
			want:   map[string]int{app1: 0},
		},
		{
			name: "stopped app is skipped",
			procs: []*resource.Process{
				mkProc(app1, 1, 128),
				mkProc(app2, 2, 128),
			},
			apps: []*resource.App{
				mkApp(app1, sp, appStateStarted),
				mkApp(app2, sp, appStateStopped), // skipped
			},
			spaces: []*resource.Space{mkSpace(sp, org)},
			want:   map[string]int{app1: 128},
		},
		{
			name: "missing space error is collected",
			procs: []*resource.Process{
				mkProc(app1, 1, 128),
			},
			apps: []*resource.App{
				mkApp(app1, "non‑existent‑space", appStateStarted),
			},
			spaces:             []*resource.Space{mkSpace(sp, org)},
			want:               map[string]int{app1: 128},
			wantMeasurementErr: map[string]error{app1: meter.ErrSpaceNotFound},
		},
		{
			name: "space present but org missing",
			procs: []*resource.Process{
				mkProc(app1, 1, 128),
			},
			apps:   []*resource.App{mkApp(app1, sp, appStateStarted)},
			spaces: []*resource.Space{mkSpace(sp, "")}, // empty org GUID
			want:   map[string]int{app1: 128},
		},
		{
			name: "large numbers",
			procs: []*resource.Process{
				mkProc(app1, hugeInstances, hugeMemory),
			},
			apps:   []*resource.App{mkApp(app1, sp, appStateStarted)},
			spaces: []*resource.Space{mkSpace(sp, org)},
			want:   map[string]int{app1: hugeValue},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()

			sut := meter.NewCFAppMeter(slog.Default(), &MockAppClient{Apps: tc.apps, Spaces: tc.spaces, AppErr: tc.appErr}, &MockProcessClient{Processes: tc.procs, Err: tc.procErr})

			got, err := sut.ReadUsage(t.Context())
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Skip further checks on error / panic cases.
			if tc.wantErr {
				return
			}

			gotMap := measurementsToMap(got)
			if !reflect.DeepEqual(tc.want, gotMap) {
				t.Fatalf("want %v, got %v", tc.want, gotMap)
			}

			gotErrsMap := measurementErrsToMap(got)
			for k, v := range tc.wantMeasurementErr {
				if !errors.Is(gotErrsMap[k], v) {
					t.Fatalf("want %v, got %v", tc.wantMeasurementErr, gotErrsMap)
				}
			}
		})
	}
}
