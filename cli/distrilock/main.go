package main

import (
	"fmt"
	"net"
	"os"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "8071"
)

func main() {
	what := os.Args[1]
	if what == "server" {
		server()
	} else if what == "client" {
		client()
	} else {
		panic("invalid choice")
	}
}

func server() {
	a, err := net.ResolveTCPAddr("tcp", CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error resolving:", err.Error())
		os.Exit(1)
	}

	// Listen for incoming connections.
	l, err := net.ListenTCP("tcp", a)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
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
