package internal

// Package internal provides the interface for the Cadence worker and workflow
// IChannel is an interface that defines the methods for sending and receiving
type IChannel interface {
	Receive(ctx Context, valuePtr interface{}) (ok bool)

	ReceiveAsync(valuePtr interface{}) (ok bool)
	ReceiveAsyncWithMoreFlag(valuePtr interface{}) (ok bool, more bool)
	Send(ctx Context, v interface{})

	SendAsync(v interface{}) (ok bool)
	Close()
}

// Context defines the methods that a workflow.Context should implement.
type Context interface {
	Value(key interface{}) interface{}
}
