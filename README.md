# distrilock
Distributed locking for pennies.

distrilock is a TCP and Websockets daemon that leverages Linux [POSIX filesystem locks](http://pubs.opengroup.org/onlinepubs/9699919799/functions/fcntl.html) and is compatible with NFSv4 mountpoints.
It does not use anything else that a Linux filesystem to enforce locking with `fcntl`; locks persistence is tied 1:1 with connection persistence e.g. **no leases**.

A typical deployment would consist of one or more distrilock daemons running on different hosts and sharing the same directory as an NFSv4 mountpoint.

## Terminology and basic functionality

A **daemon** is a distrilock server-side daemon listening for incoming connections - either TCP or Websockets.

A **client** is the client connecting to a daemon; to each client corresponds one **session** when the connection is estabilished.

Each session can acquire one more **named locks**; the connection fundamental to the session is kept alive with TCP keep-alive functionality; if such connection is interrupted or closed, the lock is immediately released and lost.

## How (distributed) locking works

1. the daemon instance tracks multiple named locks per session (= connection), which correspond to an individual file
2. client requests lock acquisition through daemon, if viable a `fcntl` write lock is acquired on the corresponding file and lock is returned as successful
3. client performs some work within the locked context
4. client releases lock through daemon, the `fcntl` write lock is released on the open file which is then closed and deleted

## NFSv4 specific notes

The default `local_lock=none` must be part of your mountpoint options; [version 4 of the NFS protocol](https://tools.ietf.org/html/rfc7530) is necessary.

## Building

After cloning the repository and putting it in an independent GOPATH directory structure, run:
```bash
$ make
if ! ls vendor/github.com/ogier/pflag/* 2>/dev/null >/dev/null; then git submodule update --init --recursive; fi
mkdir -p bin
GOBIN="/home/gdm85/.../go-distrilock/bin" go install ./cli ./cli/distrilock ./api ./api/client ./api/core ./api/client/tcp ./cli/distrilock-ws ./api/client/ws
scripts/run-tests.sh ./cli ./cli/distrilock ./api ./api/client ./api/core ./api/client/tcp ./cli/distrilock-ws ./api/client/ws
Running all tests
distrilock: listening on :63419
distrilock: listening on :63420
distrilock-ws: listening on localhost:63519
distrilock-ws: listening on localhost:63520
?   	bitbucket.org/gdm85/go-distrilock/cli	[no test files]
?   	bitbucket.org/gdm85/go-distrilock/cli/distrilock	[no test files]
?   	bitbucket.org/gdm85/go-distrilock/api	[no test files]
ok  	bitbucket.org/gdm85/go-distrilock/api/client	0.017s
?   	bitbucket.org/gdm85/go-distrilock/api/core	[no test files]
?   	bitbucket.org/gdm85/go-distrilock/api/client/tcp	[no test files]
?   	bitbucket.org/gdm85/go-distrilock/cli/distrilock-ws	[no test files]
?   	bitbucket.org/gdm85/go-distrilock/api/client/ws	[no test files]
```

The binaries `bin/distrilock` (TCP daemon) and `bin/distrilock-ws` (Websockets daemon) will be available.

## Tests

Test targets will start multiple distrilock daemons and terminate them at end of test run.
Normal tests (`make test`) are executed by default when building; in order to execute the NFS-specific tests, provide the following changes:

* add a host called `sibling` in your `/etc/hosts` pointing to a second machine running NFSv4
* run tests with `NFS_SHARE=/mnt/your-nfs-share make test`

Race condition tests:
```bash
$ make race
```

Benchmark tests:
```bash
$ make benchmark
scripts/run-tests.sh -bench=. -benchtime=1s ./cli ./cli/distrilock ./api ./api/client ./api/core ./api/client/tcp ./cli/distrilock-ws ./api/client/ws
Running all tests
distrilock: listening on :63419
distrilock-ws: listening on localhost:63519
distrilock-ws: listening on localhost:63520
distrilock: listening on :63420
?   	bitbucket.org/gdm85/go-distrilock/cli	[no test files]
?   	bitbucket.org/gdm85/go-distrilock/cli/distrilock	[no test files]
?   	bitbucket.org/gdm85/go-distrilock/api	[no test files]
BenchmarkLocksTaking/TCP_clients_suite_random_locks-4         	    5000	    238941 ns/op
BenchmarkLocksTaking/TCP_clients_suite_fixed_locks-4          	   10000	    470427 ns/op
BenchmarkLocksTaking/Websockets_binary_clients_suite_random_locks-4         	    3000	    496537 ns/op
BenchmarkLocksTaking/Websockets_binary_clients_suite_fixed_locks-4          	    3000	    488305 ns/op
BenchmarkLocksTaking/Websockets_text_clients_suite_random_locks-4           	    5000	    307426 ns/op
BenchmarkLocksTaking/Websockets_text_clients_suite_fixed_locks-4            	    5000	    299666 ns/op
PASS
ok  	bitbucket.org/gdm85/go-distrilock/api/client	12.126s
?   	bitbucket.org/gdm85/go-distrilock/api/core	[no test files]
?   	bitbucket.org/gdm85/go-distrilock/api/client/tcp	[no test files]
?   	bitbucket.org/gdm85/go-distrilock/cli/distrilock-ws	[no test files]
?   	bitbucket.org/gdm85/go-distrilock/api/client/ws	[no test files]
```

## Other Makefile targets

* **codeqa** runs various code quality measurements like `go vet`, `golint` and `errcheck`.
* **godoc** runs a local godoc HTTP server to explore package documentation
* **godoc-static** will store locally in `docs/` directory the HTML files for godoc package documentation

## How to use

Use one of the available daemons:
```bash
$ bin/distrilock --help
Usage: distrilock [--address=:13123] [--directory=.]
$ bin/distrilock-ws --help
Usage: distrilock [--address=:13124] [--directory=.]
```

Two deamons can point to the same directory - even across hosts, if using NFSv4 - if the operative system is POSIX compliant.

### Client side

Three types of clients are available:

* **TCP**, only with `bin/distrilock`
* **Websockets** with binary messages, only with `bin/distrilock-ws`
* **Websockets** with text (JSON) messages, only with `bin/distrilock-ws`

Look at the available tests for usage examples.

## FAQ

### Is the client concurrency-safe?

No. One TCP/websocket client corresponds to one session, which in turn corresponds to one connection; TCP/websocket connections are not safe to use across goroutines.

If you need a concurrency-safe client, use the provided wrapper client in `client/concurrent` as in:
```go
	// create a regular client
	c := tcp.New(addr, time.Second*3, time.Second*3, time.Second*3)

	// wrap it
	c = concurrent.New(c)
```

The concurrency wrapper simply adds a `sync.Mutex` lock/unlock before each `client.Client` method.

### How can I implement something like leases or expiration context?

Don't. If connection with the daemon is interrupted, you also have to stop assuming that the lock that was being used is still granted to your client.

For an usage pattern sensible to network interruptions, an implementation like the following is advised:
```go
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

     // start doing some intensive work
     for {
		///
		/// ... do some heavy work here, then iterate for some more heavy work
		///
		if completed {
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
```

## Other possible improvements

* the internal map sports a `sync.RWMutex` that optimizes reads; however, read optimizations are only effective if you have a high number of collisions against the same daemon instance; an option to disable RLock could be provided for the rest of scenarios
* `F_SETLKW` for waiting (and thus queue buildup) could be implemented, although it might need some thread trickery for the use of signals which would not be trivial in Go

## License

[GNU GPLv2](./LICENSE)

## Relevant links

* [File Locking and Unlocking with Fcntl](http://voyager.deanza.edu/~perry/lock.html), a nice summary on file locking with `fcntl`, also [available here in markdown format](./c_examples.md)
* [Linux flock utlity](https://github.com/karelzak/util-linux/blob/master/sys-utils/flock.c)
* [Advanced Linux Programming: fcntl: Locks and Other File Operations](http://www.informit.com/articles/article.aspx?p=23618&seqNum=4)
* [How to do distributed locking](https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html), lengthy read, but explains some common pitfalls of distributed locking implementations (and their usage)

## Similar software

There is a plethora of similar software, usually more sophisticated than distrilock:

* https://github.com/lomik/elock
* https://github.com/komarov/switchman
* https://redis.io/topics/distlock
