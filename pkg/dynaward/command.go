package dynaward

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

func NewCommand(f cmdutil.Factory, s genericiooptions.IOStreams) *cobra.Command {
	o := &Options{
		IOStreams: s,
	}

	cmd := &cobra.Command{
		Use:                   "kubectl-dynaward [ -A ]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Dynamically port-forward into the cluster"),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd))
			cmdutil.CheckErr(o.Run(f))
		},
		SuggestFor: []string{"sh"},
	}

	return cmd
}
