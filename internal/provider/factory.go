package provider

// Config holds provider-agnostic configuration passed to NewProvider in
// the internal/factory package (which is the only place that imports
// concrete providers, keeping internal/provider free of circular deps).
type Config struct {
	Cloud  string // "aws" | "mock"
	Region string
	// credentials resolved from environment / IAM role
}
