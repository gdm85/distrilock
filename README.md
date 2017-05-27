# distrilock
Distributed locking for pennies.

distrilock is a TCP and Websockets daemon that leverages Linux [POSIX filesystem locks](http://pubs.opengroup.org/onlinepubs/9699919799/functions/fcntl.html) and is compatible with NFSv4 mountpoints.
It does not use anything else that a Linux filesystem to enforce locking with `fcntl`; locks persistence is tied 1:1 with connection persistence e.g. **no leases**.

A typical deployment would consist of multiple distrilock daemons using the same directory, optionally shared over NFSv4 for redundancy.

## Terminology and basic functionality

A **daemon** is a distrilock server-side daemon listening for incoming connections - either TCP or Websockets.

A **client** is the client connecting to a daemon; to each client corresponds one **session** when the connection is estabilished.

Each session can acquire one more **named locks**; the connection fundamental to the session is kept alive with TCP keep-alive functionality; if such connection is interrupted or closed, the lock is immediately released and lost.

## How (distributed) locking works

1. the daemon instance tracks multiple named locks per session (= connection), which correspond each to an individual file
2. client requests lock acquisition through daemon, if viable a fcntl write lock is acquired on the corresponding file and lock is returned as successful
3. client perform some work within the locked context
4. client releases lock through daemon, the fcntl write lock is released on the open file which is closed and deleted

## NFSv4 specific notes

The default `local_lock=none` must be part of your mountpoint options; [version 4 of the NFS protocol](https://tools.ietf.org/html/rfc7530) is necessary.

## Building

After cloning the repository and putting it in an independent GOPATH:
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

## Relevant links

* [Linux flock utlity](https://github.com/karelzak/util-linux/blob/master/sys-utils/flock.c)
* [Advanced Linux Programming: fcntl: Locks and Other File Operations](http://www.informit.com/articles/article.aspx?p=23618&seqNum=4)
* [How to do distributed locking](https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html), lengthy read, but explains some common pitfalls of distributed locking implementations (and their usage)

## Similar software

Similar software (not as low level as distrilock though):
* https://github.com/lomik/elock
* https://github.com/komarov/switchman
* https://redis.io/topics/distlock
