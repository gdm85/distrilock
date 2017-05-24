package dlclient

import (
	"net"
	"os"
	"os/exec"
	"testing"
	"time"
)

const defaultTestAddress = ":13290"

var defaultAddr *net.TCPAddr

func init() {
	var err error
	defaultAddr, err = net.ResolveTCPAddr("tcp", defaultTestAddress)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	cmd, err := startServerProcess()
	if err != nil {
		panic(err)
	}

	// wait for server to be online
	err = waitForServer()
	if err != nil {
		panic(err)
	}

	retCode := m.Run()

	err = stopServerProcess(cmd)
	if err != nil {
		panic(err)
	}

	os.Exit(retCode)
}

func startServerProcess() (*exec.Cmd, error) {
	cmd := exec.Command("../../bin/distrilock", "--address="+defaultTestAddress)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, cmd.Start()
}

const maxFailures = 2

func waitForServer() error {
	c := New(defaultAddr, time.Second*3, time.Second*2, time.Second*2)
	var failures int
	for {
		_, err := c.IsLocked("initial-test")
		if err == nil {
			// initial test succeeded
			return nil
		}
		failures++
		if failures == maxFailures {
			return err
		}
		time.Sleep(1)
	}
}

func stopServerProcess(cmd *exec.Cmd) error {
	err := cmd.Process.Kill()
	if err != nil {
		return err
	}
	// wait will return 'killed'
	_ = cmd.Wait()
	return nil
}
