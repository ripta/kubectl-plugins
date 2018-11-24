package ssh

import (
	"fmt"

	"github.com/spf13/cobra"

	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	usage = `
	# ssh into a node by name
	%[1]s ip-10-11-201-130.us-west-2.compute.internal

	# ssh into any node in the "gpu" instance group
	%[1]s -l kops.k8s.io/instance-group=gpu

	# ssh into any master
	%[1]s -l kubernetes.io/role=master
	`
)

// NewCommand initializes an instance of the ssh command.
func NewCommand(s genopts.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kubectl-ssh [flags] [node-name]",
		Short:        "SSH into a specific Kubernetes node in a cluster or into an arbitrary node matching selectors",
		Example:      fmt.Sprintf(usage, "kubectl-ssh"),
		SilenceUsage: true,
	}

	r := &runner{
		config: newConfig(s),
	}
	return r.Bind(cmd)
}
