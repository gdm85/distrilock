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

	flags, err := flags.Parse(os.Args, defaultAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(5)
	}

	http.HandleFunc("/distrilock", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// print error continue serving other peers
			fmt.Fprintf(os.Stderr, "upgrade: %v\n", err)
			return
		}

		handleRequests(flags.Directory, conn, defaultKeepAlive)
	})

	fmt.Println("distrilock-ws: listening on", flags.Address)
	err = http.ListenAndServe(flags.Address, nil)
	if err != nil {
		fmt.Println("error listening:", err)
		os.Exit(1)
	}
}
