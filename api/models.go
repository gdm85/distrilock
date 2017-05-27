package api

import (
	"fmt"
)

// LockCommand is a lock command identifier.
type LockCommand uint8

// LockCommandResult is the result of a lock command.
type LockCommandResult uint8

const (
	// VersionMajor is the major version of the distrilock protocol
	VersionMajor = 0
	// VersionMinor is the minor version of the distrilock protocol
	VersionMinor = 1
)

const (
	// Peek is the command used to verify current status of a named lock.
	Peek LockCommand = iota
	// Acquire is the command used to request acquisition of a named lock.
	Acquire
	// Release is the command used to request release of a named lock.
	Release
)

const (
	Failed LockCommandResult = iota
	Success
	BadRequest
	InternalError
	TooBusy
)

func (lcr LockCommandResult) String() string {
	switch lcr {
	case Failed:
		return `Failed`
	case Success:
		return `Success`
	case BadRequest:
		return `BadRequest`
	case InternalError:
		return `InternalError`
	case TooBusy:
		return `TooBusy`
	}
	return fmt.Sprintf("UNKNOWN(%d)", lcr)
}

type LockRequest struct {
	VersionMajor uint8
	VersionMinor uint8
	Command      LockCommand
	LockName     string
}

type LockResponse struct {
	LockRequest
	Result   LockCommandResult
	Reason   string
	IsLocked bool // peeked information
}
