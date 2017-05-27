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

func handleRequests(directory string, conn *net.TCPConn, keepAlivePeriod time.Duration) {
	// setup keep-alive
	err := conn.SetKeepAlive(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not set keep alive: %v\n", err)
		return
	}
	err = conn.SetKeepAlivePeriod(keepAlivePeriod)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not set keep alive period: %v\n", err)
		return
	}
	//fmt.Println("a client connected")

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

	conn.Close()
	//fmt.Println("a client disconnected")

	processDisconnect(conn)
}
