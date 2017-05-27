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
scripts/run-tests.sh: line 1: 32663 Terminated              bin/$SVC --address=localhost:$BASE --directory="$TMPD"
scripts/run-tests.sh: line 1: 32664 Terminated              bin/$SVC --address=localhost:$[BASE+1] --directory="$TMPD"
```

## Relevant links

* [Linux flock utlity](https://github.com/karelzak/util-linux/blob/master/sys-utils/flock.c)
* [Advanced Linux Programming: fcntl: Locks and Other File Operations](http://www.informit.com/articles/article.aspx?p=23618&seqNum=4)
* [How to do distributed locking](https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html), lengthy read, but explains some common pitfalls of distributed locking implementations (and their usage)

## Similar software

Similar software (not as low level as distrilock though):
* https://github.com/lomik/elock
* https://github.com/komarov/switchman
* https://redis.io/topics/distlock
