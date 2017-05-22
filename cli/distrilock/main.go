// This package contains the command line interface executable 'distrilock'.
// To read its command line help, run:
/* $ bin/distrilock --help */
package main

import (
	"fmt"
	"os"
	"bufio"
	"syscall"
)

func main() {
	var err error
	what := os.Args[1]
	if what == "write" {
		err = writeLockSet("book.dat")
	} else if what == "read" {
		err = readLockSet("book.dat")
	} else {
		err = fmt.Errorf("invalid command")
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n') 
}

func writeLockSet(fname string) error {
	fi, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	fd := fi.Fd()

	var lt syscall.Flock_t
	lt.Type = syscall.F_WRLCK
	lt.Whence = int16(os.SEEK_SET)
	saveLock := lt // make a copy

	err = syscall.FcntlFlock(fd, syscall.F_GETLK, &lt)
	if err != nil {
		return err
	}
	if lt.Type == syscall.F_WRLCK {
		return fmt.Errorf("Process %v has a write lock already!", lt.Pid)
	} else if lt.Type == syscall.F_RDLCK {
		return fmt.Errorf("Process %v has a read lock already!", lt.Pid)
	}
	return syscall.FcntlFlock(fd, syscall.F_SETLK, &saveLock)
}

func readLockSet(fname string) error {
	fi, err := os.OpenFile(fname, os.O_RDONLY, 0664)
	if err != nil {
		return err
	}
	fd := fi.Fd()

	var lt syscall.Flock_t
	lt.Type = syscall.F_RDLCK
	lt.Whence = int16(os.SEEK_SET)
	lt.Len = 50
	saveLock := lt // make a copy

	err = syscall.FcntlFlock(fd, syscall.F_GETLK, &lt)
	if err != nil {
		return err
	}
	if lt.Type == syscall.F_WRLCK {
		return fmt.Errorf("Process %v has a write lock already!", lt.Pid)
	}

	return syscall.FcntlFlock(fd, syscall.F_SETLK, &saveLock)
}
