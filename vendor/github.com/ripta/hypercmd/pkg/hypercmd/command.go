package hypercmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// HyperCommand is a container of a root (the hypercommand command itself) and
// all commands under it. It is special from a regular command in that it can
// explode its commands into symlinked binaries.
type HyperCommand struct {
	root *cobra.Command
	cmds []*cobra.Command
}

// NewCommand initializes a new hypercommand to which commands can be added.
func NewCommand() *HyperCommand {
	makeSymlinksFlag := false
	binary := os.Args[0]

	h := &HyperCommand{}
	cmd := &cobra.Command{
		Use:   binary,
		Short: "Run a command in this hyperbinary",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 || !makeSymlinksFlag {
				cmd.Help()
				os.Exit(1)
			}

			return h.MakeSymlinks(binary)
		},
	}
	cmd.Flags().BoolVar(&makeSymlinksFlag, "make-symlinks", makeSymlinksFlag, "create a symlink for each hyperbinary command into the current directory")

	h.root = cmd
	return h
}

// AddCommand adds a new command to the hypercommand.
func (h *HyperCommand) AddCommand(c *cobra.Command) {
	h.root.AddCommand(c)
	h.cmds = append(h.cmds, c)
}

// Commands returns all the commands registered in the hypercommand.
func (h *HyperCommand) Commands() []*cobra.Command {
	return h.cmds
}

// ImportCommands adds all subcommands of an existing command to the hypercommand.
func (h *HyperCommand) ImportCommands(c *cobra.Command) {
	for _, cmd := range c.Commands() {
		h.AddCommand(cmd)
	}
}

// MakeSymlinks will create a symlink in the current working directory for each command.
func (h *HyperCommand) MakeSymlinks(target string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Making %d symlinks to %s in %s\n", len(h.cmds), target, wd)

	errs := []string{}
	for _, cmd := range h.cmds {
		ln := path.Join(wd, cmd.Name())
		fmt.Fprintf(os.Stderr, "Making symlink for %s at %s\n", cmd.Name(), ln)
		if err := os.Symlink(target, ln); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("could not create one or more symlinks:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// Resolve is given a name of a binary and uses it to return the correct command,
// or the hypercommand otherwise.
func (h *HyperCommand) Resolve(name string, withAliases bool) *cobra.Command {
	name = filepath.Base(name)
	for _, cmd := range h.cmds {
		if cmd.Name() == name {
			return cmd
		}
		if withAliases {
			for _, alias := range cmd.Aliases {
				if alias == name {
					return cmd
				}
			}
		}
	}
	return h.root
}
