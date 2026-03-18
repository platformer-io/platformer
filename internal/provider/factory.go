package provider

import "fmt"

type Config struct {
	Cloud  string // "aws" | "gcp" | "azure"
	Region string
	// credentials handled via env / service account
}

// NewProvider selects and returns the correct CloudProvider implementation.
// This is the only place in the codebase where cloud provider names are used as strings.
func NewProvider(config Config) (CloudProvider, error) {
	switch config.Cloud {
	case "aws":
		// imported and wired in cmd/operator/main.go to avoid circular deps
		return nil, fmt.Errorf("factory: aws provider must be wired via cmd/operator/main.go")
	default:
		return nil, fmt.Errorf("factory: unknown cloud provider %q", config.Cloud)
	}
}
