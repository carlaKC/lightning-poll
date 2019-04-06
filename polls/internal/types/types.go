package types

type PollStatus int

var (
	PollStatusUnknown  PollStatus = 0
	PollStatusCreated  PollStatus = 1
	PollStatusClosed   PollStatus = 2
	PollStatusReleased PollStatus = 3
	PollStatusPaidOut  PollStatus = 4
	pollStatusSentinel PollStatus = 5
)

func (s PollStatus) Valid() bool {
	return s > PollStatusUnknown && s < pollStatusSentinel
}

type RepayScheme int

var (
	RepaySchemeUnknown  RepayScheme = 0
	RepaySchemeMajority RepayScheme = 1
	RepaySchemeMinority RepayScheme = 2
	RepaySchemeAll      RepayScheme = 3
	RepaySchemeNone     RepayScheme = 4
	repaySchemeSentinel RepayScheme = 5
)

func (s RepayScheme) Valid() bool {
	return s > RepaySchemeUnknown && s < repaySchemeSentinel
}
