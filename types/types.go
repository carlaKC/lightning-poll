package types

// RepayScheme is a function which returns true if a vote should be refunded.
type RepayScheme func(votes map[int64]int64, optionID int64) bool
