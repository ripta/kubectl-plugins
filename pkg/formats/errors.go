package formats

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type NoCompatibleConfigError struct {
	GroupKind      schema.GroupKind
	AllowedFormats []string
	Paths          []string
}

func (e NoCompatibleConfigError) Error() string {
	sort.Strings(e.AllowedFormats)
	sort.Strings(e.Paths)

	msg := fmt.Sprintf("no suitable configuration for %q", e.GroupKind.String())
	if af := e.AllowedFormats; len(af) > 0 {
		msg += fmt.Sprintf(", requested formats %+v", af)
	}
	if fp := e.Paths; len(fp) > 0 {
		msg += fmt.Sprintf(", in search paths %+v", fp)
	}
	return msg
}
