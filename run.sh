#!/bin/bash

set -e

B=''
PASS=''
TOPIC=""
N=1000000
P=0


go build . 
./partition-order-check -n $N -t $TOPIC -b $B -c $PASS -p $P
