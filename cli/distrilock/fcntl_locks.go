// This package contains the command line interface executable 'distrilock'.
// To read its command line help, run:
/* $ bin/distrilock --help */
package main

import (
	"os"
	"syscall"
)

func acquireLockDirect(fi *os.File) error {
	fd := fi.Fd()

	var lt syscall.Flock_t
	lt.Type = syscall.F_WRLCK
	lt.Whence = int16(os.SEEK_SET)

	return syscall.FcntlFlock(fd, syscall.F_SETLK, &lt)
}

func peekLock(fi *os.File) (bool, error) {
	fd := fi.Fd()
	var lt syscall.Flock_t
	lt.Type = syscall.F_WRLCK
	lt.Whence = int16(os.SEEK_SET)

	err := syscall.FcntlFlock(fd, syscall.F_GETLK, &lt)
	if err != nil {
		return false, err
	}
	// lock could be write or read, but caller desires to know whether it is lockable or not
	return lt.Type != syscall.F_UNLCK, nil
}

func releaseLock(f *os.File) error {
	fd := f.Fd()

	var lt syscall.Flock_t
	lt.Type = syscall.F_UNLCK
	lt.Whence = int16(os.SEEK_SET)

	return syscall.FcntlFlock(fd, syscall.F_SETLK, &lt)
}
