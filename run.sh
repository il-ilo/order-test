B='xxx.servicebus.windows.net:9093' # change me
PASS=''

TOPIC=""
N=1000000

go build . 
./partition-order-check -n $N -t $TOPIC -b $B -c $PASS -p 0
