package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"sync"

	"bitbucket.org/gdm85/go-distrilock/api"
)

var (
	knownResources     = map[string]*os.File{}
	resourceAcquiredBy = map[*os.File]*net.TCPConn{}
	knownResourcesLock sync.Mutex
	validLockNameRx    = regexp.MustCompile(`^[A-Za-z0-9.\-]+$`)
)

func processRequest(client *net.TCPConn, req api.LockRequest) api.LockResponse {
	var res api.LockResponse
	res.LockRequest = req
	// override with own version
	res.VersionMajor, res.VersionMinor = api.VersionMajor, api.VersionMinor

	// validate lock name
	if !validLockNameRx.MatchString(req.LockName) {
		res.Result = api.BadRequest
		res.Reason = "invalid lock name"
		return res
	}

	switch res.Command {
	case api.Acquire:
		res.Result, res.Reason = acquire(client, req.LockName)
	case api.Release:
		res.Result, res.Reason = release(client, req.LockName)
	case api.Peek:
		res.Result, res.Reason, res.IsLocked = peek(req.LockName)
	default:
		res.Result = api.BadRequest
		res.Reason = "unknown command"
	}

	return res
}

func peek(lockName string) (api.LockCommandResult, string, bool) {
	knownResourcesLock.Lock()
	defer knownResourcesLock.Unlock()

	f, ok := knownResources[lockName]
	if !ok {
		var err error
		// differently from acquire(), file must exist here
		f, err = os.OpenFile(lockName, os.O_RDWR, 0664)
		if err != nil {
			return api.InternalError, err.Error(), false
		}

		isWriteLocked, err := peekLock(f)
		f.Close()
		if err != nil {
			return api.InternalError, err.Error(), false
		}

		// successful lock acquire
		return api.Success, "", isWriteLocked
	}

	// file is not closed here
	isWriteLocked, err := peekLock(f)
	if err != nil {
		return api.InternalError, err.Error(), false
	}

	// successful lock acquire
	return api.Success, "", isWriteLocked
}

func release(client *net.TCPConn, lockName string) (api.LockCommandResult, string) {
	knownResourcesLock.Lock()
	defer knownResourcesLock.Unlock()

	f, ok := knownResources[lockName]
	if !ok {
		return api.Failed, "lock not found"
	}
	err := releaseLock(f)
	if err != nil {
		return api.InternalError, err.Error()
	}

	// check if lock was acquired by a different client
	by, ok := resourceAcquiredBy[f]
	if !ok {
		panic("BUG: missing resource acquired by record")
	}
	if by != client {
		return api.Denied, "resource acquired in a different session"
	}

	delete(resourceAcquiredBy, f)
	f.Close()
	return api.Success, ""
}

func acquire(client *net.TCPConn, lockName string) (api.LockCommandResult, string) {
	knownResourcesLock.Lock()
	defer knownResourcesLock.Unlock()

	f, ok := knownResources[lockName]
	if !ok {
		var err error
		f, err = os.OpenFile(lockName, os.O_RDWR|os.O_CREATE, 0664)
		if err != nil {
			return api.InternalError, err.Error()
		}

		err = acquireLockDirect(f)
		if err != nil {
			f.Close()
			return api.InternalError, err.Error()
		}

		_, err = f.Write([]byte(fmt.Sprintf("%p", client)))
		if err != nil {
			f.Close()
			return api.InternalError, err.Error()
		}

		resourceAcquiredBy[f] = client
		knownResources[lockName] = f

		// successful lock acquire
		return api.Success, ""
	}

	// check if lock was acquired by a different client
	by, ok := resourceAcquiredBy[f]
	if !ok {
		panic("BUG: missing resource acquired by record")
	}
	if by != client {
		return api.Denied, "resource acquired in a different session"
	}

	// already acquired by self
	//TODO: this is a no-operation, should lock be acquired again with fcntl?
	//		and what if the re-acquisition fails? that would perhaps qualify
	//		as a different lock command?
	return api.Success, "no-op"
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
