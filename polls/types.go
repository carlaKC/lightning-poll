package polls

import (
	"time"

	"github.com/carlaKC/lightning-poll/types"
)

type Poll struct {
	ID       int64
	Question string
	Options  []*Option
	Cost     int64
	ClosesAt time.Time
	Strategy types.RepayDetails
}

type Option struct {
	ID    int64
	Value string
}
