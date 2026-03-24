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

// Package factory is the only place in the codebase that imports concrete
// cloud provider implementations. Keeping the factory here prevents a
// circular dependency between internal/provider and internal/aws.
package factory

import (
	"fmt"

	awsprovider "github.com/platformer-io/platformer/internal/aws"
	"github.com/platformer-io/platformer/internal/provider"
	"github.com/platformer-io/platformer/internal/provider/mock"
)

// NewProvider constructs the CloudProvider selected by cfg.Cloud.
func NewProvider(cfg provider.Config) (provider.CloudProvider, error) {
	switch cfg.Cloud {
	case "aws":
		return awsprovider.NewAWSProvider(cfg)
	case "mock":
		return mock.NewMockProvider(), nil
	default:
		return nil, fmt.Errorf("factory: unknown cloud provider %q", cfg.Cloud)
	}
}
