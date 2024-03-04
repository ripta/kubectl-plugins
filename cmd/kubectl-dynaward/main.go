package main

import (
	defaultflag "flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/ripta/kubectl-plugins/pkg/dynaward"
)

func main() {
	pflag.CommandLine = pflag.NewFlagSet("kubectl-show", pflag.ExitOnError)
	pflag.CommandLine.AddGoFlagSet(defaultflag.CommandLine)

	klog.InitFlags(nil)
	defer klog.Flush()

	s := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	k := genopts.NewConfigFlags(true)
	k.AddFlags(pflag.CommandLine)

	f := cmdutil.NewFactory(k)
	cmd := dynaward.NewCommand(f, s)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
