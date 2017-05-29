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
	"io/ioutil"
	"os"
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

func TestAcquireAndReleaseStale(t *testing.T) {
	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()
			lockName := generateLockName(t)

			// simulate a pre-existing stale file
			err := ioutil.WriteFile(localLockDir+lockName+lockExt, []byte("test"), 0664)
			if err != nil {
				t.Error(err)
				return
			}

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

func TestAcquireInterfere(t *testing.T) {
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

			// here an external action is simulated, e.g. `cat filename.lck`
			f, err := os.OpenFile(localLockDir+lockName+lockExt, os.O_RDWR, 0664)
			if err != nil {
				t.Error(err)
				return
			}
			_, err = ioutil.ReadAll(f)
			if err != nil {
				t.Error(err)
				return
			}

			err = f.Close()
			if err != nil {
				t.Error(err)
				return
			}

			// try to acquire with a different client
			_, err = cs.testClientB1.Acquire(lockName)
			if err == nil || err.Error() != "Failed: resource acquired by different process" {
				t.Error("expected Failed, got", err)
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

func TestPeekStaleNonExisting(t *testing.T) {

	for _, cs := range clientSuites {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			t.Parallel()
			lockName := generateLockName(t)

			// simulate a pre-existing stale file
			err := ioutil.WriteFile(localLockDir+lockName+lockExt, []byte("test"), 0664)
			if err != nil {
				t.Error(err)
				return
			}

			isLocked, err := cs.testClientA1.IsLocked(lockName)
			if err != nil || isLocked {
				t.Error("expected no error and no lock, but got", err, isLocked)
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
