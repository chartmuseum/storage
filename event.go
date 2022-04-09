package storage

// EventType represents the type of event kind
type EventType uint8

const (
	EventPutObject EventType = iota + 1
	EventDeleteObject
)
