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
	"testing"

	"github.com/gdm85/distrilock/api/client"
)

func BenchmarkSuiteAcquireAndRelease(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping test in short mode.")
	}

	lockName := generateLockName(b)

	for _, cs := range clientSuites {
		if cs.concurrencySafe {
			// not covered by this benchmark
			continue
		}
		c := cs.createNFSRemoteClient()

		// the very first acquire/release is out of the benchmark as a "warm up" for the underlying connection
		l, err := c.Acquire(lockName)
		if err != nil {
			b.Error(err)
			return
		}
		err = l.Release()
		if err != nil {
			b.Error(err)
			return
		}

		b.Run(cs.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				l, err := c.Acquire(lockName)
				if err != nil {
					b.Error(err)
					return
				}

				err = l.Release()
				if err != nil {
					b.Error(err)
					return
				}
			}
		})

		err = c.Close()
		if err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkSuiteInitialConn(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping test in short mode.")
	}

	for _, cs := range clientSuites {
		var clients chan client.Client
		cleanupCtl := make(chan struct{})

		// close connections as benchmark progresses
		// technically not a correct split of initial connection vs connection cleanup,
		// but allowed as similar to real use scenarios
		go func() {
			<-cleanupCtl
			for c := range clients {
				err := c.Close()
				if err != nil {
					b.Error(err)
					return
				}
			}
			cleanupCtl <- struct{}{}
		}()

		b.Run(cs.name, func(b *testing.B) {
			clients = make(chan client.Client, b.N)
			cleanupCtl <- struct{}{}

			for i := 0; i < b.N; i++ {
				lockName := generateLockName(b)
				c := cs.createNFSRemoteClient()

				_, err := c.Acquire(lockName)
				if err != nil {
					b.Error(err)
					return
				}

				clients <- c
			}
		})
		close(clients)

		<-cleanupCtl
		close(cleanupCtl)
	}
}
