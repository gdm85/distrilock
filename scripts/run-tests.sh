#!/bin/bash

set -e

TMPD="$(mktemp -d)"

bin/distrilock --address=:63419 --directory="$TMPD" &
A=$!

## server process B
bin/distrilock --address=:63420 --directory="$TMPD" &
B=$!

if [ ! -z "$NFS_SHARE" ]; then
	## server process C, NFS share
	bin/distrilock --address=:63421 --directory="$NFS_SHARE" &
	C=$!
	OPTS=""
else
	OPTS="-short"
fi

trap "kill $A $B $C; rm -rf '$TMPD'" EXIT

go test $OPTS -bench=. -benchtime=0.2s "$@"
