#!/usr/bin/env gnuplot

set terminal pngcairo
set output 'benchmarks/locks.png'

set xtics rotate # crucial line

#set size 0.6, 0.6

#set bars fullwidth
set xtics format " "
set ylabel "ms"

set terminal png font "Tahoma" 14

plot [0:7][] 'benchmarks/benchstats.dat' using 1:3:4 with errorbars notitle, \
	'' using 1:3:2 with labels offset 0.5,1 notitle;
