package transformers

import "fmt"

type TransformFunc func(interface{}) string

type ParameterizedTransformFunc func(params map[string]string) TransformFunc

var registry = make(map[string]ParameterizedTransformFunc)

// Lookup returns a transformation function by name, or an error.
func Lookup(name string) (ParameterizedTransformFunc, error) {
	fn, ok := registry[name]
	if !ok {
		names := make([]string, len(registry))
		for k := range registry {
			names = append(names, k)
		}
		return nil, fmt.Errorf("transform function %q not found; available: %+v", name, names)
	}
	return fn, nil
}
