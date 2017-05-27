package dlclient

import (
	"fmt"
	"sync"
	"testing"

	"bitbucket.org/gdm85/go-distrilock/api"
)

func TestAcquireContentionNFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	lockName := generateLockName(t)

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

	lockName := generateLockName(t)

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

	// fix it back
	l.c = testClientC1
	err = l.Release()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAcquireTwiceNFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	lockName := generateLockName(t)

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

	lockName := generateLockName(t)

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

func disabledTestAcquireRaceNFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	checkWithRetries := func(lockName string, maxRetries int) (int, bool, error) {
		var err error
		var isLocked bool
		var retry int
		for retry < maxRetries {
			isLocked, err = testClientD1.IsLocked(lockName)
			if err != nil {
				return retry, isLocked, err
			}

			if isLocked {
				return retry, true, nil
			}

			retry++
		}

		return retry, isLocked, err
	}

	pfix := generateLockName(t)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		lockName := fmt.Sprintf("%s-%d", pfix, i)

		wg.Add(1)
		go func(lockName string) {
			defer wg.Done()

			c := createClient(testNFSLocalAddr)
			defer func() {
				err := c.Close()
				if err != nil {
					t.Error(err)
				}
			}()

			l1, err := testClientC1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}
			retries, isLocked, err := checkWithRetries(lockName, 15)
			if err != nil || !isLocked {
				t.Errorf("%s: expected no error and lock acquired, but got err=%v and isLocked=%v after %d retries", lockName, err, isLocked, retries)

				// release resources
				l1.Release()

				return
			}
			err = l1.Release()
			if err != nil {
				t.Error(err)
			}

			if retries != 0 {
				t.Errorf("%s: lock check after %d retries", lockName, retries)
			}

		}(lockName)
	}

	wg.Wait()
}

func TestAcquireTwiceRaceNFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	pfix := generateLockName(t)

	var wg sync.WaitGroup
	for i := 0; i < 1100; i++ {
		lockName := fmt.Sprintf("%s-%d", pfix, i)

		wg.Add(1)
		go func(lockName string) {
			defer wg.Done()

			c, d := createSlowClient(testNFSLocalAddr), createClient(testNFSRemoteAddr)
			defer d.Close()
			defer c.Close()

			l1, err := c.Acquire(lockName)
			if err != nil {
				t.Error("first lock acquire:", err)
				return
			}
			err = l1.Verify()
			if err != nil {
				t.Error("first lock verify:", err)
				return
			}

			l2, err := d.Acquire(lockName)
			if err == nil {
				///
				/// somehow, two locks with the same name were retrieved
				/// now it's time for a bit of Sherlock Holmes' investigation
				///
				err = l1.Verify()
				err2 := l2.Verify()
				t.Errorf("%s: lock acquired twice, verifications: %v %v", lockName, err, err2)

				// at end, attempt to politely release both locks
				err = l1.Release()
				if err != nil {
					t.Error("first lock release:", err)
				}
				err = l2.Release()
				if err2 != nil {
					t.Error("second lock release:", err2)
				}
				return
			}
			if err.Error() != "Failed: resource acquired by different process" {
				t.Errorf("%s: expected Failed error but got err=%v", lockName, err)
				return
			}
			err = l1.Release()
			if err != nil {
				t.Error("first lock release:", err)
			}

		}(lockName)
	}

	wg.Wait()
}
