package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	flags, err := mustParseFlags(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(5)
	}
	
	if flags.Client {
		client(flags.a)
	} else {
		server(flags.a)
	}
}

func server(a *net.TCPAddr) {
	// Listen for incoming connections.
	l, err := net.ListenTCP("tcp", a)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on", a)
	for {
		// Listen for an incoming connection.
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}
