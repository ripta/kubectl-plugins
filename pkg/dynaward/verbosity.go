package dynaward

import "github.com/thediveo/enumflag/v2"

type VerbosityLevel enumflag.Flag

const (
	InfoVerbosityLevel VerbosityLevel = iota
	DebugVerbosityLevel
	TraceVerbosityLevel
)

var VerbosityLevelOptions = map[VerbosityLevel][]string{
	InfoVerbosityLevel:  {"", "i", "info"},
	DebugVerbosityLevel: {"d", "debug"},
	TraceVerbosityLevel: {"t", "trace"},
}
