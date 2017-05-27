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

func main() {
	flags, err := flags.Parse(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(5)
	}

	// validate address
	addr, err := net.ResolveTCPAddr("tcp", flags.Address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(5)
	}

	// Listen for incoming connections.
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println("distrilock: error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("distrilock: listening on", addr)
	for {
		// Listen for an incoming connection.
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error accepting: ", err.Error())
			continue
		}
		// Handle connections in a new goroutine.
		go handleRequests(flags.Directory, conn, defaultKeepAlive)
	}
}
