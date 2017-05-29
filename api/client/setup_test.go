package client_test

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
	"math/rand"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"bitbucket.org/gdm85/go-distrilock/api/client"
	"bitbucket.org/gdm85/go-distrilock/api/client/concurrent"
	"bitbucket.org/gdm85/go-distrilock/api/client/tcp"
	"bitbucket.org/gdm85/go-distrilock/api/client/ws"

	"github.com/gorilla/websocket"
)

const (
	// first and default locally running daemon
	defaultServerA = ":63419"
	// second locally running daemon
	defaultServerB = ":63420"
	// locally running daemon on an NFS-shared directory
	defaultServerC = ":63421"
	// daemon running on a separate host, sharing same directory via NFS
	defaultServerD = "sibling:63422"

	defaultWebsocketServerA = "ws://localhost:63519/distrilock"
	defaultWebsocketServerB = "ws://localhost:63520/distrilock"
	defaultWebsocketServerC = "ws://localhost:63521/distrilock"
	defaultWebsocketServerD = "ws://sibling:63522/distrilock"

	deterministicTests = true
)

var (
	clientSuites []*clientSuite
)

type clientSuite struct {
	name            string
	clientType      int
	concurrencySafe bool

	testClientA1, testClientA2 client.Client
	testClientB1               client.Client
	testClientC1               client.Client
	testClientD1               client.Client

	// internal
	testLocalAddr, testNFSLocalAddr, testNFSRemoteAddr *net.TCPAddr
}

func init() {
	if !deterministicTests {
		rand.Seed(time.Now().UTC().UnixNano())
	} else {
		rand.Seed(63419)
	}
}

// copied from process.go
const lockExt = ".lck"

var (
	shortMode    bool
	localLockDir string
)

func TestMain(m *testing.M) {
	// trick to detect short mode
	for _, arg := range os.Args[1:] {
		if arg == "-test.short=true" {
			shortMode = true
			break
		}
	}

	// local lock directory
	localLockDir = os.Getenv("LOCAL_LOCK_DIR") + "/"

	clientSuites = []*clientSuite{
		newClientSuite(0, false), newClientSuite(websocket.BinaryMessage, false), newClientSuite(websocket.TextMessage, false),
		newClientSuite(0, true), newClientSuite(websocket.BinaryMessage, true), newClientSuite(websocket.TextMessage, true),
	}

	retCode := m.Run()

	for _, cs := range clientSuites {
		cs.CloseAll()
	}

	os.Exit(retCode)
}

func newClientSuite(clientType int, concurrencySafe bool) *clientSuite {
	var cs clientSuite
	cs.concurrencySafe = concurrencySafe
	cs.clientType = clientType
	switch clientType {
	case websocket.BinaryMessage:
		cs.name = "Websockets binary clients suite"
	case websocket.TextMessage:
		cs.name = "Websockets text clients suite"
	default:
		cs.name = "TCP clients suite"
	}

	if concurrencySafe {
		cs.name += " concurrency-safe"
	}

	if clientType == 0 {
		// first server process
		var err error
		cs.testLocalAddr, err = net.ResolveTCPAddr("tcp", defaultServerA)
		if err != nil {
			panic(err)
		}

		// first process on NFS
		cs.testNFSLocalAddr, err = net.ResolveTCPAddr("tcp", defaultServerC)
		if err != nil {
			panic(err)
		}
		// second process on NFS, different machine
		if !shortMode {
			cs.testNFSRemoteAddr, err = net.ResolveTCPAddr("tcp", defaultServerD)
			if err != nil {
				panic(err)
			}
		}
	}

	cs.testClientA1 = cs.createLocalClient()
	cs.testClientA2 = cs.createLocalClient()
	cs.testClientB1 = cs.createLocalAltClient()
	cs.testClientC1 = cs.createSlowNFSLocalClient()
	if !shortMode {
		cs.testClientD1 = cs.createNFSRemoteClient()
	}

	if concurrencySafe {
		cs.testClientA1 = concurrent.New(cs.testClientA1)
		cs.testClientA2 = concurrent.New(cs.testClientA2)
		cs.testClientB1 = concurrent.New(cs.testClientB1)
		cs.testClientC1 = concurrent.New(cs.testClientC1)
		if !shortMode {
			cs.testClientD1 = concurrent.New(cs.testClientD1)
		}
	}

	return &cs
}

func (cs *clientSuite) createSlowNFSLocalClient() client.Client {
	switch cs.clientType {
	case websocket.BinaryMessage:
		return ws.NewBinary(defaultWebsocketServerC, time.Second*3, time.Second*15, time.Second*15)
	case websocket.TextMessage:
		return ws.NewJSON(defaultWebsocketServerC, time.Second*3, time.Second*15, time.Second*15)
	}
	return createTCPSlowClient(cs.testNFSLocalAddr)
}

func (cs *clientSuite) createNFSRemoteClient() client.Client {
	switch cs.clientType {
	case websocket.BinaryMessage:
		return ws.NewBinary(defaultWebsocketServerD, time.Second*3, time.Second*15, time.Second*15)
	case websocket.TextMessage:
		return ws.NewJSON(defaultWebsocketServerD, time.Second*3, time.Second*15, time.Second*15)
	}
	return createTCPClient(cs.testNFSRemoteAddr)
}

func (cs *clientSuite) createLocalClient() client.Client {
	switch cs.clientType {
	case websocket.BinaryMessage:
		return ws.NewBinary(defaultWebsocketServerA, time.Second*3, time.Second*2, time.Second*15)
	case websocket.TextMessage:
		return ws.NewJSON(defaultWebsocketServerA, time.Second*3, time.Second*2, time.Second*15)
	}
	return createTCPClient(cs.testLocalAddr)
}

func (cs *clientSuite) createLocalAltClient() client.Client {
	switch cs.clientType {
	case websocket.BinaryMessage:
		return ws.NewBinary(defaultWebsocketServerB, time.Second*3, time.Second*2, time.Second*15)
	case websocket.TextMessage:
		return ws.NewJSON(defaultWebsocketServerB, time.Second*3, time.Second*2, time.Second*15)
	}
	// a second process accessing same locks
	b, err := net.ResolveTCPAddr("tcp", defaultServerB)
	if err != nil {
		panic(err)
	}

	return createTCPClient(b)
}

func createTCPClient(a *net.TCPAddr) client.Client {
	return tcp.New(a, time.Second*3, time.Second*2, time.Second*2)
}

func createTCPSlowClient(a *net.TCPAddr) client.Client {
	return tcp.New(a, time.Second*3, time.Second*15, time.Second*15)
}

func (cs *clientSuite) CloseAll() {
	// close all clients
	for _, c := range []client.Client{cs.testClientA1, cs.testClientA2, cs.testClientB1, cs.testClientC1, cs.testClientD1} {
		if c != nil {
			_ = c.Close()
		}
	}
}

// generateLockName is an utility function to generate a randomised name of a test.
func generateLockName(bOrT interface{}) string {
	var nameV reflect.Value
	switch v := bOrT.(type) {
	case *testing.T:
		nameV = reflect.ValueOf(*v).FieldByName("name")
	case *testing.B:
		nameV = reflect.ValueOf(*v).FieldByName("name")
	default:
		panic("BUG: passed invalid type to generateLockName")
	}

	s := fmt.Sprintf("%v-%d", nameV, rand.Int())
	s = strings.Replace(s, "/", "-", -1)
	return strings.Replace(s, "#", "-", -1)
}
