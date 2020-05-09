package enums

import (
	"errors"
)

const (
	QUEUED status = iota
	CONVERTING
	COMPLETED
	FAILED
)

var statusName = []string{
	"QUEUED",
	"CONVERTING",
	"COMPLETED",
	"FAILED",
}

var statuses = []status{
	QUEUED,
	CONVERTING,
	COMPLETED,
	FAILED,
}

type status int

type Status interface {
	Name()  string
	Value() int
}

func (s status) Name() string {
	return statusName[s]
}

func (s status) Value() int {
	return int(s)
}

func StatusFromEnumValue(enumVal int) (status, error) {
	if enumVal >= len(statuses) {
		return -1, errors.New("unrecognized status")
	}
	return statuses[enumVal], nil
}