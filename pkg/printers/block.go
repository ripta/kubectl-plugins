package printers

import (
	"fmt"
	"io"
	"strings"
)

type Row []string

type Block struct {
	Rows    []Row
	numCols int
	colFill []int
}

func NewBlock(numCols int) *Block {
	return &Block{
		Rows:    make([]Row, 0),
		numCols: numCols,
		colFill: make([]int, numCols),
	}
}

func (b *Block) AddTo(col int, v string) error {
	if col >= b.numCols {
		return fmt.Errorf("column index overflow %d (max index: %d)", col, b.numCols-1)
	}

	row := b.colFill[col]
	if len(b.Rows) <= row {
		b.Rows = append(b.Rows, make(Row, b.numCols))
	}

	b.Rows[row][col] = v
	b.colFill[col]++
	return nil
}

func (b *Block) Render(w io.Writer, sep string) error {
	for i := range b.Rows {
		if _, err := fmt.Fprintln(w, strings.Join(b.Rows[i], sep)); err != nil {
			return err
		}
	}
	return nil
}
