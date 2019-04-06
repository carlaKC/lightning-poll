package types

type VoteStatus int

var (
	VoteStatusUnknown  VoteStatus = 0
	VoteStatusCreated  VoteStatus = 1
	VoteStatusExpired  VoteStatus = 2
	VoteStatusPaid     VoteStatus = 3
	VoteStatusReturned VoteStatus = 4
	VoteStatusSettled  VoteStatus = 5
	voteStatusSentinel VoteStatus = 6
)

func (s VoteStatus) Valid() bool {
	return s > VoteStatusUnknown && s < voteStatusSentinel
}
