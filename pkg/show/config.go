package show

import (
	"io/ioutil"

	"github.com/ripta/kubectl-plugins/pkg/apis/r8y/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	genopts "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type Options struct {
	genopts.IOStreams
	resource.FilenameOptions
	api.Preferences

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
		o.Preferences = raw.Preferences

		v, ok := o.Preferences.Extensions["ShowFormatter"]
		if ok {
			klog.Infof("Preferences: %+v", v)
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

	return loadFormat("./examples/resources.yaml")
}

func loadFormat(file string) error {
	s := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(s); err != nil {
		return err
	}

	for gvk := range s.AllKnownTypes() {
		klog.Infof("Registered GVK: %+v", gvk)
	}

	c := serializer.NewCodecFactory(s, serializer.EnableStrict)

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
