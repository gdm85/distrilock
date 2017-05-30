#!/bin/bash

set -e

cd docker
mv ../bin/distrilock ../bin/distrilock-ws .

trap 'mv distrilock distrilock-ws ../bin/' EXIT

docker build .
