#!/bin/bash

set -e

FIRST_ARG="$1"

export BENCH_TIME=10s
export REMOTE_SERVER=172.31.40.95
export TIMES=3

SERVERS="34.250.254.189 52.51.251.248"
REMOTE_ARGS="ubuntu $SERVERS"

function collect_benchmark() {
	local STORE TITLE
	STORE="$1"
	TITLE="$2"

	if [ "$FIRST_ARG" != "--plot-only" ]; then
		BENCH_REGEX=BenchmarkSuiteAcquireAndRelease exec scripts/remote-tests.sh $REMOTE_ARGS | tee benchmarks/${STORE}.txt
		benchstat benchmarks/${STORE}.txt | tail -n+2 | go run benchmarks/conv-data.go > benchmarks/${STORE}.dat
	fi
	benchmarks/bench.plot "$TITLE" benchmarks/${STORE}.dat benchmarks/${STORE}.svg
}

function remote_set() {
	if [ "$FIRST_ARG" != "--plot-only" ]; then
		for HOST in $SERVERS; do
			echo "ssh ubuntu@$HOST sudo sysctl -w $@"
		done | coshell --deinterlace
	fi
}

#BENCH_REGEX=BenchmarkSuiteInitialConn exec scripts/remote-tests.sh $REMOTE_ARGS | tee benchmarks/benchstats.txt
#benchstat benchmarks/benchstats.txt | tail -n+2 | go run benchmarks/conv-data.go > benchmarks/benchstats.dat
#benchmarks/bench.plot "" benchmarks/benchstats.dat benchmarks/initial-conn.svg

remote_set net.ipv4.tcp_tw_reuse=0 net.ipv4.tcp_tw_recycle=0
collect_benchmark no-tw-rr "No time-wait reuse/recycle"

remote_set net.ipv4.tcp_tw_reuse=0 net.ipv4.tcp_tw_recycle=1
collect_benchmark tw-recycle "With time-wait recycle"

remote_set net.ipv4.tcp_tw_reuse=1 net.ipv4.tcp_tw_recycle=0
collect_benchmark tw-reuse "With time-wait reuse"

remote_set net.ipv4.tcp_tw_reuse=1 net.ipv4.tcp_tw_recycle=1
collect_benchmark tw-rr "With time-wait recycle and reuse"

## final cleanup
remote_set net.ipv4.tcp_tw_reuse=0 net.ipv4.tcp_tw_recycle=0
