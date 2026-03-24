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
	APIID    string // existing API Gateway ID — only set on update, empty on create
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
	Name        string
	DatabaseARN string // optional — if non-empty, adds a scoped DynamoDB policy for this table ARN
}

type RoleResult struct {
	ID           string // Role ARN or equivalent
	ProviderMeta map[string]string
}
