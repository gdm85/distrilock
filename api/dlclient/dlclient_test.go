package dlclient

import (
	"testing"

	"bitbucket.org/gdm85/go-distrilock/api"
)

func TestAcquireAndRelease(t *testing.T) {
	lockName := generateLockName(t)

	l, err := testClientA1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}
	err = l.Release()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAcquireVerifyAndRelease(t *testing.T) {
	lockName := generateLockName(t)

	l, err := testClientA1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	err = l.Verify()
	if err != nil {
		t.Fatal(err)
	}

	err = l.Release()
	if err != nil {
		t.Fatal(err)
	}
}

func TestReleaseNonExisting(t *testing.T) {
	lockName := generateLockName(t)

	l := &Lock{
		c:    testClientA1,
		name: lockName,
	}

	err := l.Release()
	if err == nil || err.Error() != "Failed: lock not found" {
		t.Error("expected lock not found, but got", err)
	}
}

func TestPeekNonExisting(t *testing.T) {
	lockName := generateLockName(t)

	isLocked, err := testClientA1.IsLocked(lockName)
	if err != nil || isLocked {
		t.Error("expected no error and no lock, but got", err, isLocked)
	}
}

func TestPeekExisting(t *testing.T) {
	lockName := generateLockName(t)

	l, err := testClientA1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	isLocked, err := testClientA1.IsLocked(lockName)
	if err != nil || !isLocked {
		t.Error("expected no error and lock acquired, but got", err, isLocked)
		return
	}

	err = l.Release()
	if err != nil {
		t.Error(err)
	}
}

func TestAcquireTwice(t *testing.T) {
	lockName := generateLockName(t)

	l1, err := testClientA1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	l2, err := testClientA1.Acquire(lockName)
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
	lockName := generateLockName(t)

	l1, err := testClientA1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testClientA2.Acquire(lockName)
	if err == nil {
		t.Fatal("expected failure to acquire lock already acquired from other session")
	}
	e, ok := err.(*ClientError)
	if !ok {
		t.Fatal("expected client error")
	}
	if e.Result != api.Failed {
		t.Fatal("expected Failed error, got", e.Result)
	}

	// check that lock is acquired from 2nd client's perspective
	isLocked, err := testClientA2.IsLocked(lockName)
	if err != nil || !isLocked {
		t.Error("expected no error and lock acquired, but got", err, isLocked)
		return
	}

	err = l1.Release()
	if err != nil {
		t.Fatal(err)
	}

	l2, err := testClientA2.Acquire(lockName)
	if err != nil {
		t.Fatal("expected success to acquire lock after it was released, got", err)
	}

	err = l2.Release()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAcquireAndReleaseDiffProc(t *testing.T) {
	lockName := generateLockName(t)

	l, err := testClientA1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	// here something nasty happens
	l.c = testClientB1

	err = l.Release()
	if err == nil || err.Error() != "Failed: lock not found" {
		t.Fatal("expected lock not found failure, but got", err)
	}

	err = l.Verify()
	if err == nil || err.Error() != "Failed: lock not found" {
		t.Fatal("expected lock not found failure, but got", err)
	}

	// restore
	l.c = testClientA1
	err = l.Release()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAcquireTwiceDiffProc(t *testing.T) {
	lockName := generateLockName(t)

	l1, err := testClientA1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testClientB1.Acquire(lockName)
	if err == nil || err.Error() != "Failed: resource acquired by different process" {
		t.Fatal("expected failure, got", err)
	}

	err = l1.Release()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAcquireAfterDiffProcRelease(t *testing.T) {
	lockName := generateLockName(t)

	l1, err := testClientA1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Release()
	if err != nil {
		t.Fatal(err)
	}

	l1, err = testClientB1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	err = l1.Release()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAcquireContentionDiffProc(t *testing.T) {
	lockName := generateLockName(t)

	l1, err := testClientA1.Acquire(lockName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testClientB1.Acquire(lockName)
	if err == nil {
		t.Fatal("expected failure to acquire lock already acquired from other session")
	}
	e, ok := err.(*ClientError)
	if !ok {
		t.Fatal("expected client error")
	}
	if e.Result != api.Failed {
		t.Fatal("expected Failed error, got", e.Result)
	}

	// check that lock is acquired from 2nd client's perspective
	isLocked, err := testClientB1.IsLocked(lockName)
	if err != nil || !isLocked {
		t.Error("expected no error and lock acquired, but got", err, isLocked)
		return
	}

	err = l1.Release()
	if err != nil {
		t.Fatal(err)
	}

	l2, err := testClientB1.Acquire(lockName)
	if err != nil {
		t.Fatal("expected success to acquire lock after it was released, got", err)
	}

	err = l2.Release()
	if err != nil {
		t.Fatal(err)
	}
}
