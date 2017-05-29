// This package contains the command line interface executable 'distrilock'.
// To read its command line help, run:
/* $ bin/distrilock --help */
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
	"fmt"
	"net"
	"os"
	"time"

	"bitbucket.org/gdm85/go-distrilock/cli"
)

const defaultKeepAlive = time.Second * 3
const defaultAddress = ":13123"

func main() {
	f, err := flags.Parse(os.Args, defaultAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// print information about maximum number of files
	noFile, err := flags.GetNumberOfFilesLimit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(2)
	}
	if noFile <= 1024 {
		// print maximum number of files only when it's a bit low
		fmt.Println("distrilock: maximum number of files allowed is", noFile)
	}

	// validate address
	addr, err := net.ResolveTCPAddr("tcp", f.Address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(3)
	}

	// Listen for incoming connections.
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println("distrilock: error listening:", err.Error())
		os.Exit(4)
	}
	fmt.Println("distrilock: listening on", addr)
	for {
		// Listen for an incoming connection.
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error accepting: ", err.Error())
			continue
		}
		// Handle connections in a new goroutine.
		go handleRequests(f.Directory, conn, defaultKeepAlive)
	}
}
