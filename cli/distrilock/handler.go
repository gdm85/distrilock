package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"bitbucket.org/gdm85/go-distrilock/api"
)

// Handles incoming requests.
func handleRequest(directory string, conn *net.TCPConn) {
	// setup keep-alive
	err := conn.SetKeepAlive(true)
	if err != nil {
		panic(err.Error())
	}
	err = conn.SetKeepAlivePeriod(time.Second * 3)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("a client connected")

	d := gob.NewDecoder(conn)
	e := gob.NewEncoder(conn)
	for {
		var req api.LockRequest
		err = d.Decode(&req)
		if err != nil {
			if err == io.EOF {
				// other end interrupted connection
				break
			}
			fmt.Fprintln(os.Stderr, "error reading:", err.Error())
			continue
		}
		//fmt.Println("received request:", req)

		if req.VersionMajor > api.VersionMajor {
			fmt.Fprintln(os.Stderr, "skipping request with superior major version")
			continue
		}

		res := processRequest(directory, conn, req)

		// Send a response back to person contacting us.
		err = e.Encode(&res)
		if err != nil {
			if err == io.EOF {
				// other end interrupted connection
				break
			}
			fmt.Fprintln(os.Stderr, "Error writing:", err.Error())
			continue
		}
	}
	// Close the connection when you're done with it.
	conn.Close()
	fmt.Println("a client disconnected")

	processDisconnect(conn)
}
