#!/bin/bash
## remote-tests.sh
## @author gdm85
##
## Run tests by using two identical client and server hosts.
##
#

if [ ! $# -eq 3 -a ! $# -eq 2 ]; then
	echo "Usage: remote-tests.sh user server-host [client-host]" 1>&2
	exit 1
fi

USR="$1"
SERVER="$2"
CLIENT="$3"

set -e

## directory to create remotely
TMPD=$(mktemp --dry-run /tmp/gds.XXXXXXXX)

## build binary with benchmark tests
go test -c -o bin/distrilock-tests ./api/client || exit $?

## prepare server
ssh $USR@$SERVER "mkdir -p /home/$USR/bin && mkdir $TMPD"

if [ ! -z "$CLIENT" ]; then
	## prepare client
	ssh $USR@$CLIENT mkdir -p /home/$USR/bin
	CLIENT_RSYNC="rsync -v bin/distrilock-tests $USR@$CLIENT:/home/$USR/bin/distrilock-tests"
	echo "Running all remote tests with a remote client"
	CLIENT_PREFIX="ssh $USR@$CLIENT REMOTE_SERVER=$SERVER"
else
	echo "Running all remote tests with a local client"
	CLIENT_RESYNC=echo

	export REMOTE_SERVER=$SERVER
fi

## run all rsync in parallel
coshell << EOF
rsync -v bin/distrilock $USR@$SERVER:/home/$USR/bin/distrilock
rsync -v bin/distrilock-ws $USR@$SERVER:/home/$USR/bin/distrilock-ws
$CLIENT_RSYNC
EOF

###
### tcp daemon
###
SVC=distrilock
BASE=63419

ssh -t -t $USR@$SERVER bin/$SVC --address=:$[BASE+2] --directory="$TMPD" &
C=$!

###
### websocket daemon
###
SVC=distrilock-ws
BASE=63519

ssh -t -t $USR@$SERVER bin/$SVC --address=:$[BASE+2] --directory="$TMPD" &
F=$!

trap "kill $C $F" EXIT

if [ -z "$TIMES" ]; then
	TIMES=1
fi

set +e
while [ $TIMES -gt 0 ]; do
	$CLIENT_PREFIX bin/distrilock-tests -test.bench=BenchmarkLocks -test.benchtime=1s -test.run XXX

	let TIMES-=1
done

exit 0
