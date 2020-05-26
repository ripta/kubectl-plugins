package show

import (
	"github.com/pkg/errors"
	"github.com/ripta/kubectl-plugins/pkg/apis/r8y/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type Options struct {
	genopts.IOStreams
	resource.FilenameOptions
	*v1alpha1.ShowConfig

	ChunkSize     int64
	LabelSelector string
	NoHeaders     bool
	OutputFormats []string

	Namespace         string
	AllNamespaces     bool
	ExplicitNamespace bool
}

// Complete takes the command arguments and factory and infers any remaining options.
func (o *Options) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	var err error

	kcl := f.ToRawKubeConfigLoader()
	if raw, err := kcl.RawConfig(); err == nil {
		sch, err := getScheme()
		if err != nil {
			return err
		}

		if ext, err := getExtendedPreferences(raw, "ShowConfig", sch); err != nil {
			if err != ErrNoPreferences {
				return errors.Wrap(err, "extended preferences ShowConfig exists in kubeconfig but could not be parsed")
			}
			o.ShowConfig = &v1alpha1.ShowConfig{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "Default",
				},
				SearchPaths: []string{
					"$HOME/.kube/show-formats",
				},
			}
		} else {
			o.ShowConfig = ext
		}
	}

	o.Namespace, o.ExplicitNamespace, err = kcl.Namespace()
	if err != nil {
		return err
	}

	if o.AllNamespaces {
		o.ExplicitNamespace = false
	}

	return nil
}

// Validate checks the set of flags provided by the user.
func (o *Options) Validate(cmd *cobra.Command) error {
	return nil
}

func getScheme() (*runtime.Scheme, error) {
	sch := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(sch); err != nil {
		return nil, err
	}

	return sch, nil
}
