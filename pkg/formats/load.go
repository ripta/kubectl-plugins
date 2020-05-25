package formats

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/ripta/kubectl-plugins/pkg/apis/r8y/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func LoadPaths(sch *runtime.Scheme, paths []string) (*FormatBundle, error) {
	fb := &FormatBundle{
		ByAlias:     make(map[string][]*FormatContainer),
		ByName:      make(map[string]*FormatContainer),
		ByGroupKind: make(map[schema.GroupKind][]*FormatContainer),
		Decoder:     serializer.NewCodecFactory(sch, serializer.EnableStrict).UniversalDecoder(sch.PrioritizedVersionsAllGroups()...),
	}

	for _, path := range safeExpand(paths) {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			fc, err := loadSingle(fb.Decoder, path)
			if err != nil {
				return errors.Wrapf(err, "from path %s", path)
			}

			if err := fb.add(fc); err != nil {
				return errors.Wrapf(err, "adding format %s to bundle", path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return fb, nil
}

func loadSingle(d runtime.Decoder, file string) (*FormatContainer, error) {
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
	return &FormatContainer{
		ShowFormat: f,
		Path:       file,
	}, nil
}

func safeExpand(paths []string) []string {
	sani := []string{}
	for _, path := range paths {
		path = strings.Replace(path, "~", "$HOME", 1)
		d := filepath.Clean(os.ExpandEnv(path))
		if s, err := os.Stat(d); !os.IsNotExist(err) && s.IsDir() {
			sani = append(sani, d)
		}
	}
	return sani
}
