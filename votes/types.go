package votes

type Vote struct {
	ID       int64
	OptionID int64
	Amount   int64
	Hash     string
	Preimage []byte
}
