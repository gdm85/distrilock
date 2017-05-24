package main

import (
	"errors"
	"fmt"
	"net"
	"os"

	flag "github.com/ogier/pflag"
)

type flags struct {
	*flag.FlagSet
	
	Address     string
	Client bool
	Name string
		
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

	f.FlagSet.StringVarP(&f.Address, "address", "a", ":13123", "address to listen on")
	f.FlagSet.BoolVarP(&f.Client, "client", "c", false, "perform a client connection test")
	f.FlagSet.StringVarP(&f.Name, "name", "n", "", "lock name to acquire during client connection test")
	f.FlagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: distrilock [--address :13123] [--client] [--name lock-name]\n\n")
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
	
	if f.Name != "" && !f.Client {
		return nil, errors.New("lock name cannot be specified in server mode")
	}
	
	if f.Client && f.Name == "" {
		return nil, errors.New("lock name must be specified in client mode")
	}

	return &f, nil
}
