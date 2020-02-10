package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ShowConfig ...
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ShowConfig struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	SearchPaths []string `json:"searchPaths,omitempty"`
}

// ShowFormatter is the top-level type representing a set of formatter options.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ShowFormatter struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ShowFormatterSpec `json:"spec,omitempty"`
}

// ShowFormatterSpec ...
type ShowFormatterSpec struct {
	Aliases  []string                  `json:"aliases,omitempty"`
	Defaults ShowFormatterDefaultsSpec `json:"defaults,omitempty"`
	Fields   []FieldSpec               `json:"fields,omitempty"`
}

// ShowFormatterDefaultsSpec ...
type ShowFormatterDefaultsSpec struct {
	IgnoreUnknownFields bool     `json:"ignoreUnknownFields,omitempty"`
	SortBy              []string `json:"sortBy"`
}

// ShowFormatterList is the autogenerated list type.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ShowFormatterList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ShowFormatter `json:"items"`
}

// FieldSpec defines the specification of a column
type FieldSpec struct {
	Name      string `json:"name,omitempty"`
	Label     string `json:"label,omitempty"`
	JSONPath  string `json:"jsonPath,omitempty"`
	Formatter string `json:"formatter,omitempty"`
}
