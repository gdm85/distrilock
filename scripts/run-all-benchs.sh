#!/bin/bash

set -e

export BENCH_TIME=1s
export REMOTE_SERVER=172.31.40.95
export TIMES=1

REMOTE_ARGS="ubuntu 52.212.231.41 52.51.61.99"

BENCH_REGEX=BenchmarkSuiteAcquireAndRelease exec scripts/remote-tests.sh $REMOTE_ARGS | tee benchmarks/benchstats.txt

benchstat benchmarks/benchstats.txt | tail -n+2 | go run benchmarks/conv-data.go > benchmarks/benchstats.dat

benchmarks/bench.plot "" benchmarks/benchstats.dat benchmarks/acquire-release.svg


BENCH_REGEX=BenchmarkSuiteInitialConn exec scripts/remote-tests.sh $REMOTE_ARGS | tee benchmarks/benchstats.txt

benchstat benchmarks/benchstats.txt | tail -n+2 | go run benchmarks/conv-data.go > benchmarks/benchstats.dat

benchmarks/bench.plot "" benchmarks/benchstats.dat benchmarks/initial-conn.svg
