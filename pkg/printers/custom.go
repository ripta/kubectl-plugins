package printers

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/itchyny/gojq"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
)

type ColumnDefinition struct {
	Header      string
	Query       string
	Transformer func(string) string
}

type CustomPrinter struct {
	Columns   []ColumnDefinition
	Decoder   runtime.Decoder
	NoHeaders bool
	prevType  reflect.Type
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

	codes := make([]*gojq.Code, len(c.Columns))
	for i := range c.Columns {
		q, err := gojq.Parse(c.Columns[i].Query)
		if err != nil {
			return errors.Wrapf(err, "query for column %q", c.Columns[i].Header)
		}
		code, err := gojq.Compile(q)
		if err != nil {
			return errors.Wrapf(err, "compiling query for column %q", c.Columns[i].Header)
		}
		codes[i] = code
	}

	return c.smartPrint(codes, o, w)
}

func (c *CustomPrinter) printSingle(cs []*gojq.Code, o runtime.Object, w io.Writer) error {
	if u, ok := o.(*runtime.Unknown); ok {
		if len(u.Raw) > 0 {
			var err error
			if o, err = runtime.Decode(c.Decoder, u.Raw); err != nil {
				return fmt.Errorf("can't decode object for printing: %+v (%s)", err, u.Raw)
			}
		}
	}

	b := NewBlock(len(cs))
	for i := range cs {
		var iter gojq.Iter
		if u, ok := o.(runtime.Unstructured); ok {
			iter = cs[i].Run(u.UnstructuredContent())
		} else {
			iter = cs[i].Run(reflect.ValueOf(o).Elem().Interface())
		}

		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				return errors.Wrapf(err, "rendering single object")
			}
			if t, ok := v.(time.Time); ok {
				b.AddTo(i, duration.ShortHumanDuration(time.Since(t)))
			} else {
				b.AddTo(i, fmt.Sprintf("%v", v))
			}
		}
	}

	return b.Render(w, "\t")
}

func (c *CustomPrinter) smartPrint(cs []*gojq.Code, o runtime.Object, w io.Writer) error {
	if meta.IsListType(o) {
		els, err := meta.ExtractList(o)
		if err != nil {
			return err
		}
		for i := range els {
			if err := c.printSingle(cs, els[i], w); err != nil {
				return err
			}
		}
		return nil
	}

	return c.printSingle(cs, o, w)
}
