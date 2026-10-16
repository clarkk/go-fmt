package csv

import (
	"fmt"
	"cmp"
	"slices"
)

type (
	count_sep struct {
		count map[rune]int
	}
	
	count_lines_sep struct {
		count map[rune][]int
	}
)

func (c *count_sep) count_sep(sep rune, count int){
	c.count[sep] = count
}

func (c *count_sep) get_sep() (rune, error){
	return get_sep(c.count)
}

func (c *count_lines_sep) count_sep(sep rune, count int){
	c.count[sep] = append(c.count[sep], count)
}

func (c *count_lines_sep) get_sep() (rune, error){
	candidates := map[rune]int{}
	for sep, counts := range c.count {
		if len(counts) == 0 {
			continue
		}
		//	Check if all lines have the same occurrence of the separator
		max := slices.Max(counts)
		if max == 0 {
			continue
		}
		if max == slices.Min(counts) {
			candidates[sep] = max
		}
	}
	return get_sep(candidates)
}

func get_sep(count map[rune]int) (rune, error){
	length := len(count)
	if length == 0 {
		return 0, fmt.Errorf("Unable to find separator candidates")
	}
	keys := make([]rune, length)
	i := 0
	for sep := range count {
		keys[i] = sep
		i++
	}
	slices.SortFunc(keys, func(a, b rune) int {
		return cmp.Compare(count[b], count[a])
	})
	return keys[0], nil
}