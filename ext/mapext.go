package ext

import "sort"

func SortedKeys[T any](m map[string]T) []string {
	res := make([]string, 0, len(m))
	for name := range m {
		res = append(res, name)
	}
	sort.Strings(res)
	return res
}
