package show

import (
	"github.com/spf13/cobra"

	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

var (
	usage = `
	# show pods matching app=foo
	%[1]s pods -l app=foo
	`
)

// NewCommand initializes an instance of the show command.
func NewCommand(f cmdutil.Factory, s genopts.IOStreams) *cobra.Command {
	o := &Options{
		IOStreams:     s,
		ChunkSize:     500,
		OutputFormats: make([]string, 0),
	}

	cmd := &cobra.Command{
		Use:                   "kubectl-show (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...) [flags]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Display one or many resources"),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd))
			if o.ListFormats {
				cmdutil.CheckErr(listFormats(o))
			} else {
				cmdutil.CheckErr(run(o, f, args))
			}
		},
		SuggestFor: []string{"sh"},
	}

	cmd.Flags().Int64Var(&o.ChunkSize, "chunk-size", o.ChunkSize, "Return large lists in chunks rather than all at once. Pass 0 to disable. This flag is beta and may change in the future.")
	cmd.Flags().BoolVarP(&o.NoHeaders, "no-headers", "H", o.NoHeaders, "Hide headers")

	cmd.Flags().StringSliceVarP(&o.OutputFormats, "output", "o", o.OutputFormats, "Allowed format names (use --list-outputs to see)")
	cmd.Flags().BoolVar(&o.ListFormats, "list-outputs", o.ListFormats, "List all available format names for use in --output")

	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", o.AllNamespaces, "List the requested objects across all namespaces. The namespace in the current context is ignored.")
	cmd.Flags().StringVarP(&o.LabelSelector, "selector", "l", o.LabelSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")

	return cmd
}
