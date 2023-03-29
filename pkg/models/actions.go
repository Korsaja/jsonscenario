package models

// ActionFile -
//
//go:generate go run github.com/dmarkham/enumer -type=ActionFile -json -transform=snake -output=action_string.go
type ActionFile int

const (
	Unknown ActionFile = iota
	Create
	Remove
	Rename
	CTime
	Write
	Condition
)
