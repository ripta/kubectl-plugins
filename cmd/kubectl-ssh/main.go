package main

import (
	defaultflag "flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/ripta/kubectl-plugins/pkg/ssh"
	"github.com/spf13/pflag"

	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// XXX(ripta): the Kubernetes CLI runtime isn't cobra-ready yet,
	// and still not in 1.13. CBFB to dig further.
	pflag.CommandLine = pflag.NewFlagSet("kubectl-ssh", pflag.ExitOnError)
	pflag.CommandLine.AddGoFlagSet(defaultflag.CommandLine)

	klog.InitFlags(nil)
	defer klog.Flush()

	s := genopts.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	cmd := ssh.NewCommand(s)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
