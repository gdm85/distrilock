// This package contains the command line interface executable 'distrilock'.
// To read its command line help, run:
/* $ bin/distrilock --help */
package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"bitbucket.org/gdm85/go-distrilock/cli"
)

const defaultKeepAlive = time.Second * 3
const defaultAddress = ":13123"

func main() {
	f, err := flags.Parse(os.Args, defaultAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// print information about maximum number of files
	noFile, err := flags.GetNumberOfFilesLimit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(2)
	}
	if noFile <= 1024 {
		// print maximum number of files only when it's a bit low
		fmt.Println("distrilock: maximum number of files allowed is", noFile)
	}

	// validate address
	addr, err := net.ResolveTCPAddr("tcp", f.Address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(3)
	}

	// Listen for incoming connections.
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println("distrilock: error listening:", err.Error())
		os.Exit(4)
	}
	fmt.Println("distrilock: listening on", addr)
	for {
		// Listen for an incoming connection.
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error accepting: ", err.Error())
			continue
		}
		// Handle connections in a new goroutine.
		go handleRequests(f.Directory, conn, defaultKeepAlive)
	}
}
