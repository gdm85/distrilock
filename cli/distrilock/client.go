package main



import (
    "fmt"
    "net"
    "os"
    "io"
    "time"
    "encoding/gob"
)

func client() {
	a, err := net.ResolveTCPAddr("tcp", CONN_HOST+":"+CONN_PORT)
    if err != nil {
        fmt.Println("Error resolving:", err.Error())
        os.Exit(1)
    }

    conn, err := net.DialTCP("tcp", nil, a)
    if err != nil {
        fmt.Println("Dial failed:", err.Error())
        os.Exit(1)
    }
    defer conn.Close()
    
	// setup keep-alive
	err = conn.SetKeepAlive(true)
	if err != nil {
		panic(err.Error())
	}
	err = conn.SetKeepAlivePeriod(time.Second*3)
	if err != nil {
		panic(err.Error())
	}
	d := gob.NewDecoder(conn)
	e := gob.NewEncoder(conn)
	
	var req LockRequest
	req.VersionMajor, req.VersionMinor = VersionMajor, VersionMinor
	req.Command = Acquire
	req.LockName = "book"
	
	// Send a response back to person contacting us.
	  err = e.Encode(&req)
	  if err != nil {
		if err == io.EOF {
			// other end interrupted connection
			return
		}
		fmt.Println("Error writing:", err.Error())
		return
	  }
	  
	  // wait for a response
    var res LockResponse
	err = d.Decode(&res)
	  if err != nil {
		if err == io.EOF {
			// other end interrupted connection
			return
		}
		fmt.Println("Error reading:", err.Error())
		return
	  }
	  fmt.Println("received:", req)
}
