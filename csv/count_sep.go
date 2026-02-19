package csv

import (
	"fmt"
	"cmp"
	"slices"
)

type count_sep struct {
	count		map[rune]int
	count_lines	map[rune][]int
}

func newCount_sep() *count_sep {
	c := &count_sep{
		count:			map[rune]int{},
		count_lines:	map[rune][]int{},
	}
	for _, sep := range separators {
		c.count_lines[sep] = []int{}
	}
	return c
}

func (c *count_sep) count_sep(sep rune, count int){
	c.count[sep] = count
}

func (c *count_sep) count_lines_sep(sep rune, count int){
	c.count_lines[sep] = append(c.count_lines[sep], count)
}

func (c *count_sep) get_sep() (rune, error){
	length := len(c.count)
	if length == 0 {
		return 0, fmt.Errorf("Unable to find separator candidates")
	}
	
	keys := make([]rune, length)
	i := 0
	for sep := range c.count {
		keys[i] = sep
		i++
	}
	slices.SortFunc(keys, func(a, b rune) int {
		return cmp.Compare(c.count[b], c.count[a])
	})
	return keys[0], nil
}

func (c *count_sep) get_lines_sep() (rune, error){
	for sep, count := range c.count_lines {
		max := slices.Max(count)
		if max == 0 {
			continue
		}
		if max == slices.Min(count) {
			c.count_sep(sep, max)
		}
	}
	return c.get_sep()
}