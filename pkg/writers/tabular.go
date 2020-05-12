package writers

import (
	"io"

	"github.com/liggitt/tabwriter"
)

const (
	TabularMinWidth = 6
	TabularWidth    = 4
	TabularPad      = 3
	TabularPadChar  = ' '
	TabularFlags    = tabwriter.RememberWidths
)

func NewTabular(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, TabularMinWidth, TabularWidth, TabularPad, TabularPadChar, TabularFlags)
}
