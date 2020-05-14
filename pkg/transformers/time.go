package transformers

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/duration"
)

func init() {
	registry["TimeToHumanDuration"] = TimeToHumanDuration
}

// TimeToHumanDuration attempts to parse the string as a timestamp.
func TimeToHumanDuration(params map[string]string) TransformFunc {
	return func(v interface{}) string {
		t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", v))
		if err != nil {
			return fmt.Sprintf("<invalid time: %s>", v)
		}
		return duration.HumanDuration(time.Since(t))
	}
}
