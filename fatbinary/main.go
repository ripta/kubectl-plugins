package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/ripta/kubectl-plugins/pkg/hypercmd"
	"github.com/ripta/kubectl-plugins/pkg/ssh"
	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	kubectl "k8s.io/kubernetes/pkg/kubectl/cmd"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	s := genopts.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	cmd := hypercmd.NewCommand()
	cmd.AddCommand(ssh.NewCommand(s))

	kc := kubectl.NewKubectlCommand(os.Stdin, os.Stdout, os.Stderr)
	cmd.ImportCommands(kc)

	if err := cmd.Resolve(os.Args[0], true).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
