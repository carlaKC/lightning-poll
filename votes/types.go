package votes

type Vote struct {
	ID       int64
	OptionID int64
	Preimage []byte
}
