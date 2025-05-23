// Copyright (c) 2020, Peter Ohler, All rights reserved.

package jp

import (
	"reflect"

	"github.com/ohler55/ojg/gen"
)

// Nth is a subscript operator that matches the n-th element in an array for a
// JSON path expression.
type Nth int

// Append a fragment string representation of the fragment to the buffer
// then returning the expanded buffer.
func (f Nth) Append(buf []byte, bracket, first bool) []byte {
	buf = append(buf, '[')
	i := int(f)
	if i < 0 {
		buf = append(buf, '-')
		i = -i
	}
	num := [20]byte{}
	cnt := 0
	for ; i != 0; cnt++ {
		num[cnt] = byte(i%10) + '0'
		i /= 10
	}
	if 0 < cnt {
		cnt--
		for ; 0 <= cnt; cnt-- {
			buf = append(buf, num[cnt])
		}
	} else {
		buf = append(buf, '0')
	}
	buf = append(buf, ']')
	return buf
}

func (f Nth) remove(value any) (out any, changed bool) {
	out = value
	i := int(f)
	switch tv := value.(type) {
	case []any:
		if i < 0 {
			i = len(tv) + i
		}
		if 0 <= i && i < len(tv) {
			out = append(tv[:i], tv[i+1:]...)
			changed = true
		}
	case gen.Array:
		if i < 0 {
			i = len(tv) + i
		}
		if 0 <= i && i < len(tv) {
			out = append(tv[:i], tv[i+1:]...)
			changed = true
		}
	case RemovableIndexed:
		size := tv.Size()
		if i < 0 {
			i = size + i
		}
		if 0 <= i && i < size {
			tv.RemoveValueAtIndex(i)
			changed = true
		}
	default:
		if rt := reflect.TypeOf(value); rt != nil {
			if rt.Kind() == reflect.Slice {
				rv := reflect.ValueOf(value)
				cnt := rv.Len()
				if 0 < cnt {
					if i < 0 {
						i = cnt + i
					}
					if 0 <= i && i < cnt {
						nv := reflect.MakeSlice(rt, cnt-1, cnt-1)
						for j := 0; j < i; j++ {
							nv.Index(j).Set(rv.Index(j))
						}
						for j := i + 1; j < cnt; j++ {
							nv.Index(j - 1).Set(rv.Index(j))
						}
						out = nv.Interface()
						changed = true
					}
				}
			}
		}
	}
	return
}

func (f Nth) locate(pp Expr, data any, rest Expr, max int) (locs []Expr) {
	var (
		v   any
		has bool
	)
	i := int(f)
	switch td := data.(type) {
	case []any:
		if i < 0 {
			i = len(td) + i
		}
		if 0 <= i && i < len(td) {
			v = td[i]
			has = true
		}
	case gen.Array:
		if i < 0 {
			i = len(td) + i
		}
		if 0 <= i && i < len(td) {
			v = td[i]
			has = true
		}
	case Indexed:
		if i < 0 {
			i = td.Size() + i
		}
		if 0 <= i && i < td.Size() {
			v = td.ValueAtIndex(i)
			has = true
		}
	default:
		v, has = reflectGetNth(td, i)
	}
	if has {
		locs = locateNthChildHas(pp, Nth(i), v, rest, max)
	}
	return
}

// Walk follows the matching element in a slice or slice like element.
func (f Nth) Walk(rest, path Expr, nodes []any, cb func(path Expr, nodes []any)) {
	var value any
	index := int(f)
	path = append(path, f)
	switch tv := nodes[len(nodes)-1].(type) {
	case []any:
		if index < 0 {
			index += len(tv)
		}
		if index < 0 || len(tv) <= index {
			return
		}
		value = tv[index]
	case gen.Array:
		if index < 0 {
			index += len(tv)
		}
		if index < 0 || len(tv) <= index {
			return
		}
		value = tv[index]
	case Indexed:
		if index < 0 {
			index += tv.Size()
		}
		if index < 0 || tv.Size() <= index {
			return
		}
		value = tv.ValueAtIndex(index)
	default:
		var has bool
		if value, has = reflectGetNth(tv, index); !has {
			return
		}
	}
	if 0 < len(rest) {
		rest[0].Walk(rest[1:], path, append(nodes, value), cb)
	} else {
		cb(path, append(nodes, value))
	}
}
