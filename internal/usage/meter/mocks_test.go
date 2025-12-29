package meter_test

import (
	"context"

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

// MockSpaceClient is an in-memory mock implementation of [meter.CFSpaceClient].
type MockSpaceClient struct {
	Data []*resource.Space
}

func NewMockSpaceClient() *MockSpaceClient {
	return &MockSpaceClient{
		Data: make([]*resource.Space, 0),
	}
}

func (c *MockSpaceClient) ListAll(_ context.Context, _ *client.SpaceListOptions) ([]*resource.Space, error) {
	return c.Data, nil
}

// MockServiceInstanceClient is an in-memory implementation of [meter.CFServiceInstanceClient].
type MockServiceInstanceClient struct {
	// need to be able to populate this with service instances for testing
	Data []*resource.ServiceInstance
}

func NewMockServiceInstanceClient() *MockServiceInstanceClient {
	return &MockServiceInstanceClient{
		Data: make([]*resource.ServiceInstance, 0),
	}
}

func (c *MockServiceInstanceClient) ListAll(_ context.Context, _ *client.ServiceInstanceListOptions) ([]*resource.ServiceInstance, error) {
	return c.Data, nil
}

// MockServicePlanClient is an in-memory implementation of [meter.CFServicePlanClient].
type MockServicePlanClient struct {
	// need to be able to populate this with service instances for testing
	Data []*resource.ServicePlan
}

func NewMockServicePlanClient() *MockServicePlanClient {
	return &MockServicePlanClient{
		Data: make([]*resource.ServicePlan, 0),
	}
}

func (c *MockServicePlanClient) ListAll(_ context.Context, _ *client.ServicePlanListOptions) ([]*resource.ServicePlan, error) {
	return c.Data, nil
}

// MockServiceOfferingClient is an in-memory implementation of [meter.CFServiceOfferingClient].
type MockServiceOfferingClient struct {
	// need to be able to populate this with service instances for testing
	Data []*resource.ServiceOffering
}

func NewMockServiceOfferingClient() *MockServiceOfferingClient {
	return &MockServiceOfferingClient{
		Data: make([]*resource.ServiceOffering, 0),
	}
}

func (c *MockServiceOfferingClient) ListAll(_ context.Context, _ *client.ServiceOfferingListOptions) ([]*resource.ServiceOffering, error) {
	return c.Data, nil
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
