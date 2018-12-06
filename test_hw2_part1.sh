#!/bin/bash

if [ -z "$1" ]
then
    echo "Test options (sh test_hw2_part1.sh test-xxx):
        --test-routing
        --test-privateMessage
    "
    exit
fi

if [ $1 != "--test-routing" ] && [ $1 != "--test-privateMessage" ]
then

    echo "⚠️ Invalid test option."
    echo "Test options (sh test_hw2_part1.sh test-xxx):
        --test-routing
        --test-privateMessage
    "

    exit
fi


BLACK='\033[0;30m'
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
GRAY='\033[0;37m'
WHITE='\033[0;97m'
NC='\033[0m'

################################################################################
# Helpers
################################################################################

# $1 is the text to find in file $2, $3 is the number of the test
require_text_in_file() {
    if !(grep -Eq "$1" "$2")
    then
        # Test failed
        echo -e "${RED}***Test $3 FAILED***${NC}"
    else
        echo -e "${GREEN}***Test $3 PASSED***${NC}"
    fi
}

################################################################################
# Setup: build, clean and prepare output folder
################################################################################

go build
cd client
go build
cd ..

pkill -f Peerster
rm -rf testOutputs
mkdir testOutputs

NB_NODES=4

# Launch ring of peersters peersters
for i in `seq 0 $(($NB_NODES - 1))`
do
    uiPort=$((10000 + $i))
    gossipPort=$((5000 + $i))
    name="Node$i"
    nextPeerPort=$(($(($i + 1)) % $NB_NODES + 5000))

    ./Peerster -UIPort=$uiPort -gossipAddr=127.0.0.1:$gossipPort -name=$name -peers=127.0.0.1:$nextPeerPort -rtimer=2 > testOutputs/$name.out &
done


sleep 6


################################################################################
# TEST ROUTING
################################################################################

if [ $1 = "--test-routing" ]
then
    echo "${MAGENTA}Testing Routing${NC}"

    ./client/client -UIPort=10000 -msg=Hey_you,_come_on!
    sleep 6

    require_text_in_file "DSDV Node1 127.0.0.1:[0-9]{4}" "testOutputs/Node0.out" "1"
    require_text_in_file "DSDV Node3 127.0.0.1:[0-9]{4}" "testOutputs/Node0.out" "2"

    require_text_in_file "DSDV Node0 127.0.0.1:[0-9]{4}" "testOutputs/Node1.out" "3"
    require_text_in_file "DSDV Node2 127.0.0.1:[0-9]{4}" "testOutputs/Node1.out" "4"

    require_text_in_file "DSDV Node0 127.0.0.1:[0-9]{4}" "testOutputs/Node2.out" "5"
    require_text_in_file "DSDV Node1 127.0.0.1:[0-9]{4}" "testOutputs/Node2.out" "6"
    require_text_in_file "DSDV Node3 127.0.0.1:[0-9]{4}" "testOutputs/Node2.out" "7"

    require_text_in_file "DSDV Node0 127.0.0.1:[0-9]{4}" "testOutputs/Node3.out" "8"
    require_text_in_file "DSDV Node2 127.0.0.1:[0-9]{4}" "testOutputs/Node3.out" "9"

    pkill -f Peerster
    exit
fi



################################################################################
# TEST PRIVATE MESSAGE
################################################################################


if [ $1 = "--test-privateMessage" ]
then
    echo "${MAGENTA}Testing Private messages${NC}"

    ./client/client -UIPort=10000 -msg=Hey_you,_come_on2!
    sleep 4
    require_text_in_file "CLIENT MESSAGE Hey_you,_come_on2!" "testOutputs/Node0.out" "1"

    # Test sending private messages to directly connected nodes
    ./client/client -UIPort=10003 -msg=Hey_you,_come_on3! -dest=Node2
    sleep 4
    require_text_in_file "PRIVATE origin Node3 hop-limit [0-9]+ contents Hey_you,_come_on3!" "testOutputs/Node2.out" "2"

    # Test sending private messages to non directly connected nodes
    ./client/client -UIPort=10002 -msg=Hey_you,_come_on! -dest=Node0
    sleep 10
    require_text_in_file "PRIVATE origin Node2 hop-limit [0-9]+ contents Hey_you,_come_on!" "testOutputs/Node0.out" "3"

    pkill -f Peerster
    exit
fi
