package meter_test

import (
	"context"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cloud-gov/billing/internal/db"
	_ "github.com/cloud-gov/billing/internal/usage/meter" // Imported so doc comment references work.
)

func newUUID() string {
	return uuid.NewString()
}

// TODO: it would be better to use a test DB than these stubs
type StubDbQ struct {
	Org      db.CFOrg
	OrgError error
}

func (d *StubDbQ) GetCFOrg(ctx context.Context, id pgtype.UUID) (o db.CFOrg, e error) {
	o = d.Org
	e = d.OrgError

	return o, e
}

// MockAppMeterCfProvider is an in-memory implementation of [meter.AppMeterCfProvider].
type MockAppMeterCfProvider struct {
	Apps      []*resource.App
	Spaces    []*resource.Space
	Processes []*resource.Process
	Orgs      []*resource.Organization
	AppErr    error
	ProcErr   error
}

// MockServiceMeterCfProvider is an in-memory implementation of [meter.ServiceMeterCfProvider].
type MockServiceMeterCfProvider struct {
	Spaces    []*resource.Space
	Instances []*resource.ServiceInstance
	Plans     []*resource.ServicePlan
	Offerings []*resource.ServiceOffering
	Orgs      []*resource.Organization
}

func NewMockAppMeterCfProvider() *MockAppMeterCfProvider {
	return &MockAppMeterCfProvider{}
}

func NewMockServiceMeterCfProvider() *MockServiceMeterCfProvider {
	return &MockServiceMeterCfProvider{}
}

func (p *MockAppMeterCfProvider) AppsListWithSpaces(_ context.Context, _ *client.AppListOptions) ([]*resource.App, []*resource.Space, error) {
	return p.Apps, p.Spaces, p.AppErr
}

func (p *MockAppMeterCfProvider) AppsListWithSpacesAndOrgs(_ context.Context, _ *client.AppListOptions) ([]*resource.App, []*resource.Space, []*resource.Organization, error) {
	return p.Apps, p.Spaces, p.Orgs, p.AppErr
}

func (p *MockAppMeterCfProvider) ProcessesList(_ context.Context, _ *client.ProcessListOptions) ([]*resource.Process, error) {
	return p.Processes, p.ProcErr
}

func (p *MockServiceMeterCfProvider) SpacesList(_ context.Context, _ *client.SpaceListOptions) ([]*resource.Space, error) {
	return p.Spaces, nil
}

func (p *MockServiceMeterCfProvider) SpacesListWithOrgs(_ context.Context, _ *client.SpaceListOptions) ([]*resource.Space, []*resource.Organization, error) {
	return p.Spaces, p.Orgs, nil
}

func (p *MockServiceMeterCfProvider) ServiceInstancesList(_ context.Context, _ *client.ServiceInstanceListOptions) ([]*resource.ServiceInstance, error) {
	return p.Instances, nil
}

func (p *MockServiceMeterCfProvider) ServicePlansOfferingsList(_ context.Context, _ *client.ServicePlanListOptions) ([]*resource.ServicePlan, []*resource.ServiceOffering, error) {
	return p.Plans, p.Offerings, nil
}
