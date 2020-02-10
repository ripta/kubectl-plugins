package show

import (
	"io/ioutil"

	"github.com/ripta/kubectl-plugins/pkg/apis/r8y/v1alpha1"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type Options struct {
	genopts.IOStreams
	resource.FilenameOptions
	*v1alpha1.ShowConfig

	ChunkSize     int64
	LabelSelector string
	NoHeaders     bool

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
				klog.Infof("extended preferences ShowConfig exists in kubeconfig but could not be parsed: %v", err)
			}
			o.ShowConfig = &v1alpha1.ShowConfig{
				TypeMeta: v1.TypeMeta{
					Kind:       "",
					APIVersion: "",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "Default",
				},
				SearchPaths: []string{
					"$HOME/.kube/show-formatters",
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

	o.NoHeaders = cmdutil.GetFlagBool(cmd, "no-headers")
	return nil
}

// Validate checks the set of flags provided by the user.
func (o *Options) Validate(cmd *cobra.Command) error {
	return nil
}

// Run performs the get operation.
func (o *Options) Run(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	r := f.NewBuilder().Unstructured().
		NamespaceParam(o.Namespace).DefaultNamespace().AllNamespaces(o.AllNamespaces).
		FilenameParam(o.ExplicitNamespace, &o.FilenameOptions).
		LabelSelector(o.LabelSelector).
		RequestChunksOf(o.ChunkSize).ResourceTypeOrNameArgs(true, args...).
		ContinueOnError().Latest().Flatten().
		Do()
	if err := r.Err(); err != nil {
		return err
	}

	sch, err := getScheme()
	if err != nil {
		return err
	}

	c := serializer.NewCodecFactory(sch, serializer.EnableStrict)
	return loadFormat(c, "./examples/resources.yaml")
}

func getScheme() (*runtime.Scheme, error) {
	sch := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(sch); err != nil {
		return nil, err
	}

	return sch, nil
}

func loadFormat(c serializer.CodecFactory, file string) error {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	obj, gvk, err := c.UniversalDecoder().Decode(d, nil, nil)
	if err != nil {
		return err
	}

	klog.Infof("Got object: %+v", obj)
	klog.Infof("Got GVK   : %+v", gvk)
	return nil
}
