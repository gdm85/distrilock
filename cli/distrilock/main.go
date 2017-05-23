package main

import (
    "fmt"
    "net"
    "io"
    "time"
    "os"
)

const (
    CONN_HOST = "localhost"
    CONN_PORT = "8071"
)

func main() {
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

// Handles incoming requests.
func handleRequest(conn *net.TCPConn) {
	// setup keep-alive
	err := conn.SetKeepAlive(true)
	if err != nil {
		panic(err.Error())
	}
	err = conn.SetKeepAlivePeriod(time.Second*3)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("a client connected")
  // Make a buffer to hold incoming data.
  buf := make([]byte, 1024)
  for {
	// Read the incoming connection into the buffer.
	  reqLen, err := conn.Read(buf)
	  if err != nil {
		if err == io.EOF {
			// other end interrupted connection
			break
		}
		fmt.Println("Error reading:", err.Error())
		continue
	  }
	  fmt.Println("received:", string(buf[:reqLen]))
	  // Send a response back to person contacting us.
	  conn.Write([]byte("Message received."))
	}
  // Close the connection when you're done with it.
  conn.Close()
  fmt.Println("a client disconnected")
}
