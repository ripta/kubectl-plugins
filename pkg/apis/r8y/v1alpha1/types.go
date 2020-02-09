package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ShowConfig ...
type ShowConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	SearchPaths []string `json:"searchPaths,omitempty"`
}

// ShowFormatter is the top-level type representing a set of formatter options.
// +k8s:openapi-gen=true
// +resource:path=formatter
type ShowFormatter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ShowFormatterSpec `json:"spec,omitempty"`
}

type ShowFormatterSpec struct {
	Aliases  []string                  `json:"aliases,omitempty"`
	Defaults ShowFormatterDefaultsSpec `json:"defaults,omitempty"`
	Fields   []FieldSpec               `json:"fields,omitempty"`
}

type ShowFormatterDefaultsSpec struct {
	IgnoreUnknownFields bool     `json:"ignoreUnknownFields,omitempty"`
	SortBy              []string `json:"sortBy"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ShowFormatterList is the autogenerated list type.
type ShowFormatterList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Items []ShowFormatter `json:"items"`
}

// FieldSpec defines the specification of a column
type FieldSpec struct {
	Name      string `json:"name,omitempty"`
	Label     string `json:"label,omitempty"`
	JSONPath  string `json:"jsonPath,omitempty"`
	Formatter string `json:"formatter,omitempty"`
}
