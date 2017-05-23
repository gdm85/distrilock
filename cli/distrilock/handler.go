package main


import (
    "fmt"
    "net"
    "io"
    "time"
    "encoding/gob"
)

type LockCommand uint8
type LockCommandResult uint8

const (
	VersionMajor = 0
	VersionMinor = 1
	
	Peek LockCommand = iota
	Acquire
	Release	
	
	Failed LockCommandResult = iota
	Success
	BadRequest
	InternalError
	Denied
	TooBusy
)


type LockRequest struct {
	VersionMajor uint8
	VersionMinor uint8
	Command LockCommand
	LockName string
}

type LockResponse struct {
	LockRequest
	Result LockCommandResult
	Reason string
	//TODO: add peeked info
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
	
	d := gob.NewDecoder(conn)
	e := gob.NewEncoder(conn)
  for {
	  var req LockRequest
	err = d.Decode(&req)
	  if err != nil {
		if err == io.EOF {
			// other end interrupted connection
			break
		}
		fmt.Println("Error reading:", err.Error())
		continue
	  }
	  fmt.Println("received request:", req)
	  
	  if req.VersionMajor > VersionMajor {
		  fmt.Println("skipping request with superior major version")
		  continue
	  }
	  
	  var res LockResponse
	  res.LockRequest = req
	  res.VersionMajor, res.VersionMinor = VersionMajor, VersionMinor
	  //TODO: finish this
	  res.Result = Denied
	  res.Reason = "not yet implemented"
	  
	  // Send a response back to person contacting us.
	  err = e.Encode(&res)
	  if err != nil {
		if err == io.EOF {
			// other end interrupted connection
			break
		}
		fmt.Println("Error writing:", err.Error())
		continue
	  }
	}
  // Close the connection when you're done with it.
  conn.Close()
  fmt.Println("a client disconnected")
}
