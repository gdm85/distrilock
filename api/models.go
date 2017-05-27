// Package api contains the basic constants, enums and data structures for distrilock API communication.
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
	// invalidCommand is an uninitialised and invalid command.
	invalidCommand LockCommand = iota
	// Peek is the command used to verify current status of a named lock.
	Peek
	// Acquire is the command used to request acquisition of a named lock.
	Acquire
	// Release is the command used to request release of a named lock.
	Release
	// Verify is the command used to verify that a named lock has been acquired by the caller.
	Verify
)

const (
	// invalidResult is an uninitialised and invalid result.
	invalidResult LockCommandResult = iota
	// Failed is returned when the command failed with the specified reason.
	Failed
	// Success is returned when the command succeeded.
	Success
	// BadRequest is returned when the specified parameters are invalid.
	BadRequest
	// InternalError is returned when an unexpected internal error happened while serving the command.
	InternalError
)

// LockRequest is a lock command request descriptor.
type LockRequest struct {
	VersionMajor uint8
	VersionMinor uint8
	Command      LockCommand
	LockName     string
}

// LockResponse is a response to a LockRequest; it always embeds the request's command and lock name.
type LockResponse struct {
	LockRequest
	Result LockCommandResult
	// Reason is the extra human-readable text provided in case of failure, errors, success.
	Reason string
	// IsLocked is specified when peeking lock status.
	IsLocked bool
}

func (lc LockCommand) String() string {
	switch lc {
	case invalidCommand:
		return `INVALID_LOCK_COMMAND`
	case Peek:
		return `Peek`
	case Acquire:
		return `Acquire`
	case Release:
		return `Release`
	case Verify:
		return `Verify`
	}
	return fmt.Sprintf("UNKNOWN_LOCK_COMMAND(%d)", lc)
}

// String returns the human-readable description of the lock command result.
func (lcr LockCommandResult) String() string {
	switch lcr {
	case invalidResult:
		return `INVALID_LOCK_COMMAND_RESULT`
	case Failed:
		return `Failed`
	case Success:
		return `Success`
	case BadRequest:
		return `BadRequest`
	case InternalError:
		return `InternalError`
	}
	return fmt.Sprintf("UNKNOWN_LOCK_COMMAND_RESULT(%d)", lcr)
}
