package dynaward

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type Options struct {
	genericiooptions.IOStreams

	Listen string

	Control   bool
	Verbosity VerbosityLevel

	Namespaces        []string
	AllNamespaces     bool
	ExplicitNamespace bool
}

func (o *Options) Bind(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&o.Control, "control", "c", o.Control, "Enable control endpoint")
	cmd.PersistentFlags().StringVarP(&o.Listen, "listen", "L", o.Listen, "Listen IP:port")

	vl := enumflag.New(&o.Verbosity, "verbosity", VerbosityLevelOptions, enumflag.EnumCaseSensitive)
	cmd.PersistentFlags().VarP(vl, "log-level", "l", "Log verbosity level, one of: info, debug, trace (default: info)")
}

// Complete takes the command arguments and factory and infers any remaining options.
func (o *Options) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	kcl := f.ToRawKubeConfigLoader()

	ns, explicit, err := kcl.Namespace()
	if err != nil {
		return err
	}

	if ns != "" {
		o.Namespaces = []string{ns}
		o.ExplicitNamespace = explicit
	}

	// TODO(ripta): exclude kube-system?
	if o.AllNamespaces {
		o.ExplicitNamespace = false
	}

	return nil
}

// Validate checks the set of flags provided by the user.
func (o *Options) Validate(_ *cobra.Command) error {
	if len(o.Namespaces) > 0 && o.AllNamespaces {
		return fmt.Errorf("only one of --namespace or --all-namespaces may be set")
	}

	if !strings.Contains(o.Listen, ":") {
		return fmt.Errorf("expecting listen %q to contain ':' in the form of IP:PORT", o.Listen)
	}

	return nil
}
