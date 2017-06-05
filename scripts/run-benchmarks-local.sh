#!/bin/bash

set -e

function collect_benchmark() {
	local STORE TITLE
	STORE="$1"
	TITLE="$2"

	rm -f benchmarks/benchstats.txt
	make benchmarks/benchstats.txt
	cat benchmarks/benchstats.txt >> benchmarks/benchstats-${STORE}.txt
	cp benchmarks/benchstats-${STORE}.txt benchmarks/benchstats.txt
	TITLE="$TITLE" make benchmark-plot
	mv benchmarks/locks.svg benchmarks/${STORE}.svg
}

sudo sysctl -w net.ipv4.tcp_tw_reuse=0
sudo sysctl -w net.ipv4.tcp_tw_recycle=0
collect_benchmark no-tw-rr "No time-wait reuse/recycle"

sudo sysctl -w net.ipv4.tcp_tw_recycle=1
collect_benchmark tw-recycle "With time-wait recycle"
sudo sysctl -w net.ipv4.tcp_tw_recycle=0

sudo sysctl -w net.ipv4.tcp_tw_reuse=1
collect_benchmark tw-reuse "With time-wait reuse"
sudo sysctl -w net.ipv4.tcp_tw_reuse=0

sudo sysctl -w net.ipv4.tcp_tw_recycle=1
sudo sysctl -w net.ipv4.tcp_tw_reuse=1
collect_benchmark tw-rr "With time-wait recycle and reuse"
sudo sysctl -w net.ipv4.tcp_tw_recycle=0
sudo sysctl -w net.ipv4.tcp_tw_reuse=0
