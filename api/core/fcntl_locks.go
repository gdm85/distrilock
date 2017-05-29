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

func isUnlocked(fi *os.File) (bool, error) {
	fd := fi.Fd()
	var lt syscall.Flock_t
	lt.Type = syscall.F_WRLCK
	lt.Whence = int16(os.SEEK_SET)

	err := syscall.FcntlFlock(fd, syscall.F_GETLK, &lt)
	if err != nil {
		return false, err
	}

	// lock could be write or read, but caller desires only to know whether it is unlocked or not
	return lt.Type == syscall.F_UNLCK, nil
}

func releaseLock(f *os.File) error {
	fd := f.Fd()

	var lt syscall.Flock_t
	lt.Type = syscall.F_UNLCK
	lt.Whence = int16(os.SEEK_SET)

	return syscall.FcntlFlock(fd, syscall.F_SETLK, &lt)
}
