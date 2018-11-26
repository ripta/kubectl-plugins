package ssh

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/labels"
	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// Config holds configuration for kubectl-ssh.
type Config struct {
	flags *genopts.ConfigFlags
	args  []string

	Login        string
	NodeName     string
	NodeSelector labels.Selector

	genopts.IOStreams
}

func newConfig(s genopts.IOStreams) *Config {
	return &Config{
		flags:     genopts.NewConfigFlags(),
		IOStreams: s,
	}
}

// AddFlags decorates the flagset with config flags.
func (c *Config) AddFlags(flagset *pflag.FlagSet) {
	c.flags.AddFlags(flagset)
}

// Clientset creates a new client based on the REST client configuration inferred
// from the environment or set through flags.
func (c *Config) Clientset() (*kubernetes.Clientset, error) {
	rest, err := c.flags.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(rest)
}

// Complete sets config fields based on flags and arguments
func (c *Config) Complete(cmd *cobra.Command, args []string) error {
	c.args = args

	lab, err := cmd.Flags().GetString("selector")
	if err != nil {
		return err
	}

	if lab != "" {
		sel, err := labels.Parse(lab)
		if err != nil {
			return err
		}

		c.NodeSelector = sel
	}

	if len(args) > 0 {
		c.NodeName = args[0]
	}
	return nil
}

// Validate checks that required arguments and flags are provided
func (c *Config) Validate() error {
	if c.NodeSelector == nil && c.NodeName == "" {
		return fmt.Errorf("either a node name or selectors are required; try 'get nodes --show-labels' for a full listing")
	}
	if c.NodeSelector != nil && c.NodeName != "" {
		return fmt.Errorf("either a node name or selectors, but not both, are required")
	}
	return nil
}
