package dlclient

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestAcquire(t *testing.T) {
	lockName := fmt.Sprintf("testing-%d", rand.Int())

	l, err := testClient.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}
	err = l.Release()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRelease(t *testing.T) {
}

func TestPeekNonExisting(t *testing.T) {
}

func TestPeekStale(t *testing.T) {
}

func TestPeekExisting(t *testing.T) {
}
