package dynaward

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type Options struct {
	genericiooptions.IOStreams

	Namespaces        []string
	AllNamespaces     bool
	ExplicitNamespace bool
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

	return nil
}
