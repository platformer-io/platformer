package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/platformer-io/platformer/internal/provider"
)

// MockProvider is an in-memory implementation of CloudProvider for use in tests.
// All state is stored in maps and can be inspected after calls.
type MockProvider struct {
	mu        sync.Mutex
	functions map[string]*provider.FunctionResult
	endpoints map[string]*provider.APIResult
	databases map[string]*provider.DatabaseResult
	roles     map[string]*provider.RoleResult
	logGroups map[string]bool

	// Call counters for assertion in tests
	Calls map[string]int
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		functions: make(map[string]*provider.FunctionResult),
		endpoints: make(map[string]*provider.APIResult),
		databases: make(map[string]*provider.DatabaseResult),
		roles:     make(map[string]*provider.RoleResult),
		logGroups: make(map[string]bool),
		Calls:     make(map[string]int),
	}
}

// --- Compute ---

func (m *MockProvider) CreateFunction(_ context.Context, spec provider.FunctionSpec) (*provider.FunctionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["CreateFunction"]++
	result := &provider.FunctionResult{
		ID:      fmt.Sprintf("mock-function-arn::%s", spec.Name),
		Version: "1",
	}
	m.functions[spec.Name] = result
	return result, nil
}

func (m *MockProvider) UpdateFunction(_ context.Context, spec provider.FunctionSpec) (*provider.FunctionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["UpdateFunction"]++
	if _, ok := m.functions[spec.Name]; !ok {
		return nil, fmt.Errorf("mock: function %q not found", spec.Name)
	}
	result := &provider.FunctionResult{
		ID:      fmt.Sprintf("mock-function-arn::%s", spec.Name),
		Version: "2",
	}
	m.functions[spec.Name] = result
	return result, nil
}

func (m *MockProvider) DeleteFunction(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["DeleteFunction"]++
	if _, ok := m.functions[name]; !ok {
		return fmt.Errorf("mock: function %q not found", name)
	}
	delete(m.functions, name)
	return nil
}

func (m *MockProvider) GetFunction(_ context.Context, name string) (*provider.FunctionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["GetFunction"]++
	result, ok := m.functions[name]
	if !ok {
		return nil, fmt.Errorf("mock: function %q not found", name)
	}
	return result, nil
}

// --- Networking ---

func (m *MockProvider) CreateAPIEndpoint(_ context.Context, spec provider.APISpec) (*provider.APIResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["CreateAPIEndpoint"]++
	result := &provider.APIResult{
		ID:       fmt.Sprintf("mock-api-id::%s", spec.Name),
		Endpoint: fmt.Sprintf("https://mock.execute-api.local/%s", spec.Stage),
	}
	m.endpoints[result.ID] = result
	return result, nil
}

func (m *MockProvider) UpdateAPIEndpoint(_ context.Context, spec provider.APISpec) (*provider.APIResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["UpdateAPIEndpoint"]++
	if _, ok := m.endpoints[spec.APIID]; !ok {
		return nil, fmt.Errorf("mock: api endpoint %q not found", spec.APIID)
	}
	result := &provider.APIResult{
		ID:       spec.APIID,
		Endpoint: fmt.Sprintf("https://mock.execute-api.local/%s", spec.Stage),
	}
	m.endpoints[result.ID] = result
	return result, nil
}

func (m *MockProvider) DeleteAPIEndpoint(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["DeleteAPIEndpoint"]++
	if _, ok := m.endpoints[id]; !ok {
		return fmt.Errorf("mock: api endpoint %q not found", id)
	}
	delete(m.endpoints, id)
	return nil
}

func (m *MockProvider) GetAPIEndpoint(_ context.Context, id string) (*provider.APIResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["GetAPIEndpoint"]++
	result, ok := m.endpoints[id]
	if !ok {
		return nil, fmt.Errorf("mock: api endpoint %q not found", id)
	}
	return result, nil
}

// --- Database ---

func (m *MockProvider) CreateDatabase(_ context.Context, spec provider.DatabaseSpec) (*provider.DatabaseResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["CreateDatabase"]++
	result := &provider.DatabaseResult{
		ID: fmt.Sprintf("mock-db-id::%s", spec.Name),
	}
	m.databases[spec.Name] = result
	return result, nil
}

func (m *MockProvider) DeleteDatabase(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["DeleteDatabase"]++
	if _, ok := m.databases[name]; !ok {
		return fmt.Errorf("mock: database %q not found", name)
	}
	delete(m.databases, name)
	return nil
}

func (m *MockProvider) GetDatabase(_ context.Context, name string) (*provider.DatabaseResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["GetDatabase"]++
	result, ok := m.databases[name]
	if !ok {
		return nil, fmt.Errorf("mock: database %q not found", name)
	}
	return result, nil
}

// --- IAM / Identity ---

func (m *MockProvider) CreateExecutionRole(_ context.Context, spec provider.RoleSpec) (*provider.RoleResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["CreateExecutionRole"]++
	result := &provider.RoleResult{
		ID: fmt.Sprintf("mock-role-arn::%s", spec.Name),
	}
	m.roles[spec.Name] = result
	return result, nil
}

func (m *MockProvider) DeleteExecutionRole(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["DeleteExecutionRole"]++
	if _, ok := m.roles[name]; !ok {
		return fmt.Errorf("mock: role %q not found", name)
	}
	delete(m.roles, name)
	return nil
}

func (m *MockProvider) GetExecutionRole(_ context.Context, name string) (*provider.RoleResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["GetExecutionRole"]++
	result, ok := m.roles[name]
	if !ok {
		return nil, fmt.Errorf("mock: role %q not found", name)
	}
	return result, nil
}

// --- Observability ---

func (m *MockProvider) CreateLogGroup(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["CreateLogGroup"]++
	m.logGroups[name] = true
	return nil
}

func (m *MockProvider) DeleteLogGroup(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls["DeleteLogGroup"]++
	if !m.logGroups[name] {
		return fmt.Errorf("mock: log group %q not found", name)
	}
	delete(m.logGroups, name)
	return nil
}
