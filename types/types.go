package types

// RepaySchemeFunc is a function which returns true if a vote should be refunded.
type RepaySchemeFunc func(votes map[int64]int64, optionID int64) bool

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
		return !repayMaximum(votes, optionID)
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

var schemes = map[RepayScheme]RepaySchemeFunc{
	RepaySchemeUnknown:  neverRepay,
	RepaySchemeMajority: repayMaximum,
	RepaySchemeMinority: repayMinimum,
	RepaySchemeAll:      alwaysRepay,
	RepaySchemeNone:     neverRepay,
}

func (s RepayScheme) GetScheme() RepaySchemeFunc {
	scheme, ok := schemes[s]
	if !ok {
		return notFound
	}

	return scheme
}

func (s RepayScheme)GetDetails()RepayDetails{
	return allSchemes[s]
}

type RepayDetails struct {
	Name        string
	Description string
}

var allSchemes = map[RepayScheme]RepayDetails{
	RepaySchemeMajority:{
		Name:        "Repay Majority",
		Description: "Repay voters who vote for the most popular option",
	},
	RepaySchemeMinority:{
		Name:        "Repay Non-Majority",
		Description: "Repay voters who do not vote for the most popular option",
	},
	RepaySchemeAll:{
		Name:        "Repay Everybody",
		Description: "Repay all voters",
	},
	RepaySchemeNone:{
		Name:        "Repay Nobody",
		Description: "Repay no voters (bc broke or an asshole)",
	},
}

func GetRepaySchemes() map[RepayScheme]RepayDetails {
	return allSchemes
}

