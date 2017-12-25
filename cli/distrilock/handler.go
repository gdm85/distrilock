package main

/* distrilock - https://github.com/gdm85/distrilock
Copyright (C) 2017 gdm85
This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/gdm85/distrilock/api"
	"github.com/gdm85/distrilock/api/core"
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

		res := core.ProcessRequest(directory, conn, req)

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

	_ = conn.Close()
	//fmt.Println("a client disconnected")

	core.ProcessDisconnect(conn)
}
