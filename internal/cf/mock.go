package cf

import (
	"context"
	"slices"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
)

// MockServiceInstanceClient is an in-memory implementation of [client.ServiceInstanceClient].
type MockServiceInstanceClient struct {
	// need to be able to populate this with service instances for testing
	ServiceInstances []*resource.ServiceInstance
}

func NewMockServiceInstanceClient() MockServiceInstanceClient {
	return MockServiceInstanceClient{
		ServiceInstances: make([]*resource.ServiceInstance, 0),
	}
}

func (c *MockServiceInstanceClient) ListAll(ctx context.Context, opts *client.ServiceInstanceListOptions) ([]*resource.ServiceInstance, error) {
	return c.ServiceInstances, nil
}

// MockSpaceClient is an in-memory mock implementation of [client.SpaceClient].
type MockSpaceClient struct {
	Spaces []*resource.Space
	// OrgsForSpaces maps from Space GUIDs to organizations.
	OrgsForSpaces map[string]*resource.Organization
}

func NewMockSpaceClient() MockSpaceClient {
	return MockSpaceClient{
		Spaces:        make([]*resource.Space, 0),
		OrgsForSpaces: make(map[string]*resource.Organization),
	}
}

func (c *MockSpaceClient) GetIncludeOrganization(ctx context.Context, guid string) (*resource.Space, *resource.Organization, error) {
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
