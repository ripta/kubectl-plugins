package formats

// Options captures formatting options.
type Options struct {
	// AllowedFormats is a map of allowed formats. The value is ignored.
	AllowedFormats map[string]bool
	// NoHeaders is whether headers should be printed or not.
	NoHeaders bool
}

// Allows returns whether the format container is allowed by options or not.
func (opts Options) Allows(fc *FormatContainer) bool {
	if opts.AllowedFormats == nil {
		return false
	}
	for _, a := range fc.Spec.Aliases {
		if _, ok := opts.AllowedFormats[a]; ok {
			return true
		}
	}
	return false
}
