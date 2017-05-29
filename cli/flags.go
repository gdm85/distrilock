// Package flags defines the command line interface flags for both distrilock and distrilock-ws.
package flags

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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	flag "github.com/ogier/pflag"
)

// Flags contains the command line interface flags for a distrilock daemon.
type Flags struct {
	*flag.FlagSet

	Address   string
	Directory string
}

// Parse parses valid command-line flags for distrilock or returns an error; if help flag was selected, it exits the process.
func Parse(args []string, defaultAddress string) (*Flags, error) {
	if len(args) < 1 {
		return nil, errors.New("empty arguments")
	}
	var f Flags
	f.FlagSet = flag.NewFlagSet(args[0], flag.ExitOnError)

	f.FlagSet.StringVarP(&f.Address, "address", "a", defaultAddress, "address to listen on")
	f.FlagSet.StringVarP(&f.Directory, "directory", "d", ".", "directory where to locate locked files")
	f.FlagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: distrilock [--address=%s] [--directory=.]\n\n", defaultAddress)
		flag.PrintDefaults()
	}

	err := f.FlagSet.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	if len(f.FlagSet.Args()) != 0 {
		return nil, errors.New("unknown extra command line arguments specified")
	}

	// validate directory
	f.Directory, err = filepath.Abs(f.Directory)
	if err != nil {
		return nil, err
	}
	f.Directory += "/"

	return &f, nil
}

// GetNumberOfFilesLimit returns the (hard) limit for maximum number of files for the user running current process.
func GetNumberOfFilesLimit() (uint64, error) {
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return 0, err
	}
	return limit.Max, nil
}
