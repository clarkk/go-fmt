package csv

import (
	"fmt"
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
	return determinate_sep(c.count)
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
	return determinate_sep(candidates)
}

func determinate_sep(count map[rune]int) (rune, error){
	var (
		sep			rune
		max_count	int
	)
	for _, candidate := range separators {
		if n := count[candidate]; n > max_count {
			sep = candidate
			max_count = n
		}
	}
	if max_count == 0 {
		return 0, fmt.Errorf("Unable to find separator candidates")
	}
	return sep, nil
}