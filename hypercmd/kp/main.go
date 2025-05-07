package main

import (
	"fmt"
	"os"

	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/ripta/hypercmd/pkg/hypercmd"

	"github.com/ripta/kubectl-plugins/pkg/dynaward"
	"github.com/ripta/kubectl-plugins/pkg/show"
	"github.com/ripta/kubectl-plugins/pkg/ssh"
)

func main() {
	cmd := hypercmd.New("kp")
	cmd.Root().SilenceErrors = true
	cmd.Root().SilenceUsage = true

	s := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	k := genopts.NewConfigFlags(true)
	k.AddFlags(cmd.Root().Flags())

	m := cmdutil.NewMatchVersionFlags(k)
	m.AddFlags(cmd.Root().Flags())

	f := cmdutil.NewFactory(m)

	cmd.AddCommand(dynaward.NewCommand(f, s))
	cmd.AddCommand(ssh.NewCommand(s))
	cmd.AddCommand(show.NewCommand(f, s))

	sub, err := cmd.Resolve(os.Args, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := sub.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
