package dlclient

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"reflect"
	"testing"
	"time"
)

const (
	defaultServerA = ":63419"
	defaultServerB = ":63420"
	defaultServerC = ":63421"
	defaultServerD = "sibling:63422"

	deterministicTests = true
)

var (
	testClientA1, testClientA2 *Client
	testClientB1               *Client
	testClientC1               *Client
	testClientD1               *Client
)

func init() {
	if !deterministicTests {
		rand.Seed(time.Now().UTC().UnixNano())
	} else {
		rand.Seed(63419)
	}

	// first server process
	a, err := net.ResolveTCPAddr("tcp", defaultServerA)
	if err != nil {
		panic(err)
	}
	// a second process accessing same locks
	b, err := net.ResolveTCPAddr("tcp", defaultServerB)
	if err != nil {
		panic(err)
	}

	// first process on NFS
	c, err := net.ResolveTCPAddr("tcp", defaultServerC)
	if err != nil {
		panic(err)
	}
	// second process on NFS, different machine
	d, err := net.ResolveTCPAddr("tcp", defaultServerD)
	if err != nil {
		panic(err)
	}

	testClientA1 = New(a, time.Second*3, time.Second*2, time.Second*2)
	testClientA2 = New(a, time.Second*3, time.Second*2, time.Second*2)
	testClientB1 = New(b, time.Second*3, time.Second*2, time.Second*2)
	testClientC1 = New(c, time.Second*3, time.Second*15, time.Second*15)
	testClientD1 = New(d, time.Second*3, time.Second*2, time.Second*2)
}

func TestMain(m *testing.M) {
	retCode := m.Run()

	// close all clients
	for _, c := range []*Client{testClientA1, testClientA2, testClientB1, testClientC1, testClientD1} {
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
