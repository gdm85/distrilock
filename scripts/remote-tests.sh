#!/bin/bash
## remote-tests.sh
## @author gdm85
##
## Run tests by using two identical client and server hosts.
##
#

if [ ! $# -eq 3 ]; then
	echo "Usage: remote-tests.sh user server-host client-host" 1>&2
	exit 1
fi

USR="$1"
SERVER="$2"
CLIENT="$3"

set -e
set -x
## directory to create remotely
TMPD=$(mktemp --dry-run /tmp/gds.XXXXXXXX)

## build binary with benchmark tests
go test -c -o bin/distrilock-tests -bench=BenchmarkLocks -benchtime=1s ./api/client || exit $?

## prepare server
ssh $USR@$SERVER "mkdir -p /home/$USR/bin && mkdir $TMPD"
rsync -v bin/distrilock $USR@$SERVER:/home/$USR/bin/distrilock
rsync -v bin/distrilock-ws $USR@$SERVER:/home/$USR/bin/distrilock-ws

## prepare client
ssh $USR@$CLIENT mkdir -p /home/$USR/bin
rsync -v bin/distrilock-tests $USR@$CLIENT:/home/$USR/bin/distrilock-tests

###
### tcp daemon
###
SVC=distrilock
BASE=63419

ssh $USR@$SERVER bin/$SVC --address=:$[BASE+2] --directory="$TMPD" &
C=$!

###
### websocket daemons
###
SVC=distrilock-ws
BASE=63519

ssh $USR@$SERVER bin/$SVC --address=localhost:$[BASE+2] --directory="$TMPD" &
F=$!

trap "kill $C $F" EXIT

if [ -z "$TIMES" ]; then
	TIMES=1
fi

echo "Running all remote tests"
set +e
while [ $TIMES -gt 0 ]; do
	ssh $USR@$SERVER bin/distrilock-tests
	let TIMES-=1
done

exit 0
