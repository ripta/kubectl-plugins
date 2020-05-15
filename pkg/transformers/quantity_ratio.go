package transformers

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
)

func init() {
	registry["QuantityRatio"] = QuantityRatio
}

func QuantityRatio(params map[string]string) TransformFunc {
	sep, ok := params["separator"]
	if !ok {
		sep = " "
	}

	return func(v interface{}) string {
		ins, ok := v.([]interface{})
		if !ok {
			return fmt.Sprintf("%+v", v)
		}

		outs := make([]string, len(ins))
		for i := range ins {
			s, ok := ins[i].(string)
			if ok {
				q, err := resource.ParseQuantity(s)
				if err != nil {
					outs[i] = s + "?"
				} else {
					outs[i] = fmt.Sprintf("%d", q.MilliValue())
				}
			} else {
				outs[i] = fmt.Sprintf("%v", ins[i])
			}
		}
		return strings.Join(outs, sep)
	}
}
