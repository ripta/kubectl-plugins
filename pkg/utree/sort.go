package utree

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type SortedUnstructured []unstructured.Unstructured

func (s SortedUnstructured) Len() int {
	return len(s)
}

func (s SortedUnstructured) Less(i, j int) bool {
	a, b := s[i], s[j]
	if a.GetKind() != b.GetKind() {
		return a.GetKind() < b.GetKind()
	}
	return a.GetName() < b.GetName()
}

func (s SortedUnstructured) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
