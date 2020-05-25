package show

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/ripta/kubectl-plugins/pkg/formats"
	"github.com/ripta/kubectl-plugins/pkg/writers"
	"k8s.io/apimachinery/pkg/runtime"
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

	cfg := o.ShowConfig
	if cfg == nil {
		return errors.New("could not load config for kubectl-show (internal error)")
	}

	sch, err := getScheme()
	if err != nil {
		return errors.Wrap(err, "getting scheme")
	}

	fb, err := formats.LoadPaths(sch, cfg.SearchPaths)
	if err != nil {
		return errors.Wrap(err, "loading formats")
	}

	for a, fs := range fb.ByAlias {
		for i, f := range fs {
			klog.V(4).Infof("ShowFormat %q: %d %s", a, i, f.GetName())
		}
	}

	infos, err := r.Infos()
	if err != nil {
		return errors.Wrap(err, "retrieving resource information")
	}

	klog.V(4).Infof("Asked for %d GVK(s) (multiple=%+v)", len(infos), requestSpansMultipleGVKs(infos))
	return printTo(os.Stdout, infos, fb)
}

func printTo(w io.Writer, infos []*resource.Info, fb *formats.FormatBundle) error {
	legacyscheme := runtime.NewScheme()
	t := writers.NewTabular(w)
	for i := range infos {
		info := infos[i]
		printer, err := fb.ToPrinter(info.Mapping)
		if err != nil {
			return errors.Wrapf(err, "retrieving printer for object index #%d of total %d", i, len(infos))
		}

		igv := info.Mapping.GroupVersionKind.GroupKind().WithVersion(runtime.APIVersionInternal).GroupVersion()
		internal, err := legacyscheme.ConvertToVersion(info.Object, igv)
		if err != nil {
			klog.V(6).Infof("could not convert to internal version %q: %+v (falling back to external version)", igv, err)
			internal = info.Object
		}

		printer.PrintObj(internal, t)
	}
	t.Flush()
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
