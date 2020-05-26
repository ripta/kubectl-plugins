package formats

import (
	"github.com/itchyny/gojq"
	"github.com/pkg/errors"
	"github.com/ripta/kubectl-plugins/pkg/apis/r8y/v1alpha1"
	"github.com/ripta/kubectl-plugins/pkg/printers"
	"github.com/ripta/kubectl-plugins/pkg/transformers"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cliprinters "k8s.io/cli-runtime/pkg/printers"
	"k8s.io/klog"
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

// ToPrinter selects the best resource printer based on GVK and formatting options.
func (fb *FormatBundle) ToPrinter(mapping *meta.RESTMapping, opts Options) (cliprinters.ResourcePrinterFunc, error) {
	gk := mapping.GroupVersionKind.GroupKind()
	of := gk.String()

	// limit to format containers that understand how to print the groupkind in question
	klog.V(4).Infof("searching for formatter matching %+v", opts.AllowedFormats)
	fcs, ok := fb.ByGroupKind[gk]
	if !ok {
		return nil, genericclioptions.NoCompatiblePrinterError{OutputFormat: &of}
	}
	if len(fcs) < 1 {
		return nil, genericclioptions.NoCompatiblePrinterError{OutputFormat: &of}
	}

	// select best format container based on options
	for i := range fcs {
		if opts.Allows(fcs[i]) {
			klog.V(3).Infof("chose output formatter %q", fcs[i].Name)
			return fcs[i].ToPrinter(opts)
		}
	}

	// choose first as default
	klog.V(4).Info("falling back to default formatter")
	return fcs[0].ToPrinter(opts)
}

type FormatContainer struct {
	*v1alpha1.ShowFormat
	Path string

	prevPrinter cliprinters.ResourcePrinterFunc
}

func (fc *FormatContainer) ToPrinter(opts Options) (cliprinters.ResourcePrinterFunc, error) {
	if fc.prevPrinter != nil {
		return fc.prevPrinter, nil
	}

	// Transform ShowFormatter field specifications into custom column formatters
	cs := make([]printers.ColumnDefinition, len(fc.Spec.Fields))
	for i := range fc.Spec.Fields {
		q, err := gojq.Parse(fc.Spec.Fields[i].Query)
		if err != nil {
			return nil, errors.Wrapf(err, "query for column %q", fc.Spec.Fields[i].Name)
		}
		code, err := gojq.Compile(q)
		if err != nil {
			return nil, errors.Wrapf(err, "compiling query for column %q", fc.Spec.Fields[i].Name)
		}

		cs[i] = printers.ColumnDefinition{
			Header:        fc.Spec.Fields[i].Label,
			Query:         fc.Spec.Fields[i].Query,
			CompiledQuery: code,
		}

		if t := fc.Spec.Fields[i].Transformer; t != nil {
			fn, err := transformers.Lookup(t.Name)
			if err != nil {
				klog.V(1).Infof("no transformer for column %q: %+v", fc.Spec.Fields[i].Name, err)
				fn = transformers.Identity
			}
			cs[i].Transformer = fn(t.Params)
		}
	}

	// Prevent decoding into internal versions by specifying version parameters
	d := scheme.Codecs.UniversalDecoder(scheme.Scheme.PrioritizedVersionsAllGroups()...)

	// Piggy-back onto custom column implementation
	ccp := printers.CustomPrinter{
		Columns:   cs,
		Decoder:   d,
		NoHeaders: opts.NoHeaders,
	}
	fc.prevPrinter = ccp.PrintObj
	return ccp.PrintObj, nil
}
