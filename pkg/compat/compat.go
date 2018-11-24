package compat

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/util/homedir"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

type pluginFlag struct {
	Name      string `yaml:"name,omitempty"`
	Shorthand string `yaml:"shorthand,omitempty"`
	Desc      string `yaml:"desc,omitempty"`
}

type pluginMeta struct {
	APIVersion string       `yaml:"apiVersion,omitempty"`
	Kind       string       `yaml:"kind,omitempty"`
	Name       string       `yaml:"name,omitempty"`
	Command    string       `yaml:"command,omitempty"`
	ShortDesc  string       `yaml:"shortDesc,omitempty"`
	Flags      []pluginFlag `yaml:"flags,omitempty"`
}

func (r *compat) Bind(c *cobra.Command) *cobra.Command {
	if c.RunE != nil {
		return c
	}
	c.RunE = r.Run

	return c
}

func (r *compat) Run(c *cobra.Command, args []string) error {
	pluginsDir := filepath.Join(homedir.HomeDir(), ".kube", "plugins")
	for _, cmd := range r.gen() {
		if !strings.HasPrefix(cmd.Name(), "kubectl-") {
			continue
		}
		flagDefs, visitor := genVisit(cmd.Name())
		cmd.Flags().VisitAll(visitor)

		trimmedName := strings.TrimPrefix(cmd.Name(), "kubectl-")
		plug := &pluginMeta{
			APIVersion: "kubectl.config.k8s.io/v1alpha1",
			Kind:       "KubectlPluginConfiguration",
			Name:       trimmedName,
			Command:    cmd.Name(),
			ShortDesc:  cmd.Short,
			Flags:      *flagDefs,
		}
		data, err := yaml.Marshal(plug)
		if err != nil {
			return err
		}

		specFile := filepath.Join(pluginsDir, trimmedName, "plugin.yaml")
		if _, err := os.Stat(specFile); !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "skipped %s, because file already exists\n", specFile)
			continue
		}

		os.MkdirAll(filepath.Join(pluginsDir, trimmedName), 0755)
		f, err := os.OpenFile(specFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func genVisit(name string) (*[]pluginFlag, func(*pflag.Flag)) {
	flagDefs := &[]pluginFlag{}
	return flagDefs, func(f *pflag.Flag) {
		*flagDefs = append(*flagDefs, pluginFlag{
			Name:      f.Name,
			Shorthand: f.Shorthand,
			Desc:      f.Usage,
		})
	}
}
