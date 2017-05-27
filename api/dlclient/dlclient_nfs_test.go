package dlclient

import (
	"fmt"
	"sync"
	"math/rand"
	"testing"

	"bitbucket.org/gdm85/go-distrilock/api"
)

func TestAcquireContentionNFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	lockName := fmt.Sprintf("testing-%d", rand.Int())

	l1, err := testClientC1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testClientD1.Acquire(lockName)
	if err == nil {
		t.Fatal("expected failure to acquire lock already acquired from other session")
	}
	e, ok := err.(*ClientError)
	if !ok {
		t.Fatal("expected client error, got", err)
	}
	if e.Result != api.Failed {
		t.Fatal("expected Failed error, got", e.Result)
	}

	// check that lock is acquired from 2nd client's perspective
	isLocked, err := testClientD1.IsLocked(lockName)
	if err != nil || !isLocked {
		t.Error("expected no error and lock acquired, but got", err, isLocked)
		return
	}

	err = l1.Release()
	if err != nil {
		t.Fatal(err)
	}

	l2, err := testClientD1.Acquire(lockName)
	if err != nil {
		t.Fatal("expected success to acquire lock after it was released, got", err)
	}

	err = l2.Release()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAcquireAndReleaseNFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	lockName := fmt.Sprintf("testing-%d", rand.Int())

	l, err := testClientC1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	// here something nasty happens
	l.c = testClientD1

	err = l.Release()
	if err == nil || err.Error() != "Failed: lock not found" {
		t.Fatal("expected lock not found failure, but got", err)
	}
}

func TestAcquireTwiceNFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	lockName := fmt.Sprintf("testing-%d", rand.Int())

	l1, err := testClientC1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testClientD1.Acquire(lockName)
	if err == nil || err.Error() != "Failed: resource acquired by different process" {
		t.Fatal("expected failure, got", err)
	}

	err = l1.Release()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAcquireAfterReleaseNFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	lockName := fmt.Sprintf("testing-%d", rand.Int())

	l1, err := testClientC1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Release()
	if err != nil {
		t.Fatal(err)
	}

	l1, err = testClientD1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Release()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAcquireRaceNFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	
	pfix := fmt.Sprintf("testing-%d", rand.Int())
	
	var wg sync.WaitGroup
	for i:=0;i<1000;i++ {
		lockName := fmt.Sprintf("%s-%d", pfix, i)
		
		wg.Add(1)
		go func(lockName string) {
			defer wg.Done()
			l1, err := testClientC1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}
			isLocked, err := testClientD1.IsLocked(lockName)
			if err != nil || !isLocked {
				t.Errorf("%s: expected no error and lock acquired, but got err=%v and isLocked=%v", lockName, err, isLocked)
				return
			}
			err = l1.Release()
			if err != nil {
				t.Error(err)
			}
		}(lockName)
	}
	
	wg.Wait()
}

