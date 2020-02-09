package show

import (
	"strings"

	"github.com/spf13/cobra"

	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var (
	usage = `
	# show pods matching app=foo
	%[1]s pods -l app=foo
	`
)

// NewCommand initializes an instance of the show command.
func NewCommand(s genopts.IOStreams) *cobra.Command {
	k := genopts.NewConfigFlags(false)
	m := cmdutil.NewMatchVersionFlags(k)
	f := cmdutil.NewFactory(m)

	cmd := get.NewCmdGet("", f, s)
	cmd.Use = "kubectl-show (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...) [flags]"
	cmd.Example = strings.Replace(cmd.Example, " get ", " show ", -1)
	cmd.SuggestFor = []string{}

	cmd.Flags().Set("output", "custom-columns")

	// cfg := newConfig(s)
	// r := &runner{
	// 	config: cfg,
	// }
	// r.Bind(cmd)
	return cmd
}
