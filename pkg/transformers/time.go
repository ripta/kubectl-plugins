package transformers

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/duration"
)

func init() {
	registry["DurationToHumanDuration"] = DurationToHumanDuration
	registry["TimeToHumanDuration"] = TimeToHumanDuration
}

// DurationToHumanDuration attempts to parse the string as a timestamp.
func DurationToHumanDuration(params map[string]string) TransformFunc {
	return func(v interface{}) string {
		if v == nil {
			if def, ok := params["whenEmpty"]; ok {
				return def
			}
			return ""
		}
		switch n := v.(type) {
		case int64:
			return duration.HumanDuration(time.Duration(n) * time.Second)
		case float64:
			return duration.HumanDuration(time.Duration(int64(n)) * time.Second)
		}
		return fmt.Sprintf("<invalid duration: %v (%T)>", v, v)
	}
}

// TimeToHumanDuration attempts to parse the string as a timestamp.
func TimeToHumanDuration(params map[string]string) TransformFunc {
	return func(v interface{}) string {
		if v == nil {
			if def, ok := params["whenEmpty"]; ok {
				return def
			}
			return ""
		}
		t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", v))
		if err != nil {
			return fmt.Sprintf("<invalid time: %s>", v)
		}
		return duration.HumanDuration(time.Since(t))
	}
}
