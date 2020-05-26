package printers

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/ripta/kubectl-plugins/pkg/transformers"

	"github.com/itchyny/gojq"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

type ColumnDefinition struct {
	Header        string
	Query         string
	CompiledQuery *gojq.Code
	Transformer   transformers.TransformFunc
}

type CustomPrinter struct {
	Columns       []ColumnDefinition
	Decoder       runtime.Decoder
	IgnoreMissing bool
	NoHeaders     bool

	prevType reflect.Type
}

func (c *CustomPrinter) PrintObj(o runtime.Object, w io.Writer) error {
	t := reflect.TypeOf(o)
	if !c.NoHeaders && t != c.prevType {
		hs := make([]string, len(c.Columns))
		for i := range c.Columns {
			hs[i] = c.Columns[i].Header
		}
		fmt.Fprintln(w, strings.Join(hs, "\t"))
		c.prevType = t
	}

	return c.smartPrint(o, w)
}

func (c *CustomPrinter) printSingle(o runtime.Object, w io.Writer) error {
	if u, ok := o.(*runtime.Unknown); ok {
		if len(u.Raw) > 0 {
			var err error
			if o, err = runtime.Decode(c.Decoder, u.Raw); err != nil {
				return fmt.Errorf("can't decode object for printing: %+v (%s)", err, u.Raw)
			}
		}
	}

	b := NewBlock(len(c.Columns))
	for i := range c.Columns {
		var iter gojq.Iter
		if u, ok := o.(runtime.Unstructured); ok {
			iter = c.Columns[i].CompiledQuery.Run(u.UnstructuredContent())
		} else {
			iter = c.Columns[i].CompiledQuery.Run(reflect.ValueOf(o).Elem().Interface())
		}

		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				if c.IgnoreMissing {
					b.AddTo(i, "") // missing
					break
				} else {
					return errors.Wrapf(err, "printing value for column %s", c.Columns[i].Header)
				}
			} else if tr := c.Columns[i].Transformer; tr != nil {
				b.AddTo(i, tr(v))
			} else {
				b.AddTo(i, fmt.Sprintf("%v", v))
			}
		}
	}

	return b.Render(w, "\t")
}

func (c *CustomPrinter) smartPrint(o runtime.Object, w io.Writer) error {
	if meta.IsListType(o) {
		els, err := meta.ExtractList(o)
		if err != nil {
			return err
		}
		for i := range els {
			if err := c.printSingle(els[i], w); err != nil {
				return errors.Wrap(err, "rendering single object")
			}
		}
		return nil
	}

	return c.printSingle(o, w)
}
