package client_test

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
	"bitbucket.org/gdm85/go-distrilock/api/client/tcp"
)

const (
	defaultServerA = ":63419"
	defaultServerB = ":63420"
	defaultServerC = ":63421"
	defaultServerD = "sibling:63422"

	deterministicTests = true
)

var (
	tcpClientSuite, websocketClientSuite *clientSuite
	clientSuites                         []*clientSuite
)

type clientSuite struct {
	name string
	websockets bool

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

func newClientSuite(websockets bool) *clientSuite {
	var cs clientSuite
	cs.websockets = websockets
	if websockets {
		cs.name = "Websockets clients suite"
	} else {
		cs.name = "TCP clients suite"
	}
	
	if websockets {
	} else {
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
		cs.testNFSRemoteAddr, err = net.ResolveTCPAddr("tcp", defaultServerD)
		if err != nil {
			panic(err)
		}
	}

	cs.testClientA1 = cs.createLocalClient()
	cs.testClientA2 = cs.createLocalClient()
	cs.testClientB1 = cs.createLocalAltClient()
	cs.testClientC1 = cs.createSlowNFSLocalClient()
	cs.testClientD1 = cs.createNFSRemoteClient()

	return &cs
}

func (cs *clientSuite) createSlowNFSLocalClient() client.Client {
	if cs.websockets {
		panic("WRITE ME")
	}
	return createSlowClient(cs.testNFSLocalAddr)
}

func (cs *clientSuite) createNFSRemoteClient() client.Client {
	if cs.websockets {
		panic("WRITE ME")
	}
	return createClient(cs.testNFSRemoteAddr)
}

func (cs *clientSuite) createLocalClient() client.Client {
	if cs.websockets {
		panic("WRITE ME")
	}
	return createClient(cs.testLocalAddr)
}

func (cs *clientSuite) createLocalAltClient() client.Client {
	if cs.websockets {
		panic("WRITE ME")
	}
		// a second process accessing same locks
		b, err := net.ResolveTCPAddr("tcp", defaultServerB)
		if err != nil {
			panic(err)
		}

	return createClient(b)
}


func (cs *clientSuite) CloseAll() {
	// close all clients
	for _, c := range []client.Client{cs.testClientA1, cs.testClientA2, cs.testClientB1, cs.testClientC1, cs.testClientD1} {
		err := c.Close()
		if err != nil {
			panic(err)
		}
	}
}

func createClient(a *net.TCPAddr) client.Client {
	return dlclient.New(a, time.Second*3, time.Second*2, time.Second*2)
}

func createSlowClient(a *net.TCPAddr) client.Client {
	return dlclient.New(a, time.Second*3, time.Second*15, time.Second*15)
}

func TestMain(m *testing.M) {
	tcpClientSuite = newClientSuite(false)
	//websocketClientSuite = newClientSuite(true)

	//clientSuites = []*clientSuite{tcpClientSuite, websocketClientSuite}
	clientSuites = []*clientSuite{tcpClientSuite}

	retCode := m.Run()

	for _, cs := range clientSuites {
		cs.CloseAll()
	}

	os.Exit(retCode)
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
