package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

const defaultKeepAlive = time.Second * 3

func main() {
	flags, err := mustParseFlags(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(5)
	}

	// Listen for incoming connections.
	l, err := net.ListenTCP("tcp", flags.a)
	if err != nil {
		fmt.Println("distrilock: error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("distrilock: listening on", flags.a)
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
