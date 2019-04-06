package types

type VoteStatus int

var (
	VoteStatusUnknown  VoteStatus = 0
	VoteStatusCreated  VoteStatus = 1
	VoteStatusExpired  VoteStatus = 2
	VoteStatusPaid     VoteStatus = 3
	VoteStatusReturned VoteStatus = 4
	voteStatusSentinel VoteStatus = 5
)

func (s VoteStatus) Valid() bool {
	return s > VoteStatusUnknown && s < voteStatusSentinel
}
