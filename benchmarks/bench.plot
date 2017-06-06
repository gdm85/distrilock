#!/bin/bash

if [ ! $# -eq 3 ]; then
	echo "Usage: bench.plot title input.dat output.png" 1>&2
	exit 1
fi

TITLE="$1"
INPUT="$2"
OUTPUT="$3"

MIN_Y=0.5
MAX_Y=1.6
PLOT_Y="[$MIN_Y:$MAX_Y]"
#PLOT_Y="[]"

if [ -z "$TITLE" ]; then
	TITLE=notitle
else
	TITLE="title '$TITLE'"
fi

cat<<EOF | gnuplot

set terminal svg font "Tahoma,12"
set output '$OUTPUT'

set xtics rotate # crucial line

#set bars fullwidth
set xtics format " "
set ylabel "ms"
set xlabel "client type"

plot [0:7]$PLOT_Y '$INPUT' using 1:3:4 with errorbars $TITLE, \
	'' using 1:3:2 with labels offset 2,0.5 notitle;
EOF
