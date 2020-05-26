package show

import (
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/pkg/errors"
	"github.com/ripta/kubectl-plugins/pkg/formats"
	"github.com/ripta/kubectl-plugins/pkg/writers"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func listFormats(o *Options) error {
	sch, err := getScheme()
	if err != nil {
		return errors.Wrap(err, "getting scheme")
	}

	fb, err := formats.LoadPaths(sch, o.ShowConfig.SearchPaths)
	if err != nil {
		return errors.Wrap(err, "loading formats")
	}

	t := writers.NewTabular(o.IOStreams.Out)
	headers := []string{"NAME", "GVK", "PATH"}
	fmt.Fprintf(t, strings.Join(headers, "\t")+"\n")

	for _, fcs := range fb.ByGroupKind {
		for _, fc := range fcs {
			fields := []string{
				fc.Spec.Aliases[0],
				fc.Spec.ComponentKinds[0].String(),
				fc.Path,
			}
			fmt.Fprintf(t, strings.Join(fields, "\t")+"\n")
		}
	}

	return t.Flush()
}

func mustCompile(s string) *gojq.Code {
	q, err := gojq.Parse(s)
	if err != nil {
		cmdutil.CheckErr(errors.Wrapf(err, "parsing query %q", s))
	}
	code, err := gojq.Compile(q)
	if err != nil {
		cmdutil.CheckErr(errors.Wrapf(err, "compiling query %q", s))
	}
	return code
}
