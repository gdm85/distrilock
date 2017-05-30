#!/bin/bash
## run-tests.sh
## @author gdm85
##
## Run tests by using different local and remote daemons.
##
#

set -e

TMPD="$(mktemp -d)"

###
### tcp daemons
###
SVC=distrilock
BASE=63419

## local daemon A
bin/$SVC --address=:$BASE --directory="$TMPD" &
A=$!

## local daemon B
bin/$SVC --address=:$[BASE+1] --directory="$TMPD" &
B=$!

if [ ! -z "$NFS_SHARE" ]; then
	## local daemon C, on an NFS share
	bin/$SVC --address=:$[BASE+2] --directory="$NFS_SHARE" &
	C=$!
	OPTS=""
else
	OPTS="-short"
fi

###
### websocket daemons
###
SVC=distrilock-ws
BASE=63519

## local daemon A
bin/$SVC --address=localhost:$BASE --directory="$TMPD" &
D=$!

## local daemon B
bin/$SVC --address=localhost:$[BASE+1] --directory="$TMPD" &
E=$!

if [ ! -z "$NFS_SHARE" ]; then
	## local daemon C, on an NFS share
	bin/$SVC --address=localhost:$[BASE+2] --directory="$NFS_SHARE" &
	F=$!
fi

trap "kill $A $B $C $D $E $F; rm -rf '$TMPD'" EXIT

if [ -z "$TIMES" ]; then
	TIMES=1
fi

echo "Running all tests"
set +e
while [ $TIMES -gt 0 ]; do
	LOCAL_LOCK_DIR="$TMPD" go test $OPTS "$@" || exit $?

	let TIMES-=1
done

exit 0
