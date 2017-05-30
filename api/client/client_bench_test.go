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

	"bitbucket.org/gdm85/go-distrilock/api/client"
)

func BenchmarkLocksTaking(b *testing.B) {
	lockName := generateLockName(b)

	for _, cs :=  range clientSuites {
		fixedClient := cs.createLocalClient()

		b.Run(cs.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var c client.Client
				// if clients are concurrency-safe, then no need to get a new client for each lock
				if cs.concurrencySafe {
					c = fixedClient
				} else {
					c = cs.createLocalClient()
				}

				l, err := c.Acquire(lockName)
				if err != nil {
					b.Error(err)
					c.Close()
					return
				}
				err = l.Release()
				if err != nil {
					b.Error(err)
					c.Close()
					return
				}

				err = c.Close()
				if err != nil {
					b.Error(err)
					return
				}
			}
		})
	}
}
