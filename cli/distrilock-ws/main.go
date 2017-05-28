// This package contains the command-line-interface executable to serve the distrilock API over websockets.
// To read its command line help, run:
/* $ bin/distrilock --help */
package main

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
