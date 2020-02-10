package show

import (
	"github.com/pkg/errors"
	"github.com/ripta/kubectl-plugins/pkg/apis/r8y/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/tools/clientcmd/api"
)

// ErrNoPreferences is the singleton returned when kubeconfig contains no extended preferences.
var ErrNoPreferences = errors.New("no preferences in kubeconfig")

func getExtendedPreferences(raw api.Config, name string, sch *runtime.Scheme) (*v1alpha1.ShowConfig, error) {
	ext := raw.Preferences.Extensions
	if ext == nil {
		return nil, ErrNoPreferences
	}

	v, ok := ext[name]
	if !ok {
		return nil, ErrNoPreferences
	}

	u, ok := v.(*runtime.Unknown)
	if !ok {
		return nil, errors.Errorf("could not understand how to load extended preference %q with type %T", name, v)
	}

	c := serializer.NewCodecFactory(sch, serializer.DisableStrict)
	obj, _, err := c.UniversalDecoder(sch.PrioritizedVersionsAllGroups()...).Decode(u.Raw, nil, nil)
	if err != nil {
		return nil, err
	}

	cfg, ok := obj.(*v1alpha1.ShowConfig)
	if !ok {
		return nil, errors.Errorf("could not convert %+v to *v1alpha1.ShowConfig", obj)
	}

	return cfg, nil
}
