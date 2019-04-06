package polls

type Poll struct {
	ID       int64
	Question string
	Options  []*Option
}

type Option struct {
	ID    int64
	Value string
}
