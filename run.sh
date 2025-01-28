#!/bin/bash

set -e

B=''
PASS=''
TOPIC=""
N=1000000
P=0

NC=100

go build . 
./ordchec  -n $N -t $TOPIC -b $B -c $PASS -p $P -col $NC
