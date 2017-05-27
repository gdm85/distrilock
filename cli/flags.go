// Package flags defines the command line interface flags for both distrilock and distrilock-ws.
package flags

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
