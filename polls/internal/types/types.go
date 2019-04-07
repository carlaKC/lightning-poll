package types

type PollStatus int

var (
	PollStatusUnknown   PollStatus = 0
	PollStatusCreated   PollStatus = 1
	PollStatusClosed    PollStatus = 2
	PollStatusReleased  PollStatus = 3
	PollStatusPayingOut PollStatus = 4
	PollStatusPaidOut   PollStatus = 5
	pollStatusSentinel  PollStatus = 6
)

func (s PollStatus) Valid() bool {
	return s > PollStatusUnknown && s < pollStatusSentinel
}
