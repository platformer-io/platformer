// Package v1 contains API Schema definitions for the platformer v1 API group.
// +kubebuilder:object:generate=true
// +groupName=platformer.io
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is the group and version for all PlatFormer CRDs.
	GroupVersion = schema.GroupVersion{Group: "platformer.io", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
