package formats

import (
	"github.com/ripta/kubectl-plugins/pkg/apis/r8y/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

type FormatBundle struct {
	ByName  map[string]*v1alpha1.ShowFormat
	Decoder runtime.Decoder
}

func (fb *FormatBundle) ToPrinter() (printers.ResourcePrinterFunc, error) {
	return nil, nil
}
