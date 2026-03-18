package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=sapp
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.apiEndpoint`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ServerlessApp is the top-level CRD. Developers declare one of these and
// PlatFormer provisions all required cloud infrastructure automatically.
type ServerlessApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerlessAppSpec   `json:"spec,omitempty"`
	Status ServerlessAppStatus `json:"status,omitempty"`
}

// ServerlessAppSpec defines the desired state of a ServerlessApp.
type ServerlessAppSpec struct {
	// Runtime is the Lambda execution runtime. MVP supports "nodejs20.x" only.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=nodejs20.x
	Runtime string `json:"runtime"`

	// MemoryMB is the Lambda memory allocation in megabytes (128–10240).
	// +kubebuilder:validation:Minimum=128
	// +kubebuilder:validation:Maximum=10240
	// +kubebuilder:default=512
	MemoryMB int32 `json:"memoryMB,omitempty"`

	// TimeoutSecs is the maximum Lambda execution duration in seconds (1–900).
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=900
	// +kubebuilder:default=30
	TimeoutSecs int32 `json:"timeoutSecs,omitempty"`

	// Code points to the deployment artifact in S3.
	// +kubebuilder:validation:Required
	Code CodeSource `json:"code"`

	// Environment variables injected into the Lambda function.
	// +optional
	Environment map[string]string `json:"environment,omitempty"`

	// API configures the HTTP API Gateway endpoint.
	// +optional
	API *APIConfig `json:"api,omitempty"`

	// Database configures optional DynamoDB tables for this app.
	// +optional
	Database *DatabaseConfig `json:"database,omitempty"`
}

// CodeSource points to the Lambda deployment artifact.
type CodeSource struct {
	// S3Bucket is the bucket containing the deployment zip/image.
	// +kubebuilder:validation:Required
	S3Bucket string `json:"s3Bucket"`

	// S3Key is the object key of the deployment artifact.
	// +kubebuilder:validation:Required
	S3Key string `json:"s3Key"`
}

// APIConfig configures the HTTP API Gateway in front of the Lambda.
type APIConfig struct {
	// Enabled controls whether an API Gateway endpoint is provisioned.
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Stage is the API Gateway deployment stage name.
	// +kubebuilder:default="prod"
	Stage string `json:"stage,omitempty"`
}

// DatabaseConfig configures DynamoDB tables for the app.
type DatabaseConfig struct {
	// Tables is the list of DynamoDB tables to provision.
	// +kubebuilder:validation:MinItems=1
	Tables []TableSpec `json:"tables"`
}

// TableSpec defines a single DynamoDB table.
type TableSpec struct {
	// Name is the DynamoDB table name.
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

// ServerlessAppStatus defines the observed state written back by the controller.
type ServerlessAppStatus struct {
	// Phase is the high-level lifecycle state of the app.
	// +kubebuilder:validation:Enum=Pending;Provisioning;Ready;Failed;Deleting
	Phase string `json:"phase,omitempty"`

	// FunctionARN is the ARN of the provisioned Lambda function.
	FunctionARN string `json:"functionARN,omitempty"`

	// FunctionVersion is the published Lambda version.
	FunctionVersion string `json:"functionVersion,omitempty"`

	// ExecutionRoleARN is the ARN of the IAM execution role.
	ExecutionRoleARN string `json:"executionRoleARN,omitempty"`

	// APIEndpoint is the public HTTPS URL of the API Gateway endpoint.
	APIEndpoint string `json:"apiEndpoint,omitempty"`

	// APIID is the API Gateway API ID.
	APIID string `json:"apiID,omitempty"`

	// LogGroupName is the CloudWatch log group for this function.
	LogGroupName string `json:"logGroupName,omitempty"`

	// Conditions holds standard Kubernetes condition objects.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the .metadata.generation the controller last reconciled.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

// ServerlessAppList contains a list of ServerlessApp.
type ServerlessAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServerlessApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServerlessApp{}, &ServerlessAppList{})
}
