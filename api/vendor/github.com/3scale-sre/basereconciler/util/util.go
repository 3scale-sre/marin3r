package util

// Returns a pointer of any type
func Pointer[T any](t T) *T {
	return &t
}

// MergeMaps merges a list maps into the first one. B overrides A if keys collide.
func MergeMaps(base map[string]string, merges ...map[string]string) map[string]string {
	for _, m := range merges {
		for key, value := range m {
			base[key] = value
		}
	}
	return base
}

func ConvertStringSlice[T1 ~string, T2 ~string](collection []T1) []T2 {
	out := make([]T2, 0, len(collection))
	for _, item := range collection {
		out = append(out, T2(item))
	}
	return out
}

func ContainsBy[T any](collection []T, predicate func(item T) bool) bool {
	for _, item := range collection {
		if predicate(item) {
			return true
		}
	}
	return false
}
