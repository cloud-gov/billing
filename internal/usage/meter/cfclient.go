package meter

import (
	"context"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
)

type AppClient interface {
	Apps
	Processes
}

type ServiceClient interface {
	Spaces
	ServiceInstances
	ServicePlans
}

type Spaces interface {
	SpacesList(context.Context, *client.SpaceListOptions) ([]*resource.Space, error)
}
type ServiceInstances interface {
	ServiceInstancesList(context.Context, *client.ServiceInstanceListOptions) ([]*resource.ServiceInstance, error)
}
type ServicePlans interface {
	ServicePlansOfferingsList(context.Context, *client.ServicePlanListOptions) ([]*resource.ServicePlan, []*resource.ServiceOffering, error)
}
type Processes interface {
	ProcessesList(context.Context, *client.ProcessListOptions) ([]*resource.Process, error)
}
type Apps interface {
	AppsListWithSpaces(context.Context, *client.AppListOptions) ([]*resource.App, []*resource.Space, error)
}

type CFAdapter struct {
	*client.Client
}

func (c *CFAdapter) AppsListWithSpaces(ctx context.Context, opts *client.AppListOptions) ([]*resource.App, []*resource.Space, error) {
	return c.Applications.ListIncludeSpacesAll(ctx, opts)
}

func (c *CFAdapter) ProcessesList(ctx context.Context, opts *client.ProcessListOptions) ([]*resource.Process, error) {
	return c.Processes.ListAll(ctx, opts)
}

func (c *CFAdapter) SpacesList(ctx context.Context, opts *client.SpaceListOptions) ([]*resource.Space, error) {
	return c.Spaces.ListAll(ctx, opts)
}

func (c *CFAdapter) ServiceInstancesList(ctx context.Context, opts *client.ServiceInstanceListOptions) ([]*resource.ServiceInstance, error) {
	return c.ServiceInstances.ListAll(ctx, opts)
}

func (c *CFAdapter) ServicePlansOfferingsList(ctx context.Context, opts *client.ServicePlanListOptions) ([]*resource.ServicePlan, []*resource.ServiceOffering, error) {
	return c.ServicePlans.ListIncludeServiceOfferingAll(ctx, opts)
}
