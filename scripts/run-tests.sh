#!/bin/bash

set -e

TMPD="$(mktemp -d)"

bin/distrilock --address=:63419 --directory="$TMPD" &
A=$!

## server process B
bin/distrilock --address=:63420 --directory="$TMPD" &
B=$!

trap "kill $A $B; rm -rf '$TMPD'" EXIT

go test "$@"
