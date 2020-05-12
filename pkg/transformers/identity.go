package transformers

import "fmt"

func init() {
	registry[""] = Identity
	registry["Identity"] = Identity
}

// Identity is the identity function that always returns its input.
func Identity(v interface{}) string {
	return fmt.Sprintf("%v", v)

}
