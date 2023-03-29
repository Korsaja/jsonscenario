package ffuncs

import "github.com/pkg/errors"

const TimeFormat = "2006-01-02 15:04:05"

// Result -
//
//go:generate go run github.com/dmarkham/enumer -type=Result -json -transform=snake -output=const_string.go
type Result int

const (
	Failed Result = iota
	Success
)

var ErrFalseCondition = errors.New("false condition")
