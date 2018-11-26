package ssh

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	// ErrNoPreferredAddress is returned when a node has no preferred address.
	ErrNoPreferredAddress = errors.New("node has no preferred address")
	// ErrNoRoutableAddress is returned when a node has preferred addresses, but none of them routable.
	ErrNoRoutableAddress = errors.New("node has no routable address")
)

type runner struct {
	config *Config
}

// Bind adds flags to the command and sets the command's run action.
func (r *runner) Bind(c *cobra.Command) {
	c.RunE = r.Run

	// Legacy (kubectl <= 1.11) flag passing
	legacySelector := os.Getenv("KUBECTL_PLUGINS_LOCAL_FLAG_SELECTOR")

	genopts.NewResourceBuilderFlags().
		WithLabelSelector(legacySelector).
		AddFlags(c.Flags())
	r.config.AddFlags(c.Flags())
}

// Run is the action that checks for configuration, finds the correct node(s)
// and executes ssh against one of them.
func (r *runner) Run(c *cobra.Command, args []string) error {
	if err := r.config.Complete(c, args); err != nil {
		return err
	}
	if err := r.config.Validate(); err != nil {
		return err
	}

	nodes, err := loadNodes(r.config)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return fmt.Errorf("no nodes found")
	}

	node := nodes[0]
	if len(nodes) > 1 {
		fmt.Fprintf(os.Stderr, "Randomly choosing from %d nodes in the cluster\n", len(nodes))
		node = nodes[rand.Intn(len(nodes))]
	}

	return execSSH(node, r.config.Login)
}

// execSSH executes an ssh session against the node.
func execSSH(n v1.Node, login string) error {
	addr, err := getAddressByType(n.Status.Addresses, []v1.NodeAddressType{v1.NodeInternalIP})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Connecting to %s (%s)\n", addr, n.Name)
	args := []string{"ssh"}
	if login != "" {
		args = append(args, "-l", login)
	}
	args = append(args, addr)

	cmd, err := exec.LookPath("ssh")
	if err != nil {
		return err
	}
	return syscall.Exec(cmd, args, os.Environ())
}

// loadNodes finds all nodes matching the name or selectors.
func loadNodes(cfg *Config) ([]v1.Node, error) {
	cs, err := cfg.Clientset()
	if err != nil {
		return nil, err
	}

	if cfg.NodeName != "" {
		node, err := cs.CoreV1().Nodes().Get(cfg.NodeName, metav1.GetOptions{})
		return []v1.Node{*node}, err
	}

	lopts := metav1.ListOptions{}
	if sel := cfg.NodeSelector; sel != nil {
		lopts.LabelSelector = sel.String()
	}

	nodes, err := cs.CoreV1().Nodes().List(lopts)
	if err != nil {
		return nil, err
	}
	return nodes.Items, nil
}

func getAddressByType(addrs []v1.NodeAddress, types []v1.NodeAddressType) (string, error) {
	for _, typ := range types {
		for _, addr := range addrs {
			if addr.Type == typ {
				return addr.Address, nil
			}
		}
	}
	return "", ErrNoPreferredAddress
}
