package show

import (
	"github.com/spf13/cobra"

	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/cmd/get"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

var (
	usage = `
	# show pods matching app=foo
	%[1]s pods -l app=foo
	`
)

// NewCommand initializes an instance of the show command.
func NewCommand(s genopts.IOStreams) *cobra.Command {
	k := genopts.NewConfigFlags()
	m := cmdutil.NewMatchVersionFlags(k)
	f := cmdutil.NewFactory(m)
	cmd := get.NewCmdGet("", f, s1)

	// cfg := newConfig(s)
	// cmd := &cobra.Command{
	// 	Use:          "kubectl-show [flags] [node-name]",
	// 	Short:        "Show one or more resources in a customizable format",
	// 	Example:      fmt.Sprintf(usage, "kubectl-ssh"),
	// 	SilenceUsage: true,
	// }

	// r := &runner{
	// 	config: cfg,
	// }
	// r.Bind(cmd)
	return cmd
}
