package formats

import (
	"github.com/pkg/errors"
	"github.com/ripta/kubectl-plugins/pkg/apis/r8y/v1alpha1"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/cmd/get"
	"k8s.io/kubectl/pkg/scheme"
)

type FormatBundle struct {
	ByAlias     map[string][]*FormatContainer
	ByGroupKind map[schema.GroupKind][]*FormatContainer
	ByName      map[string]*FormatContainer
	Decoder     runtime.Decoder
}

func (fb *FormatBundle) add(fc *FormatContainer) error {
	if dupe, ok := fb.ByName[fc.GetName()]; ok {
		return errors.Errorf("found multiple formats named %q (previously %q, now %q)", fc.GetName(), dupe.Path, fc.Path)
	}

	fb.ByName[fc.GetName()] = fc

	for _, mgk := range fc.Spec.ComponentKinds {
		gk := schema.GroupKind{
			Group: mgk.Group,
			Kind:  mgk.Kind,
		}
		if _, ok := fb.ByGroupKind[gk]; !ok {
			fb.ByGroupKind[gk] = make([]*FormatContainer, 0)
		}
		fb.ByGroupKind[gk] = append(fb.ByGroupKind[gk], fc)
	}

	for _, a := range fc.Spec.Aliases {
		if _, ok := fb.ByAlias[a]; !ok {
			fb.ByAlias[a] = make([]*FormatContainer, 0)
		}
		fb.ByAlias[a] = append(fb.ByAlias[a], fc)
	}

	return nil
}

func (fb *FormatBundle) ToPrinter(mapping *meta.RESTMapping) (printers.ResourcePrinterFunc, error) {
	gk := mapping.GroupVersionKind.GroupKind()
	of := gk.String()
	fcs, ok := fb.ByGroupKind[gk]
	if !ok {
		return nil, genericclioptions.NoCompatiblePrinterError{OutputFormat: &of}
	}
	if len(fcs) < 1 {
		return nil, genericclioptions.NoCompatiblePrinterError{OutputFormat: &of}
	}
	return fcs[0].ToPrinter()
}

type FormatContainer struct {
	*v1alpha1.ShowFormat
	Path string
}

func (fc *FormatContainer) ToPrinter() (printers.ResourcePrinterFunc, error) {
	// Transform ShowFormatter field specifications into custom column formatters
	cs := make([]get.Column, len(fc.Spec.Fields))
	for i := range fc.Spec.Fields {
		fs, err := get.RelaxedJSONPathExpression(fc.Spec.Fields[i].JSONPath)
		if err != nil {
			return nil, errors.Wrapf(err, "error processing path expression for field %q in %q", fc.Spec.Fields[i].Label, fc.Path)
		}
		cs[i] = get.Column{
			Header:    fc.Spec.Fields[i].Label,
			FieldSpec: fs,
		}
	}

	// Prevent decoding into internal versions by specifying version parameters
	d := scheme.Codecs.UniversalDecoder(scheme.Scheme.PrioritizedVersionsAllGroups()...)

	// Piggy-back onto internal custom column implementation
	ccp := get.CustomColumnsPrinter{
		Columns:   cs,
		Decoder:   d,
		NoHeaders: false,
	}
	return ccp.PrintObj, nil
}
