package types

import ext_types "lightning-poll/types"

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

var (
	neverRepay  = func(votes map[int64]int64, optionID int64) bool { return false }
	alwaysRepay = func(votes map[int64]int64, optionID int64) bool { return true }
	notFound    = func(votes map[int64]int64, optionID int64) bool { return false }

	repayMaximum = func(votes map[int64]int64, optionID int64) bool {
		return getExtreme(votes, optionID, func(challenge, existing int64) bool { return challenge > existing })
	}
	repayMinimum = func(votes map[int64]int64, optionID int64) bool {
		return getExtreme(votes, optionID, func(challenge, existing int64) bool { return challenge < existing })
	}
)

func getExtreme(votes map[int64]int64, optionID int64, beats func(challenge, existing int64) bool) bool {
	var voteCount int64
	var payout []int64
	first := true
	for k, v := range votes {
		if beats(v, voteCount) {
			payout = []int64{k}
			voteCount = v
		}
		// cover the case where 2 options have the same vote
		if v == voteCount {
			payout = append(payout, k)
		}

		if first {
			first = false
			payout = []int64{k}
			voteCount = v
		}
	}

	for _, p := range payout {
		if p == optionID {
			return true
		}
	}

	return false
}

var schemes = map[RepayScheme]ext_types.RepayScheme{
	RepaySchemeUnknown:  neverRepay,
	RepaySchemeMajority: repayMaximum,
	RepaySchemeMinority: repayMinimum,
	RepaySchemeAll:      alwaysRepay,
	RepaySchemeNone:     neverRepay,
}

func (s RepayScheme) GetScheme() ext_types.RepayScheme {
	scheme, ok := schemes[s]
	if !ok {
		return notFound
	}

	return scheme
}
