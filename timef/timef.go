package timef

import (
	"fmt"
	"time"
)

func File() string {
	now := time.Now()
	return fmt.Sprintf("%d%02d%02d_%02d%02d%02d.%09d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond())
}