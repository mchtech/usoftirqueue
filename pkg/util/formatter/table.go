package formatter

import (
	"fmt"
	"sort"
	"strings"
)

type comparable interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string
}

type Table[TK comparable, TC any] struct {
	rows     []item[TK, TC]
	maxWidth []int
}

type item[TK comparable, TC any] struct {
	colume1 colume1[TK]
	columes colume[TC]
}

type colume1[T comparable] struct {
	raw T
	str string
}

type colume[T any] struct {
	raw T
	str string
}

// impl sort.Interface
func (t *Table[TK, TC]) Len() int {
	return len(t.rows)
}
func (t *Table[TK, TC]) Less(i, j int) bool {
	return t.rows[i].colume1.raw < t.rows[j].colume1.raw
}
func (t *Table[TK, TC]) Swap(i, j int) {
	t.rows[i], t.rows[j] = t.rows[j], t.rows[i]
}

func NewTableFromMap[TK comparable, TC any](m map[TK]TC, header []string) *Table[TK, TC] {
	i := 0
	t := &Table[TK, TC]{
		rows:     make([]item[TK, TC], len(m)),
		maxWidth: []int{0, 0},
	}
	for k, v := range m {
		t.rows[i] = item[TK, TC]{
			colume1: colume1[TK]{raw: k, str: fmt.Sprintf("%v", k)},
			columes: colume[TC]{raw: v, str: fmt.Sprintf("%v", v)},
		}
		i++
	}
	sort.Sort(t)

	t.rows = append([]item[TK, TC]{{colume1: colume1[TK]{str: header[0]}, columes: colume[TC]{str: header[1]}}}, t.rows...)
	for _, row := range t.rows {
		if l := len(row.colume1.str); l > t.maxWidth[0] {
			t.maxWidth[0] = l
		}
		if l := len(row.columes.str); l > t.maxWidth[1] {
			t.maxWidth[1] = l
		}
	}
	return t
}

func (t *Table[TK, TC]) String() string {
	var b strings.Builder
	l := len(t.rows)

	p := func(row item[TK, TC]) {
		fmt.Fprintf(&b, "│ %-*s │ %-*s │\n", t.maxWidth[0], row.colume1.str, t.maxWidth[1], row.columes.str)
	}

	for i, row := range t.rows {
		if i == 0 {
			fmt.Fprintf(&b, "┌─%s─┬─%s─┐\n", strings.Repeat("─", t.maxWidth[0]), strings.Repeat("─", t.maxWidth[1]))
			p(row)
			fmt.Fprintf(&b, "├─%s─┼─%s─┤\n", strings.Repeat("─", t.maxWidth[0]), strings.Repeat("─", t.maxWidth[1]))
		} else if i < l-1 {
			p(row)
			fmt.Fprintf(&b, "├─%s─┼─%s─┤\n", strings.Repeat("─", t.maxWidth[0]), strings.Repeat("─", t.maxWidth[1]))
		} else {
			p(row)
			fmt.Fprintf(&b, "└─%s─┴─%s─┘\n", strings.Repeat("─", t.maxWidth[0]), strings.Repeat("─", t.maxWidth[1]))
		}
	}
	return b.String()
}
