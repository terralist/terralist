package batch

type Batch interface {
	Add(input Input) Batch
	Commit() (Output, error)
}
