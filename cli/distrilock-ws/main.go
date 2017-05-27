// distrilock-ws is the command-line-interface executable to run the distrilock API over websockets.
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	//"bitbucket.org/gdm85/go-distrilock/api/client/ws"

	"github.com/gorilla/websocket"
)

const defaultKeepAlive = time.Second * 3
const address = `:8080`
const directory = `./`

func main() {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	http.HandleFunc("/distrilock", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// print error continue serving other peers
			fmt.Fprintf(os.Stderr, "upgrade: %v\n", err)
			return
		}

		handleRequests(directory, conn, defaultKeepAlive)
	})

	fmt.Printf("distrilock-ws: listening at address %s\n", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println("error listening:", err)
		os.Exit(1)
	}
}
