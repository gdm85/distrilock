// This package contains the command line interface executable 'distrilock'.
// To read its command line help, run:
/* $ bin/distrilock --help */
package main

import (
	"fmt"
	"os"
	"bufio"
	"syscall"
	"time"
)

func main() {
	var err error
	var f *os.File
	
	what := os.Args[1]
	if what == "write" {
		err = writeLockSet("book.dat")
	} else if what == "read" {
		err = readLockSet("book.dat")
	} else if what == "acquire" {
		f, err = acquireLock("book.dat")
	} else {
		err = fmt.Errorf("invalid command")
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	if what == "acquire" {
		fmt.Println("waiting 3 seconds before release")
		time.Sleep(3 * time.Second)
		err = releaseLock(f)
	}
	
	
	fmt.Println("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n') 
}

func acquireLock(fname string) (*os.File, error) {
	fi, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return nil, err
	}
	fd := fi.Fd()

	var lt syscall.Flock_t
	lt.Type = syscall.F_WRLCK
	lt.Whence = int16(os.SEEK_SET)
	saveLock := lt // make a copy

	return fi, syscall.FcntlFlock(fd, syscall.F_SETLK, &saveLock)
}

func releaseLock(f *os.File) error {
	fd := f.Fd()

	var lt syscall.Flock_t
	lt.Type = syscall.F_UNLCK
	lt.Whence = int16(os.SEEK_SET)

	return syscall.FcntlFlock(fd, syscall.F_SETLK, &lt)
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
