// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	CreateCFOrg(ctx context.Context, arg CreateCFOrgParams) (CFOrg, error)
	CreateCustomer(ctx context.Context, arg CreateCustomerParams) (Customer, error)
	CreateResource(ctx context.Context, arg CreateResourceParams) (Resource, error)
	CreateResourceKind(ctx context.Context, arg CreateResourceKindParams) (ResourceKind, error)
	CreateTier(ctx context.Context, arg CreateTierParams) (Tier, error)
	DeleteCFOrg(ctx context.Context, id uuid.UUID) error
	DeleteCustomer(ctx context.Context, id int64) error
	DeleteResource(ctx context.Context, id int32) error
	DeleteResourceKind(ctx context.Context, id int32) error
	DeleteTier(ctx context.Context, id int32) error
	GetCFOrg(ctx context.Context, id uuid.UUID) (CFOrg, error)
	GetCustomer(ctx context.Context, id int64) (Customer, error)
	GetResource(ctx context.Context, id int32) (Resource, error)
	GetResourceKind(ctx context.Context, id int32) (ResourceKind, error)
	GetTier(ctx context.Context, id int32) (Tier, error)
	ListCFOrgs(ctx context.Context) ([]CFOrg, error)
	ListCustomers(ctx context.Context) ([]Customer, error)
	ListResourceKind(ctx context.Context) ([]ResourceKind, error)
	ListResources(ctx context.Context) ([]Resource, error)
	ListTiers(ctx context.Context) ([]Tier, error)
	UpdateCFOrg(ctx context.Context, arg UpdateCFOrgParams) error
	UpdateCustomer(ctx context.Context, arg UpdateCustomerParams) error
	UpdateResource(ctx context.Context, arg UpdateResourceParams) error
	UpdateResourceKind(ctx context.Context, arg UpdateResourceKindParams) error
	UpdateTier(ctx context.Context, arg UpdateTierParams) error
}

var _ Querier = (*Queries)(nil)
