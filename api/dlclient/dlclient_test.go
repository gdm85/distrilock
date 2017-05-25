package dlclient

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"bitbucket.org/gdm85/go-distrilock/api"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestAcquireAndRelease(t *testing.T) {
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

func TestReleaseNonExisting(t *testing.T) {
	lockName := fmt.Sprintf("testing-%d", rand.Int())

	l := &Lock{
		c:    testClient,
		name: lockName,
	}

	err := l.Release()
	if err == nil || err.Error() != "Failed: lock not found" {
		t.Error("expected lock not found, but got", err)
	}
}

func TestPeekNonExisting(t *testing.T) {
	lockName := fmt.Sprintf("testing-%d", rand.Int())

	isLocked, err := testClient.IsLocked(lockName)
	if err != nil || isLocked {
		t.Error("expected no error and no lock, but got", err, isLocked)
	}
}

func TestPeekExisting(t *testing.T) {
	lockName := fmt.Sprintf("testing-%d", rand.Int())

	l, err := testClient.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	isLocked, err := testClient.IsLocked(lockName)
	if err != nil || !isLocked {
		t.Error("expected no error and lock acquired, but got", err, isLocked)
		return
	}

	err = l.Release()
	if err != nil {
		t.Error(err)
	}
}

func TestPeekStale(t *testing.T) {
}

func TestAcquireTwice(t *testing.T) {
	lockName := fmt.Sprintf("testing-%d", rand.Int())

	l1, err := testClient.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	l2, err := testClient.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Release()
	if err != nil {
		t.Fatal(err)
	}
	err = l2.Release()
	if err == nil || err.Error() != "Failed: lock not found" {
		t.Fatal("expected lock not found error, but got", err)
	}
}

func TestAcquireContention(t *testing.T) {
	lockName := fmt.Sprintf("testing-%d", rand.Int())

	l1, err := testClient.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testClient2.Acquire(lockName)
	if err == nil {
		t.Fatal("expected failure to acquire lock already acquired from other session")
	}
	e, ok := err.(*ClientError)
	if !ok {
		t.Fatal("expected client error")
	}
	if e.Result != api.Denied {
		t.Fatal("expected Denied error, got", e.Result)
	}

	// check that lock is acquired from 2nd client's perspective
	isLocked, err := testClient2.IsLocked(lockName)
	if err != nil || !isLocked {
		t.Error("expected no error and lock acquired, but got", err, isLocked)
		return
	}

	err = l1.Release()
	if err != nil {
		t.Fatal(err)
	}

	l2, err := testClient2.Acquire(lockName)
	if err != nil {
		t.Fatal("expected success to acquire lock after it was released, got", err)
	}

	err = l2.Release()
	if err != nil {
		t.Fatal(err)
	}
}
