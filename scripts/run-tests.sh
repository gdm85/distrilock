#!/bin/bash

set -e

TMPD="$(mktemp -d)"

bin/distrilock --address=:63419 --directory="$TMPD" &
P=$!

trap "kill $P; rm -rf '$TMPD'" EXIT

go test "$@"
