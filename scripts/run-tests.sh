#!/bin/bash

set -e

TMPD="$(mktemp -d)"

## local daemon A
bin/distrilock --address=:63419 --directory="$TMPD" &
A=$!

## local daemon B
bin/distrilock --address=:63420 --directory="$TMPD" &
B=$!

if [ ! -z "$NFS_SHARE" ]; then
	## local daemon C, on an NFS share
	bin/distrilock --address=:63421 --directory="$NFS_SHARE" &
	C=$!
	OPTS=""
else
	OPTS="-short"
fi

trap "kill $A $B $C; rm -rf '$TMPD'" EXIT

echo "Running all tests"
go test $OPTS "$@"
