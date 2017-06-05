#!/bin/bash
## remote-tests.sh
## @author gdm85
##
## Run tests by using two identical client and server hosts.
##
#

if [ ! $# -eq 3 -a ! $# -eq 2 ]; then
	echo "Usage: remote-tests.sh user server-host [client-host]" 1>&2
	echo "	REMOTE_SERVER can be used to provide a local address for the remote server" 1>&2
	echo "	BENCH_TIME can be used to specify test.benchtime (default 1s)" 1>&2
	echo "	BENCH_REGEX can be used to specify test.bench" 1>&2
	exit 1
fi

USR="$1"
SERVER="$2"
CLIENT="$3"

if [ -z "$REMOTE_SERVER" ]; then
	REMOTE_SERVER="$SERVER"
fi

if [ -z "$BENCH_TIME" ]; then
	BENCH_TIME=1s
fi

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
	CLIENT_RSYNC="rsync bin/distrilock-tests $USR@$CLIENT:/home/$USR/bin/distrilock-tests"
	echo "Running all remote tests with a remote client"
	CLIENT_PREFIX="ssh $USR@$CLIENT REMOTE_SERVER=$REMOTE_SERVER"
else
	echo "Running all remote tests with a local client"
	CLIENT_RESYNC=echo

	export REMOTE_SERVER=$REMOTE_SERVER
fi

## run all rsync in parallel
coshell << EOF
rsync bin/distrilock $USR@$SERVER:/home/$USR/bin/distrilock
rsync bin/distrilock-ws $USR@$SERVER:/home/$USR/bin/distrilock-ws
$CLIENT_RSYNC
EOF

###
### tcp daemon
###
SVC=distrilock
PORT=63422

ssh -t -t $USR@$SERVER "bin/$SVC --address=:$PORT --directory=$TMPD" &
C=$!

###
### websocket daemon
###
SVC=distrilock-ws
WSPORT=63522

ssh -t -t $USR@$SERVER "bin/$SVC --address=:$WSPORT --directory=$TMPD" &
F=$!

trap "kill $C $F" EXIT

function test_ports() {
	nc -z -w5 $SERVER $PORT && \
	nc -z -w5 $SERVER $WSPORT
}

set +e

echo "Waiting for daemons to start listening"
FAILS=0
while ! test_ports; do
	let FAILS+=1
	if [ $FAILS -eq 5 ]; then
		echo "ERROR: could not see ports open" 1>&2
		exit 1
	fi
	sleep 1
done

if [ -z "$TIMES" ]; then
	TIMES=1
fi

while [ $TIMES -gt 0 ]; do
	$CLIENT_PREFIX bin/distrilock-tests -test.bench=$BENCH_REGEX -test.benchtime=$BENCH_TIME -test.run XXX

	let TIMES-=1
done

exit 0
