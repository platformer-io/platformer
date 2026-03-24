// Copyright 2026 PlatFormer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
