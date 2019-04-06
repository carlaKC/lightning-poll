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

// repayScheme is a function which returns true if a vote should be refunded.
type repayScheme func(votes map[int64]int64, optionID int64) bool

var (
	noRepayment = func(votes map[int64]int64, optionID int64) bool { return false }
	notFound    = func(votes map[int64]int64, optionID int64) bool { return false }
)

var schemes = map[RepayScheme]repayScheme{
	RepaySchemeUnknown: noRepayment,
}

func (s RepayScheme) GetScheme() repayScheme {
	scheme, ok := schemes[s]
	if !ok {
		return notFound
	}

	return scheme
}
