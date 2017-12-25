# distrilock
Distributed locking for pennies.

distrilock is a daemon that leverages Linux [POSIX filesystem locks](http://pubs.opengroup.org/onlinepubs/9699919799/functions/fcntl.html) and is compatible with NFSv4 mountpoints.
It does not use anything else that a Linux filesystem to enforce locking with `fcntl`; locks persistence is tied 1:1 with connection persistence e.g. **no leases**.

A typical deployment would consist of one or more distrilock daemons running on different hosts and sharing the same directory as an NFSv4 mountpoint.

distrilock daemons run on the following ports depending on type of service:
* distrilock: port 40800 (TCP)
* distrilock-ws: port 40801 (HTTP websockets)

## Terminology and basic functionality

A **daemon** is a distrilock server-side daemon listening for incoming connections - either TCP or Websockets.

A **client** is the client connecting to a daemon; to each client corresponds one **session** when the connection is estabilished.

Each session can acquire one more **named locks**; the connection fundamental to the session is kept alive with TCP keep-alive functionality; if such connection is interrupted or closed, the lock is immediately released and lost.

## How (distributed) locking works

1. the daemon instance tracks multiple named locks per session (= connection), which correspond to an individual file
2. client requests lock acquisition through daemon, if viable a `fcntl` write lock is acquired on the corresponding file and lock is returned as successful
3. client performs some work within the locked context
4. client releases lock through daemon, the `fcntl` write lock is released on the open file which is then closed and deleted

## Limitations

The daemon is effectively limited by the maximum number of open file descriptors and TCP connections that can be held; one file descriptor for the lock and one for the TCP connection will be necessary at anytime.
An informational message is printed when the daemon process has maximum 1024 or less file descriptors available.

The client is affected only by the TCP connections and one file descriptor per TCP connection limitation.

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
?   	github.com/gdm85/distrilock/cli	[no test files]
?   	github.com/gdm85/distrilock/cli/distrilock	[no test files]
?   	github.com/gdm85/distrilock/api	[no test files]
ok  	github.com/gdm85/distrilock/api/client	0.017s
?   	github.com/gdm85/distrilock/api/core	[no test files]
?   	github.com/gdm85/distrilock/api/client/tcp	[no test files]
?   	github.com/gdm85/distrilock/cli/distrilock-ws	[no test files]
?   	github.com/gdm85/distrilock/api/client/ws	[no test files]
```

The following binaries will be available in `bin/` directory:
* `distrilock` (TCP daemon)
* `distrilock-ws` (Websockets daemon)

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
```

## Benchmarks

Default (no timewait reuse/recycling):
```
BenchmarkLocksTaking/TCP_clients_suite-4         	   10000	    525924 ns/op
BenchmarkLocksTaking/Websockets_binary_clients_suite-4         	    3000	    706655 ns/op
BenchmarkLocksTaking/Websockets_text_clients_suite-4           	    5000	    462627 ns/op
BenchmarkLocksTaking/TCP_clients_suite_concurrency-safe-4      	    1000	   2443507 ns/op
BenchmarkLocksTaking/Websockets_binary_clients_suite_concurrency-safe-4         	    3000	    710488 ns/op
BenchmarkLocksTaking/Websockets_text_clients_suite_concurrency-safe-4           	    5000	    518831 ns/op
PASS
```

With TCP timewait recycling `sysctl -w net.ipv4.tcp_tw_recycle=1`:
```
BenchmarkLocksTaking/TCP_clients_suite-4         	   10000	    302553 ns/op
BenchmarkLocksTaking/Websockets_binary_clients_suite-4         	    3000	    689159 ns/op
BenchmarkLocksTaking/Websockets_text_clients_suite-4           	    5000	    425821 ns/op
BenchmarkLocksTaking/TCP_clients_suite_concurrency-safe-4      	   10000	    309075 ns/op
BenchmarkLocksTaking/Websockets_binary_clients_suite_concurrency-safe-4         	    3000	    683702 ns/op
BenchmarkLocksTaking/Websockets_text_clients_suite_concurrency-safe-4           	    5000	    433064 ns/op
PASS
```

With TCP timewait reuse `sysctl -w net.ipv4.tcp_tw_reuse=1`:
```
BenchmarkLocksTaking/TCP_clients_suite-4         	   10000	    309728 ns/op
BenchmarkLocksTaking/Websockets_binary_clients_suite-4         	    3000	    689373 ns/op
BenchmarkLocksTaking/Websockets_text_clients_suite-4           	    5000	    439981 ns/op
BenchmarkLocksTaking/TCP_clients_suite_concurrency-safe-4      	   10000	    307191 ns/op
BenchmarkLocksTaking/Websockets_binary_clients_suite_concurrency-safe-4         	    3000	    740121 ns/op
BenchmarkLocksTaking/Websockets_text_clients_suite_concurrency-safe-4           	    5000	    450616 ns/op
PASS
```

**NOTE**: these benchmarks run on same host, thus they do not correspond to a realistic scenario where the daemon would be running on a separate host.

See blog post for more accurate benchmarks.

## Other Makefile targets

* **godoc** runs a local godoc HTTP server to explore package documentation.
* **godoc-static** will store locally in `docs/` directory the HTML files for godoc package documentation.
* **codeqa** runs various code quality measurements like `go vet`, `golint` and `errcheck`.
* **simplify** formats and simplifies Go source of this project.
* **docker-image** builds a docker image for distrilock and distrilock-ws

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

If you wish to use a client in a concurrency-safe fashion, wrap it with `concurrent.New`; this would allow to save the time of the TCP connection setup and re-use the connection.

A minimal example is available in [example/main.go](./example/main.go).

## FAQ

### Is the TCP/Websocket client concurrency-safe?

No. One TCP/websocket client corresponds to one session, which in turn corresponds to one connection; TCP/websocket connections are not safe to use across goroutines.

If you need a concurrency-safe client, use the provided wrapper client in `client/concurrent` as in:
```go
	// create a regular client
	c := tcp.New(addr, time.Second*3, time.Second*3, time.Second*3)

	// wrap it
	c = concurrent.New(c)
```

The concurrency wrapper simply adds a `sync.Mutex` lock/unlock before each method of the `client.Client` interface.

## Shall I use one client for all locks or one client for each lock?

It matters only if you plan to acquire a lot of locks from a single process. The server-side lock acquisition bottleneck will always be there regardless of what type of client you use.
When creating many clients, keep in mind TCP connections limits and file descriptor limits.

### How can I implement something like leases or expiration context?

Don't. If connection with the daemon is interrupted, you also have to stop assuming that the lock that was being used is still granted to your client.

For an usage pattern sensible to network interruptions, see the use of `Verify()` in the provided example [example/main.go](./example/main.go).

### What about a gRPC client/server?

[Go gRPC clients](https://github.com/grpc/grpc-go/) have complex retry policies and generally cannot satisfy the persistence requirement.

## Other possible improvements

* the internal map sports a `sync.RWMutex` that optimizes reads; however, read optimizations are only effective if you have a high number of collisions against the same daemon instance; an option to disable RLock could be provided for the rest of scenarios
* `F_SETLKW` for waiting (and thus queue buildup) could be implemented, although it might need some thread trickery for the use of signals which would not be trivial in Go
* TCP-level improvements: possibility to use `SO_REUSEADDR` when connecting to a TCP/Websockets daemon (currently not possible in Go: https://github.com/golang/go/issues/9661)
* TCP-level improvements: `SO_FASTOPEN` support

## License

[GNU GPLv2](./LICENSE)

## Credits

Thanks to the people that implemented POSIX locks in Linux and NFSv4; the `man 2 fcntl` page and the online sources that you can find in 'Relevant links' section were helpful in understanding and prototyping.

## Relevant links

* [File Locking and Unlocking with Fcntl](http://voyager.deanza.edu/~perry/lock.html), a nice summary on file locking with `fcntl`, also [available here in markdown format](./c_examples.md)
* [Linux flock utlity](https://github.com/karelzak/util-linux/blob/master/sys-utils/flock.c)
* [Advanced Linux Programming: fcntl: Locks and Other File Operations](http://www.informit.com/articles/article.aspx?p=23618&seqNum=4)
* [How to do distributed locking](https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html), lengthy read, but explains some common pitfalls of distributed locking implementations (and their usage)

## Similar software

There is a plethora of similar but more sophisticated software:

* https://coreos.com/blog/etcd3-a-new-etcd.html - etcd3 supposedly supports solid distributed locking, according to the release news
* [The Redlock algorithm](https://redis.io/topics/distlock)
