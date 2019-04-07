package polls

import (
	"time"
)

type Poll struct {
	ID       int64
	Question string
	Options  []*Option
	Cost     int64
	ClosesAt time.Time
}

type Option struct {
	ID    int64
	Value string
}
