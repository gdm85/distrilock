package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"sync"
)

type LockCommand uint8
type LockCommandResult uint8

const (
	VersionMajor = 0
	VersionMinor = 1
)

const (
	Peek LockCommand = iota
	Acquire
	Release
)

const (
	Failed LockCommandResult = iota
	Success
	BadRequest
	InternalError
	Denied
	TooBusy
)

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

var (
	knownResources     = map[string]*os.File{}
	resourceAcquiredBy = map[*os.File]*net.TCPConn{}
	knownResourcesLock sync.Mutex
	validLockNameRx    = regexp.MustCompile(`^[A-Za-z0-9.\-]+$`)
)

func processRequest(client *net.TCPConn, req LockRequest) LockResponse {
	var res LockResponse
	res.LockRequest = req
	// override with own version
	res.VersionMajor, res.VersionMinor = VersionMajor, VersionMinor

	// validate lock name
	if !validLockNameRx.MatchString(req.LockName) {
		res.Result = BadRequest
		res.Reason = "invalid lock name"
		return res
	}

	switch res.Command {
	case Acquire:
		res.Result, res.Reason = acquire(client, req.LockName)
	case Release:
		res.Result, res.Reason = release(client, req.LockName)
	case Peek:
		res.Result, res.Reason, res.IsLocked = peek(req.LockName)
	default:
		res.Result = BadRequest
		res.Reason = "unknown command"
	}

	return res
}

func peek(lockName string) (LockCommandResult, string, bool) {
	knownResourcesLock.Lock()
	defer knownResourcesLock.Unlock()

	f, ok := knownResources[lockName]
	if !ok {
		var err error
		// differently from acquire(), file must exist here
		f, err = os.OpenFile(lockName, os.O_RDWR, 0664)
		if err != nil {
			return InternalError, err.Error(), false
		}

		isWriteLocked, err := peekLock(f)
		f.Close()
		if err != nil {
			return InternalError, err.Error(), false
		}

		// successful lock acquire
		return Success, "", isWriteLocked
	}

	// file is not closed here
	isWriteLocked, err := peekLock(f)
	if err != nil {
		return InternalError, err.Error(), false
	}

	// successful lock acquire
	return Success, "", isWriteLocked
}

func release(client *net.TCPConn, lockName string) (LockCommandResult, string) {
	knownResourcesLock.Lock()
	defer knownResourcesLock.Unlock()

	f, ok := knownResources[lockName]
	if !ok {
		return Failed, "lock not found"
	}
	err := releaseLock(f)
	if err != nil {
		return InternalError, err.Error()
	}

	// check if lock was acquired by a different client
	by, ok := resourceAcquiredBy[f]
	if !ok {
		panic("BUG: missing resource acquired by record")
	}
	if by != client {
		return Denied, "resource acquired in a different session"
	}

	delete(resourceAcquiredBy, f)
	f.Close()
	return Success, ""
}

func acquire(client *net.TCPConn, lockName string) (LockCommandResult, string) {
	knownResourcesLock.Lock()
	defer knownResourcesLock.Unlock()

	f, ok := knownResources[lockName]
	if !ok {
		var err error
		f, err = os.OpenFile(lockName, os.O_RDWR|os.O_CREATE, 0664)
		if err != nil {
			return InternalError, err.Error()
		}

		err = acquireLockDirect(f)
		if err != nil {
			f.Close()
			return InternalError, err.Error()
		}

		_, err = f.Write([]byte(fmt.Sprintf("%p", client)))
		if err != nil {
			f.Close()
			return InternalError, err.Error()
		}

		resourceAcquiredBy[f] = client
		knownResources[lockName] = f

		// successful lock acquire
		return Success, ""
	}

	// check if lock was acquired by a different client
	by, ok := resourceAcquiredBy[f]
	if !ok {
		panic("BUG: missing resource acquired by record")
	}
	if by != client {
		return Denied, "resource acquired in a different session"
	}

	// already acquired by self
	//TODO: this is a no-operation, should acquire again with fcntl?
	return Success, "no-op"
}

func processDisconnect(client *net.TCPConn) {
	knownResourcesLock.Lock()

	var filesToDrop []*os.File

	// perform (inefficient) reverse lookups for deletions
	for f, by := range resourceAcquiredBy {
		if by == client {
			f.Close()
			filesToDrop = append(filesToDrop, f)
			delete(resourceAcquiredBy, f)
		}
	}
	for _, droppedF := range filesToDrop {
		for name, f := range knownResources {
			if f == droppedF {
				delete(knownResources, name)
				break
			}
		}
	}

	knownResourcesLock.Unlock()
}
