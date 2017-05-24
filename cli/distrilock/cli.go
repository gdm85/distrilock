package main

import (
	"errors"
	"fmt"
	"net"
	"os"

	flag "github.com/ogier/pflag"
)

const defaultAddress = ":13123"

type flags struct {
	*flag.FlagSet

	Address string
	Name    string

	// some parsed options
	a *net.TCPAddr
}

// mustParseFlags parses valid command-line flags for distrilock or returns an error; if help flag was selected, it exits the process.
func mustParseFlags(args []string) (*flags, error) {
	if len(args) < 1 {
		return nil, errors.New("empty arguments")
	}
	var f flags
	f.FlagSet = flag.NewFlagSet(args[0], flag.ExitOnError)

	f.FlagSet.StringVarP(&f.Address, "address", "a", defaultAddress, "address to listen on")
	f.FlagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: distrilock [--address %s]\n\n", defaultAddress)
		flag.PrintDefaults()
	}

	err := f.FlagSet.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	if len(f.FlagSet.Args()) != 0 {
		return nil, errors.New("unknown extra command line arguments specified")
	}

	// validate address
	f.a, err = net.ResolveTCPAddr("tcp", f.Address)
	if err != nil {
		return nil, err
	}

	return &f, nil
}
