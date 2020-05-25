package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/ripta/hypercmd/pkg/hypercmd"
	"github.com/ripta/kubectl-plugins/pkg/show"
	"github.com/ripta/kubectl-plugins/pkg/ssh"

	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cmd := hypercmd.NewCommand()
	s := genopts.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	k := genopts.NewConfigFlags(true)
	k.AddFlags(cmd.Flags())

	m := cmdutil.NewMatchVersionFlags(k)
	m.AddFlags(cmd.Flags())

	f := cmdutil.NewFactory(m)

	cmd.AddCommand(ssh.NewCommand(s))
	cmd.AddCommand(show.NewCommand(f, s))

	if err := cmd.Resolve(os.Args[0], true).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
