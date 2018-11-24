package compat

import (
	"github.com/spf13/cobra"
)

type compat struct {
	gen func() []*cobra.Command
}

// NewGeneratorCommand builds a new gen-legacy-plugins command.
func NewGeneratorCommand(gen func() []*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "gen-legacy-plugins [flags] [node-name]",
		Short:        "Generate legacy plugin definition files in $HOME/.kube/plugins",
		SilenceUsage: true,
	}

	r := &compat{
		gen: gen,
	}
	return r.Bind(cmd)
}
