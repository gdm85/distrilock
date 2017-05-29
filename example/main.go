// This package contains an example of a client connecting to a distrilock daemon listening on default port.
package main

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
	"time"
	"net"

	"bitbucket.org/gdm85/go-distrilock/api/client/tcp"
)

func main() {
	// resolve daemon address
	addr, err := net.ResolveTCPAddr("tcp", ":13123")
	if err != nil {
		panic(err)
	}

	// create client
	c := tcp.New(addr, time.Second*3, time.Second*3, time.Second*3)

	// acquire lock
	l, err := c.Acquire("my-named-lock")
	if err != nil {
		panic(err)
	}

	const totalWorkUnits = 3
	workDone := 0

	// start doing some intensive work
	for {
		///
		/// do some heavy work here, then iterate for some more heavy work
		time.Sleep(5)
		workDone++
		///
		if workDone == totalWorkUnits {
			break
		}

		// verify lock is still in good health
		err := l.Verify()
		if err != nil {
			panic(err)
		}
	}

	// release lock
	err = l.Release()
	if err != nil {
		panic(err)
	}

	// close connection
	err = c.Close()
	if err != nil {
		panic(err)
	}
}
