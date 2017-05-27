package client_test

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"reflect"
	"testing"
	"time"
	
	"bitbucket.org/gdm85/go-distrilock/api/client/tcp"
	"bitbucket.org/gdm85/go-distrilock/api/client"
)

const (
	defaultServerA = ":63419"
	defaultServerB = ":63420"
	defaultServerC = ":63421"
	defaultServerD = "sibling:63422"

	deterministicTests = true
)

type clientSuite struct {
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

func initClientSuite(websockets bool) *clientSuite {
	var cs clientSuite
	// first server process
	var err error
	cs.testLocalAddr, err = net.ResolveTCPAddr("tcp", defaultServerA)
	if err != nil {
		panic(err)
	}
	// a second process accessing same locks
	b, err := net.ResolveTCPAddr("tcp", defaultServerB)
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

	cs.testClientA1 = cs.createLocalClient()
	cs.testClientA2 = cs.createLocalClient()
	cs.testClientB1 = createClient(b)
	cs.testClientC1 = cs.createSlowNFSLocalClient()
	cs.testClientD1 = cs.createNFSRemoteClient()
	
	return &cs
}

func(cs *clientSuite) createSlowNFSLocalClient() client.Client {
	return createSlowClient(cs.testNFSLocalAddr)
}

func(cs *clientSuite) createNFSRemoteClient() client.Client {
	return createClient(cs.testNFSRemoteAddr)
}

func(cs *clientSuite) createLocalClient() client.Client {
	return createClient(cs.testLocalAddr)
}

func createClient(a *net.TCPAddr) client.Client {
	return dlclient.New(a, time.Second*3, time.Second*2, time.Second*2)
}

func createSlowClient(a *net.TCPAddr) client.Client {
	return dlclient.New(a, time.Second*3, time.Second*15, time.Second*15)
}

var cs *clientSuite

func TestMain(m *testing.M) {
	cs = initClientSuite(false)
	
	retCode := m.Run()
	
	// close all clients
	for _, c := range []client.Client{cs.testClientA1, cs.testClientA2, cs.testClientB1, cs.testClientC1, cs.testClientD1} {
		err := c.Close()
		if err != nil {
			panic(err)
		}
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

	return fmt.Sprintf("%v-%d", nameV, rand.Int())

}
