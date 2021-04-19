#!/bin/bash

trap "rm main;kill 0" EXIT
go build -o main
./main -port=10086 -apiport=8091 &
./main -port=10087
./main -port=10088
wait


