// Package core defines the POSIX-specific primitives to acquire, peek and release locks using fcntl and basic file opening OS capabilities.
package core

/* distrilock - https://github.com/gdm85/distrilock
Copyright (C) 2017 gdm85
This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

import (
	"net"
	"os"
	"regexp"
	"sync"
	"syscall"

	"github.com/gdm85/distrilock/api"
)

const lockExt = ".lck"

var (
	validLockNameRx    = regexp.MustCompile(`^[A-Za-z0-9.\-_]+$`)
	knownResources     = map[string]*os.File{}
	resourceAcquiredBy = map[*os.File]*net.TCPConn{}
	knownResourcesLock sync.RWMutex
)

// ProcessRequest will process the lock command request and return a response.
func ProcessRequest(directory string, client *net.TCPConn, req api.LockRequest) api.LockResponse {
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
		res.Result, res.Reason = acquire(client, req.LockName, directory)
	case api.Release:
		res.Result, res.Reason = release(client, req.LockName, directory)
	case api.Peek:
		res.Result, res.Reason, res.IsLocked = peek(req.LockName, directory)
	case api.Verify:
		res.Result, res.Reason = verifyOwnership(client, req.LockName, directory)
	default:
		res.Result = api.BadRequest
		res.Reason = "unknown command"
	}

	return res
}

// ProcessDisconnect releases sessions and resources associated to the disconnected client.
func ProcessDisconnect(client *net.TCPConn) {
	knownResourcesLock.Lock()

	var filesToDrop []*os.File

	// perform (inefficient) reverse lookups for deletions
	for f, by := range resourceAcquiredBy {
		if by == client {
			// from fcntl(2):
			// > As well as being removed by an explicit F_UNLCK, record locks are
			// > automatically released when the process terminates or if it closes any
			// > file descriptor referring to a file on which locks are held.
			//NOTE: here it is problematic to ignore the close error, because it could mean that file was not closed and thus lock not released
			_ = f.Close()

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

func shortAcquire(client *net.TCPConn, f *os.File, fullLock bool) (api.LockCommandResult, string) {
	// check if lock was acquired by a different client
	by, ok := resourceAcquiredBy[f]
	if fullLock {
		knownResourcesLock.Unlock()
	} else {
		knownResourcesLock.RUnlock()
	}
	if !ok {
		panic("BUG: missing resource acquired by record")
	}
	if by != client {
		return api.Failed, "resource acquired through a different session"
	}

	// lock was already acquired by this session, and it must still be held by us
	// however, note that no re-acquire check is performed here (like in Verify)
	// the client can call Verify to force such check
	return api.Success, "no-op"
}

func acquire(client *net.TCPConn, lockName, directory string) (api.LockCommandResult, string) {
	knownResourcesLock.RLock()

	f, ok := knownResources[lockName]
	if ok {
		return shortAcquire(client, f, false)
	}
	knownResourcesLock.RUnlock()
	knownResourcesLock.Lock()

	// check again, as meanwhile lock could have been created
	f, ok = knownResources[lockName]
	if ok {
		return shortAcquire(client, f, true)
	}

	var err error
	f, err = os.OpenFile(directory+lockName+lockExt, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		knownResourcesLock.Unlock()

		return api.InternalError, err.Error()
	}

	err = acquireLockDirect(f)
	if err != nil {
		_ = f.Close()
		knownResourcesLock.Unlock()

		if e, ok := err.(syscall.Errno); ok {
			if e == syscall.EAGAIN || e == syscall.EACCES { // to be POSIX-compliant, both errors must be checked
				return api.Failed, "resource acquired by different process"
			}
		}

		return api.InternalError, err.Error()
	}

	// writing to file is avoided as it's not necessary

	resourceAcquiredBy[f] = client
	knownResources[lockName] = f
	knownResourcesLock.Unlock()

	// successful lock acquire
	return api.Success, ""
}

func peek(lockName, directory string) (api.LockCommandResult, string, bool) {
	knownResourcesLock.RLock()
	defer knownResourcesLock.RUnlock()

	f, ok := knownResources[lockName]
	if ok {
		// same as the no-op in acquire(), it is assumed here that lock is still held by this process
		return api.Success, "", true
	}

	var err error
	// differently from acquire(), file must exist here
	f, err = os.OpenFile(directory+lockName+lockExt, os.O_RDONLY, 0664)
	if err != nil {
		if e, ok := err.(*os.PathError); ok {
			if e.Err == syscall.ENOENT {
				return api.Success, "", false
			}
		}
		return api.InternalError, err.Error(), false
	}

	isUnlocked, err := isUnlocked(f)
	_ = f.Close()
	if err != nil {
		return api.InternalError, err.Error(), false
	}

	return api.Success, "", !isUnlocked
}

func release(client *net.TCPConn, lockName, directory string) (api.LockCommandResult, string) {
	knownResourcesLock.RLock()

	f, ok := knownResources[lockName]
	if !ok {
		knownResourcesLock.RUnlock()
		return api.Failed, "lock not found"
	}

	// check if lock was acquired by a different client
	by, ok := resourceAcquiredBy[f]
	if !ok {
		panic("BUG: missing resource acquired by record")
	}
	if by != client {
		knownResourcesLock.RUnlock()
		return api.Failed, "resource acquired through a different session"
	}
	knownResourcesLock.RUnlock()
	knownResourcesLock.Lock()

	f, ok = knownResources[lockName]
	if !ok {
		knownResourcesLock.Unlock()
		return api.Failed, "lock not found"
	}

	// check if lock was acquired by a different client
	by, ok = resourceAcquiredBy[f]
	if !ok {
		panic("BUG: missing resource acquired by record")
	}
	if by != client {
		knownResourcesLock.Unlock()
		return api.Failed, "resource acquired through a different session"
	}

	err := releaseLock(f)
	if err != nil {
		knownResourcesLock.Unlock()
		return api.InternalError, err.Error()
	}

	delete(knownResources, lockName)
	delete(resourceAcquiredBy, f)
	_ = f.Close()
	err = os.Remove(directory + lockName + lockExt)

	knownResourcesLock.Unlock()

	if err != nil {
		return api.InternalError, err.Error()
	}

	return api.Success, ""
}

// verifyOwnership verifies that specified client has acquired lock through this node.
func verifyOwnership(client *net.TCPConn, lockName, directory string) (api.LockCommandResult, string) {
	knownResourcesLock.RLock()

	f, ok := knownResources[lockName]
	if !ok {
		knownResourcesLock.RUnlock()
		return api.Failed, "lock not found"
	}

	// check if lock was acquired by a different client
	by, ok := resourceAcquiredBy[f]
	knownResourcesLock.RUnlock()
	if !ok {
		panic("BUG: missing resource acquired by record")
	}
	if by != client {
		return api.Failed, "resource acquired through a different session"
	}
	knownResourcesLock.Lock()
	f, ok = knownResources[lockName]
	if !ok {
		knownResourcesLock.Unlock()
		return api.Failed, "lock not found"
	}

	// check if lock was acquired by a different client
	by, ok = resourceAcquiredBy[f]
	if !ok {
		panic("BUG: missing resource acquired by record")
	}
	if by != client {
		knownResourcesLock.Unlock()
		return api.Failed, "resource acquired through a different session"
	}

	// lock was already acquired by self
	// thus re-acquiring lock must succeed
	err := acquireLockDirect(f)
	knownResourcesLock.Unlock()
	if err != nil {
		if e, ok := err.(syscall.Errno); ok {
			if e == syscall.EAGAIN || e == syscall.EACCES { // to be POSIX-compliant, both errors must be checked
				return api.Failed, "resource acquired by different process"
			}
		}

		return api.InternalError, err.Error()
	}

	// successful lock re-acquisition
	return api.Success, ""
}
