package pubsub

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)
