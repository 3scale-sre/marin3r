package util

import (
	"strings"

	"github.com/google/go-cmp/cmp"
)

func IgnoreProperty(p string) cmp.Option {
	return cmp.FilterPath(
		func(path cmp.Path) bool {
			if field, ok := path.Last().(cmp.StructField); ok {
				return strings.HasPrefix(field.Name(), p)
			}
			return false
		},
		cmp.Ignore(),
	)
}
