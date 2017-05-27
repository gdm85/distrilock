package client_test

import (
	"testing"

	"bitbucket.org/gdm85/go-distrilock/api"
	"bitbucket.org/gdm85/go-distrilock/api/client"
)

func TestAcquireAndRelease(t *testing.T) {
	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()
			lockName := generateLockName(t)
			l, err := cs.testClientA1.Acquire(lockName)
			if err != nil {
				t.Error(lockName, err)
				return
			}
			err = l.Release()
			if err != nil {
				t.Error(err)
				return
			}
		})
	}
}

func TestAcquireVerifyAndRelease(t *testing.T) {
	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()

			lockName := generateLockName(t)

			l, err := cs.testClientA1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}

			err = l.Verify()
			if err != nil {
				t.Error(err)
				return
			}

			err = l.Release()
			if err != nil {
				t.Error(err)
				return
			}
		})
	}
}

func TestReleaseNonExisting(t *testing.T) {
	lockName := generateLockName(t)

	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()

			l := &client.Lock{
				Client: cs.testClientA1,
				Name:   lockName,
			}

			err := l.Release()
			if err == nil || err.Error() != "Failed: lock not found" {
				t.Error("expected lock not found, but got", err)
			}
		})
	}
}

func TestPeekNonExisting(t *testing.T) {
	lockName := generateLockName(t)

	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()

			isLocked, err := cs.testClientA1.IsLocked(lockName)
			if err != nil || isLocked {
				t.Error("expected no error and no lock, but got", err, isLocked)
			}
		})
	}
}

func TestPeekExisting(t *testing.T) {
	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()

			lockName := generateLockName(t)

			l, err := cs.testClientA1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}

			isLocked, err := cs.testClientA1.IsLocked(lockName)
			if err != nil || !isLocked {
				t.Error("expected no error and lock acquired, but got", err, isLocked)
				return
			}

			err = l.Release()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAcquireTwice(t *testing.T) {
	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()

			lockName := generateLockName(t)

			l1, err := cs.testClientA1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}

			l2, err := cs.testClientA1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}

			err = l1.Release()
			if err != nil {
				t.Error(err)
				return
			}
			err = l2.Release()
			if err == nil || err.Error() != "Failed: lock not found" {
				t.Error("expected lock not found error, but got", err)
				return
			}
		})
	}
}

func TestAcquireContention(t *testing.T) {
	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()

			lockName := generateLockName(t)

			l1, err := cs.testClientA1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}

			_, err = cs.testClientA2.Acquire(lockName)
			if err == nil {
				t.Error("expected failure to acquire lock already acquired from other session")
				return
			}
			e, ok := err.(*client.Error)
			if !ok {
				t.Error("expected client error")
				return
			}
			if e.Result != api.Failed {
				t.Error("expected Failed error, got", e.Result)
				return
			}

			// check that lock is acquired from 2nd client's perspective
			isLocked, err := cs.testClientA2.IsLocked(lockName)
			if err != nil || !isLocked {
				t.Error("expected no error and lock acquired, but got", err, isLocked)
				return
			}

			err = l1.Release()
			if err != nil {
				t.Error(err)
				return
			}

			l2, err := cs.testClientA2.Acquire(lockName)
			if err != nil {
				t.Error("expected success to acquire lock after it was released, got", err)
				return
			}

			err = l2.Release()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAcquireAndReleaseDiffProc(t *testing.T) {
	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()

			lockName := generateLockName(t)

			l, err := cs.testClientA1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}

			// here something nasty happens
			l.Client = cs.testClientB1

			err = l.Release()
			if err == nil || err.Error() != "Failed: lock not found" {
				t.Error("expected lock not found failure, but got", err)
				return
			}

			err = l.Verify()
			if err == nil || err.Error() != "Failed: lock not found" {
				t.Error("expected lock not found failure, but got", err)
				return
			}

			// restore
			l.Client = cs.testClientA1
			err = l.Release()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAcquireTwiceDiffProc(t *testing.T) {
	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()

			lockName := generateLockName(t)

			l1, err := cs.testClientA1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}

			_, err = cs.testClientB1.Acquire(lockName)
			if err == nil || err.Error() != "Failed: resource acquired by different process" {
				t.Error("expected failure, got", err)
				return
			}

			err = l1.Release()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAcquireAfterDiffProcRelease(t *testing.T) {
	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()

			lockName := generateLockName(t)

			l1, err := cs.testClientA1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}

			err = l1.Release()
			if err != nil {
				t.Error(err)
				return
			}

			l1, err = cs.testClientB1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}

			err = l1.Release()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAcquireContentionDiffProc(t *testing.T) {
	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()

			lockName := generateLockName(t)

			l1, err := cs.testClientA1.Acquire(lockName)
			if err != nil {
				t.Error(err)
				return
			}

			_, err = cs.testClientB1.Acquire(lockName)
			if err == nil {
				t.Error("expected failure to acquire lock already acquired from other session")
				return
			}
			e, ok := err.(*client.Error)
			if !ok {
				t.Error("expected client error")
				return
			}
			if e.Result != api.Failed {
				t.Error("expected Failed error, got", e.Result)
				return
			}

			// check that lock is acquired from 2nd client's perspective
			isLocked, err := cs.testClientB1.IsLocked(lockName)
			if err != nil || !isLocked {
				t.Error("expected no error and lock acquired, but got", err, isLocked)
				return
			}

			err = l1.Release()
			if err != nil {
				t.Error(err)
				return
			}

			l2, err := cs.testClientB1.Acquire(lockName)
			if err != nil {
				t.Error("expected success to acquire lock after it was released, got", err)
				return
			}

			err = l2.Release()
			if err != nil {
				t.Error(err)
				return
			}
		})
	}
}