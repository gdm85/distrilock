package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"bitbucket.org/gdm85/go-distrilock/api"
	"bitbucket.org/gdm85/go-distrilock/api/core"

	"github.com/gorilla/websocket"
)

func handleRequests(directory string, wsconn *websocket.Conn, keepAlivePeriod time.Duration) {
	var conn *net.TCPConn
	{
		var ok bool
		c := wsconn.UnderlyingConn()
		conn, ok = c.(*net.TCPConn)
		if !ok {
			fmt.Fprintf(os.Stderr, "found connection type %T, but %T expected\n", c, conn)
			return
		}

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
	}

Loop:
	for {
		messageType, r, err := wsconn.NextReader()
		if err != nil {
			_, ok := err.(*websocket.CloseError)
			if ok {
				// other end interrupted connection
				break
			}
			fmt.Fprintf(os.Stderr, "error getting next reader: %v\n", err)
			wsconn.Close()
			return
		}

		var req api.LockRequest

		switch messageType {
		case websocket.BinaryMessage:
			d := gob.NewDecoder(r)
			err = d.Decode(&req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error decoding binary message: %v\n", err)
				continue
			}
		case websocket.TextMessage:
			d := json.NewDecoder(r)
			err = d.Decode(&req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error decoding JSON message: %v\n", err)
				continue
			}
		default:
			panic("BUG: only BinaryMessage or TextMessage types expected")
		}

		if req.VersionMajor > api.VersionMajor {
			fmt.Fprintln(os.Stderr, "skipping request with superior major version")
			continue
		}

		res := core.ProcessRequest(directory, conn, req)

		// reply with same type as last message
		w, err := wsconn.NextWriter(messageType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting next writer: %v\n", err)
			continue
		}

		// send the response
		switch messageType {
		case websocket.BinaryMessage:
			e := gob.NewEncoder(w)

			err = e.Encode(&res)
			w.Close()
			if err != nil {
				_, ok := err.(*websocket.CloseError)
				if ok {
					// other end interrupted connection
					break Loop
				}
				fmt.Fprintln(os.Stderr, "error writing binary response:", err.Error())
				continue
			}
		case websocket.TextMessage:
			e := json.NewEncoder(w)

			err = e.Encode(&res)
			w.Close()
			if err != nil {
				if err == io.EOF {
					// other end interrupted connection
					break Loop
				}
				fmt.Fprintln(os.Stderr, "error writing JSON response:", err.Error())
				continue
			}
		}
	}

	// Close the connection when you're done with it.
	wsconn.Close()
	//fmt.Println("a client disconnected")

	core.ProcessDisconnect(conn)
}
