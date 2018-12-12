package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// SchemeBuilder is a local alias
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
)

func init() {
	localSchemeBuilder.Register(knownTypes)
}

// Resource returns a group-qualified resource given an unqualified one.
func Resource(rsrc string) schema.GroupResource {
	return GroupVersion.WithResource(rsrc).GroupResource()
}

// knownTypes adds the list of package-local types to the runtime scheme.
func knownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(GroupVersion, &ShowFormatter{}, &ShowFormatterList{})
	metav1.AddToGroupVersion(s, GroupVersion)
	return nil
}
