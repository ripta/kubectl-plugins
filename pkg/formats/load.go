package formats

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/ripta/kubectl-plugins/pkg/apis/r8y/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func LoadPaths(sch *runtime.Scheme, paths []string) (*FormatBundle, error) {
	fb := &FormatBundle{
		ByName:  make(map[string]*v1alpha1.ShowFormat),
		Decoder: serializer.NewCodecFactory(sch, serializer.EnableStrict).UniversalDecoder(sch.PrioritizedVersionsAllGroups()...),
	}

	for _, path := range safeExpand(paths) {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			sf, err := loadSingle(fb.Decoder, path)
			if err != nil {
				return err
			}

			if _, ok := fb.ByName[sf.GetName()]; ok {
				return errors.Errorf("found multiple formats named %q", sf.GetName())
			}

			fb.ByName[sf.GetName()] = sf
			// TODO(ripta): handle sf.
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return fb, nil
}

func loadSingle(d runtime.Decoder, file string) (*v1alpha1.ShowFormat, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	obj, _, err := d.Decode(b, nil, nil)
	if err != nil {
		return nil, err
	}

	f, ok := obj.(*v1alpha1.ShowFormat)
	if !ok {
		return nil, errors.Errorf("could not convert %+v to *v1alpha1.ShowFormat", obj)
	}
	return f, nil
}

func safeExpand(paths []string) []string {
	sani := []string{}
	for _, path := range paths {
		d := filepath.Clean(os.ExpandEnv(path))
		if s, err := os.Stat(d); !os.IsNotExist(err) && s.IsDir() {
			sani = append(sani, d)
		}
	}
	return sani
}
