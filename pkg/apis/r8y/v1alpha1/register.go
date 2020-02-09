package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/ripta/kubectl-plugins/pkg/apis/r8y"
)

var (
	// GroupVersion is the package's schema
	GroupVersion = schema.GroupVersion{
		Group:   r8y.GroupName,
		Version: "v1alpha1",
	}

	// GroupInternal is the internal representation
	GroupInternal = schema.GroupVersion{
		Group:   r8y.GroupName,
		Version: runtime.APIVersionInternal,
	}

	// SchemeBuilder is a local alias
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme registers this API group and version to a scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

// Resource returns a group-qualified resource given an unqualified one.
func Resource(rsrc string) schema.GroupResource {
	return GroupVersion.WithResource(rsrc).GroupResource()
}

// addKnownTypes adds the list of package-local types to the runtime scheme.
func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(GroupVersion, &ShowFormatter{}, &ShowFormatterList{})
	s.AddKnownTypes(GroupInternal, &ShowFormatter{}, &ShowFormatterList{})
	return nil
}
