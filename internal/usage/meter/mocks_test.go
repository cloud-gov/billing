package meter_test

import (
	"context"
	"slices"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/google/uuid"

	_ "github.com/cloud-gov/billing/internal/usage/meter" // Imported so doc comment references work.
)

func newUUID() string {
	return uuid.NewString()
}

// MockProcessClient is an in-memory implementation of [meter.CFAppClient].
type MockAppClient struct {
	Apps   []*resource.App
	Spaces []*resource.Space
	AppErr error
}

func NewMockAppClient() *MockAppClient {
	return &MockAppClient{
		Apps:   []*resource.App{},
		Spaces: []*resource.Space{},
	}
}

func (c *MockAppClient) ListIncludeSpacesAll(_ context.Context, _ *client.AppListOptions) ([]*resource.App, []*resource.Space, error) {
	return c.Apps, c.Spaces, c.AppErr
}

// MockServiceInstanceClient is an in-memory implementation of [meter.CFServiceInstanceClient].
type MockServiceInstanceClient struct {
	// need to be able to populate this with service instances for testing
	ServiceInstances []*resource.ServiceInstance
}

func NewMockServiceInstanceClient() *MockServiceInstanceClient {
	return &MockServiceInstanceClient{
		ServiceInstances: make([]*resource.ServiceInstance, 0),
	}
}

func (c *MockServiceInstanceClient) ListAll(_ context.Context, _ *client.ServiceInstanceListOptions) ([]*resource.ServiceInstance, error) {
	return c.ServiceInstances, nil
}

// MockSpaceClient is an in-memory mock implementation of [meter.CFSpaceClient].
type MockSpaceClient struct {
	Spaces []*resource.Space
	// OrgsForSpaces maps from Space GUIDs to organizations.
	OrgsForSpaces map[string]*resource.Organization
}

func NewMockSpaceClient() *MockSpaceClient {
	return &MockSpaceClient{
		Spaces:        make([]*resource.Space, 0),
		OrgsForSpaces: make(map[string]*resource.Organization),
	}
}

func (c *MockSpaceClient) GetIncludeOrganization(_ context.Context, guid string) (*resource.Space, *resource.Organization, error) {
	spaceIdx := slices.IndexFunc(c.Spaces, func(s *resource.Space) bool {
		return s.GUID == guid
	})
	if spaceIdx < 0 {
		return nil, nil, client.ErrNoResultsReturned
	}
	space := c.Spaces[spaceIdx]
	org, ok := c.OrgsForSpaces[space.GUID]
	if !ok {
		return nil, nil, client.ErrNoResultsReturned
	}

	return space, org, nil
}

// MockProcessClient is an in-memory implementation of [meter.CFProcessClient].
type MockProcessClient struct {
	Processes []*resource.Process
	Err       error
}

func NewMockProcessClient() *MockProcessClient {
	return &MockProcessClient{
		Processes: []*resource.Process{},
	}
}

func (c *MockProcessClient) ListAll(_ context.Context, _ *client.ProcessListOptions) ([]*resource.Process, error) {
	return c.Processes, c.Err
}
