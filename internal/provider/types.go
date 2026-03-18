package provider

type FunctionSpec struct {
	Name          string
	Runtime       string // "nodejs18" | "python3.11" | "go1.21"
	MemoryMB      int
	TimeoutSecs   int
	Environment   map[string]string
	ExecutionRole string // role ARN or equivalent
	CodeBucket    string
	CodeKey       string
}

type FunctionResult struct {
	ID           string            // Lambda ARN or GCP function URL
	Endpoint     string
	Version      string
	ProviderMeta map[string]string // escape hatch for provider-specific data
}

type APISpec struct {
	Name     string
	TargetID string // Lambda ARN or GCP function ID
	Protocol string // "HTTP" | "WEBSOCKET"
	Stage    string
}

type APIResult struct {
	ID       string
	Endpoint string // Public HTTPS URL
}

type TableSpec struct {
	Name string
}

type DatabaseSpec struct {
	Name   string
	Type   string // "key-value" | "relational"
	Tables []TableSpec
}

type DatabaseResult struct {
	ID           string
	Endpoint     string
	ProviderMeta map[string]string
}

type RoleSpec struct {
	Name       string
	PolicyDocs []string // JSON policy documents
}

type RoleResult struct {
	ID           string // Role ARN or equivalent
	ProviderMeta map[string]string
}
