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
)

func BenchmarkLocksTaking(b *testing.B) {
	type lockTakingTest struct {
		name    string
		nameGen func() string
		cs      *clientSuite
	}

	fixedLockName := generateLockName(b)

	var benchmarks []lockTakingTest

	// add a couple of benchmarks for each clients suite
	for _, cs := range clientSuites {
		benchmarks = append(benchmarks, []lockTakingTest{
			{
				name: cs.name + " random locks",
				nameGen: func() string {
					return generateLockName(b)
				},
				cs: cs,
			},
			{
				name: cs.name + " fixed locks",
				nameGen: func() string {
					return fixedLockName
				},
				cs: cs,
			},
		}...)
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lockName := bm.nameGen()

				c := bm.cs.createLocalClient()

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
