package core

const (
	_           = iota
	EventDelete = iota
	EventPut
)

type EventType byte

type Event struct {
	Sequence  uint64
	Key       string
	Value     string
	EventType EventType
}

type TransactionLogger interface {
	WritePut(key, value string)
	WriteDelete(key string)
	Err() <-chan error
	Run()
	ReadEvents() (<-chan Event, <-chan error)
}
