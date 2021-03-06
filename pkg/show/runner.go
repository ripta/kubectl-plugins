package show

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/ripta/kubectl-plugins/pkg/formats"
	"github.com/ripta/kubectl-plugins/pkg/utree"
	"github.com/ripta/kubectl-plugins/pkg/writers"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// Run performs the get operation.
func run(o *Options, f cmdutil.Factory, args []string) error {
	klog.V(4).Infof("args = %+v, ns = %s", args, o.Namespace)
	b := f.NewBuilder().Unstructured().
		NamespaceParam(o.Namespace).DefaultNamespace().AllNamespaces(o.AllNamespaces).
		FilenameParam(o.ExplicitNamespace, &o.FilenameOptions).
		LabelSelector(o.LabelSelector).
		RequestChunksOf(o.ChunkSize).ResourceTypeOrNameArgs(true, args...).
		ContinueOnError().Latest().Flatten()

	r := b.Do()
	if err := r.Err(); err != nil {
		return errors.Wrap(err, "performing request")
	}

	sch, err := getScheme()
	if err != nil {
		return errors.Wrap(err, "getting scheme")
	}

	fb, err := formats.LoadPaths(sch, o.ShowConfig.SearchPaths)
	if err != nil {
		return errors.Wrap(err, "loading formats")
	}

	if klog.V(5) {
		for a, fs := range fb.ByAlias {
			for i, f := range fs {
				klog.Infof("ShowFormat bundle %q: %d %s", a, i, f.GetName())
			}
		}
	}

	infos, err := r.Infos()
	if err != nil {
		return errors.Wrap(err, "retrieving resource information")
	}

	if requestSpansMultipleGVKs(infos) {
		klog.V(4).Infof("retrieved %d resources spanning multiple GVKs", len(infos))
	} else {
		klog.V(4).Infof("retrieved %d resources", len(infos))
	}

	fo := formats.Options{
		NoHeaders: o.NoHeaders,
	}
	if o.OutputFormats != nil {
		fo.AllowedFormats = make(map[string]bool)
		for i, k := range o.OutputFormats {
			fo.AllowedFormats[k] = (i == 0)
		}
	}

	if o.Toposort {
		infos = toposort(infos)
		// fo.NoHeaders = true
		// dumpTo(o.IOStreams.Out, infos)
	}
	return printTo(o.IOStreams.Out, fb, infos, fo)
}

func dumpTo(w io.Writer, infos []*resource.Info) {
	objs := &utree.UnstructuredTree{}

	for i := range infos {
		info := infos[i]
		obj := info.Object
		u, ok := obj.(runtime.Unstructured)
		if !ok {
			klog.Infof("object not unstructured: %T", obj)
			continue
		}

		qq := unstructured.Unstructured{Object: u.UnstructuredContent()}
		objs.Add(qq)
	}

	for _, uj := range objs.Roots() {
		klog.Infof("%s%s: %s %s", "", uj.GetKind(), uj.GetName(), uj.GetUID())
		objs.Walk(uj, func(d int, u unstructured.Unstructured) error {
			klog.Infof("%s %s: %s %s", strings.Repeat("»", d+1), u.GetKind(), u.GetName(), u.GetUID())
			return nil
		})
	}
}

func toposort(infos []*resource.Info) []*resource.Info {
	byUID := make(map[types.UID]*resource.Info, 0)
	objs := &utree.UnstructuredTree{}

	for i := range infos {
		info := infos[i]
		obj := info.Object
		u, ok := obj.(runtime.Unstructured)
		if !ok {
			klog.Infof("object not unstructured: %T", obj)
			continue
		}

		qq := unstructured.Unstructured{Object: u.UnstructuredContent()}
		byUID[qq.GetUID()] = info
		objs.Add(qq)
	}

	sorted := make([]*resource.Info, 0, len(infos))
	for _, uj := range objs.Roots() {
		sorted = append(sorted, byUID[uj.GetUID()])
		objs.Walk(uj, func(d int, u unstructured.Unstructured) error {
			info := byUID[u.GetUID()]
			u.SetName(strings.Repeat("  ", d) + "└ " + u.GetName())
			sorted = append(sorted, info)
			return nil
		})
	}

	return sorted
}

func printTo(w io.Writer, fb *formats.FormatBundle, infos []*resource.Info, opts formats.Options) error {
	perr := []error{}

	legacyscheme := runtime.NewScheme()
	t := writers.NewTabular(w)
	for i := range infos {
		info := infos[i]
		printer, err := fb.ToPrinter(info.Mapping, opts)
		if err != nil {
			return errors.Wrapf(err, "retrieving printer for object index %d of total %d", i, len(infos))
		}

		igv := info.Mapping.GroupVersionKind.GroupKind().WithVersion(runtime.APIVersionInternal).GroupVersion()
		internal, err := legacyscheme.ConvertToVersion(info.Object, igv)
		if err != nil {
			klog.V(6).Infof("could not convert to internal version %q: %+v (falling back to external version)", igv, err)
			internal = info.Object
		}

		if err := printer.PrintObj(internal, t); err != nil {
			perr = append(perr, errors.Wrapf(err, "error printing object index %d of total %d", i, len(infos)))
		}

	}
	t.Flush()

	if len(perr) > 0 {
		serr := make([]string, len(perr))
		for i := range perr {
			serr[i] = fmt.Sprintf("  (%d): %v", i, perr[i])
		}
		return fmt.Errorf("ran into %d errors while printing output:\n%s", len(perr), strings.Join(serr, "\n"))
	}
	return nil
}

func requestSpansMultipleGVKs(infos []*resource.Info) bool {
	if len(infos) < 2 {
		return false
	}

	gvk := infos[0].Mapping.GroupVersionKind
	for _, info := range infos {
		if info.Mapping.GroupVersionKind != gvk {
			return true
		}
	}
	return false
}
