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
