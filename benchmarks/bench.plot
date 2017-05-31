#!/bin/bash

if [ ! $# -eq 3 ]; then
	echo "Usage: bench.plot title input.dat output.png" 1>&2
	exit 1
fi

TITLE="$1"
INPUT="$2"
OUTPUT="$3"

if [ -z "$TITLE" ]; then
	TITLE=notitle
else
	TITLE="title '$TITLE'"
fi

cat<<EOF | gnuplot

set terminal pngcairo
set output '$OUTPUT'

set xtics rotate # crucial line

#set size 0.6, 0.6

#set bars fullwidth
set xtics format " "
set ylabel "ms"
set xlabel "client type"

set terminal png font "Tahoma" 12

plot [0:7][] '$INPUT' using 1:3:4 with errorbars $TITLE, \
	'' using 1:3:2 with labels offset 2,0.5 notitle;
EOF
