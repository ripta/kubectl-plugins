package show

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/ripta/kubectl-plugins/pkg/apis/r8y/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/tools/clientcmd/api"
)

// ErrNoPreferences is the singleton returned when kubeconfig contains no extended preferences.
var ErrNoPreferences = errors.New("no preferences in kubeconfig")

func getShowConfigPreferences(raw api.Config, sch *runtime.Scheme) (*v1alpha1.ShowConfig, error) {
	ext := raw.Preferences.Extensions
	if ext == nil {
		return nil, ErrNoPreferences
	}

	errs := []error{}
	for name, v := range ext {
		u, ok := v.(*runtime.Unknown)
		if !ok {
			errs = append(errs, errors.Errorf("could not understand how to load extended preference %q with type %T", name, v))
			continue
		}

		c := serializer.NewCodecFactory(sch, serializer.DisableStrict)
		obj, _, err := c.UniversalDecoder(sch.PrioritizedVersionsAllGroups()...).Decode(u.Raw, nil, nil)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "universal decoder for preference %q", name))
			continue
		}

		cfg, ok := obj.(*v1alpha1.ShowConfig)
		if !ok {
			errs = append(errs, errors.Wrapf(err, "could not convert preference %q from %T (%+v)", name, obj, obj))
			continue
		}

		return cfg, nil
	}

	if len(errs) > 0 {
		serrs := make([]string, len(errs))
		for i, err := range errs {
			serrs[i] = fmt.Sprintf("  (%d) %v", i, err)
		}
		return nil, errors.Errorf("the following errors occurred:\n%s", strings.Join(serrs, "\t"))
	}

	return nil, ErrNoPreferences
}
