package timef

import (
	"fmt"
	"time"
)

func Date(i uint64) string {
	t := time.Unix(int64(i), 0)
	return fmt.Sprintf("%02d-%02d-%d", t.Day(), t.Month(), t.Year())
}

func File() string {
	now := time.Now()
	return fmt.Sprintf("%d%02d%02d_%02d%02d%02d.%09d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond())
}