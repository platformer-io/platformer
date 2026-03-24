package provider

import (
	"context"
)

type CloudProvider interface {

	// Compute
	CreateFunction(ctx context.Context, spec FunctionSpec) (*FunctionResult, error)
	UpdateFunction(ctx context.Context, spec FunctionSpec) (*FunctionResult, error)
	DeleteFunction(ctx context.Context, name string) error
	GetFunction(ctx context.Context, name string) (*FunctionResult, error)

	// Networking
	CreateAPIEndpoint(ctx context.Context, spec APISpec) (*APIResult, error)
	UpdateAPIEndpoint(ctx context.Context, spec APISpec) (*APIResult, error)
	DeleteAPIEndpoint(ctx context.Context, id string, functionName string) error
	GetAPIEndpoint(ctx context.Context, id string) (*APIResult, error)

	// Database
	CreateDatabase(ctx context.Context, spec DatabaseSpec) (*DatabaseResult, error)
	DeleteDatabase(ctx context.Context, name string) error
	GetDatabase(ctx context.Context, name string) (*DatabaseResult, error)

	// IAM / Identity
	CreateExecutionRole(ctx context.Context, spec RoleSpec) (*RoleResult, error)
	DeleteExecutionRole(ctx context.Context, name string) error
	GetExecutionRole(ctx context.Context, name string) (*RoleResult, error)

	// Observability
	CreateLogGroup(ctx context.Context, name string) error
	DeleteLogGroup(ctx context.Context, name string) error
}
