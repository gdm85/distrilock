// This package contains the command-line-interface executable to serve the distrilock API over websockets.
// To read its command line help, run:
/* $ bin/distrilock-ws --help */
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
	"fmt"
	"net/http"
	"os"
	"time"

	"bitbucket.org/gdm85/go-distrilock/cli"

	"github.com/gorilla/websocket"
)

const defaultKeepAlive = time.Second * 3
const defaultAddress = ":13124"

func main() {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

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

	http.HandleFunc("/distrilock", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// print error continue serving other peers
			fmt.Fprintf(os.Stderr, "upgrade: %v\n", err)
			return
		}

		handleRequests(f.Directory, conn, defaultKeepAlive)
	})

	fmt.Println("distrilock-ws: listening on", f.Address)
	err = http.ListenAndServe(f.Address, nil)
	if err != nil {
		fmt.Println("error listening:", err)
		os.Exit(3)
	}
}
