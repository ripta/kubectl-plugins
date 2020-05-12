package show

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/ripta/kubectl-plugins/pkg/formats"
	"github.com/ripta/kubectl-plugins/pkg/writers"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func run(o *Options, f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	klog.V(4).Infof("args = %+v, ns = %s", args, o.Namespace)
	r := f.NewBuilder().Unstructured().
		NamespaceParam(o.Namespace).DefaultNamespace().AllNamespaces(o.AllNamespaces).
		FilenameParam(o.ExplicitNamespace, &o.FilenameOptions).
		LabelSelector(o.LabelSelector).
		RequestChunksOf(o.ChunkSize).ResourceTypeOrNameArgs(true, args...).
		ContinueOnError().Latest().Flatten().
		Do()
	if err := r.Err(); err != nil {
		return errors.Wrap(err, "initializing builder")
	}

	sch, err := getScheme()
	if err != nil {
		return errors.Wrap(err, "getting scheme")
	}

	cfg := o.ShowConfig
	if cfg == nil {
		return errors.New("could not get config for kubectl-show (internal error)")
	}

	fb, err := formats.LoadPaths(sch, cfg.SearchPaths)
	if err != nil {
		return errors.Wrap(err, "loading formats")
	}

	// klog.V(4).Infof("kubectl-show format bundle: %+v", fb)
	//
	// for _, f := range fb.ByName {
	// 	klog.Infof("Format %s: %+v", f.GetName(), f.Spec)
	// }
	//
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
