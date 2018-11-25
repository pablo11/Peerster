#!/bin/bash

if [ -z "$1" ]
then
    echo "Test options (sh test_hw2_part1.sh test-xxx):
        --test-filesharing
    "
    exit
fi

if [ $1 != "--test-filesharing" ]
then

    echo "⚠️ Invalid test option."
    echo "Test options (sh test_hw2_part1.sh test-xxx):
        --test-filesharing
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

NB_NODES=2

# Launch ring of peersters peersters
for i in `seq 0 $(($NB_NODES - 1))`
do
    uiPort=$((10000 + $i))
    gossipPort=$((5000 + $i))
    name="Node$i"
    nextPeerPort=$(($(($i + 1)) % $NB_NODES + 5000))

    ./Peerster -UIPort=$uiPort -gossipAddr=127.0.0.1:$gossipPort -name=$name -peers=127.0.0.1:$nextPeerPort -rtimer=1 > testOutputs/$name.out &
done


sleep 4


################################################################################
# TEST ROUTING
################################################################################

if [ $1 = "--test-filesharing" ]
then
    echo "${MAGENTA}Testing File sharing${NC}"

    ./client/client -UIPort=10000 -file=120chunks.txt
    sleep 3
    ./client/client -UIPort=10001 -file=120chunks.txt -dest=Node0 -request=818b03ad01499483cec670451ded6feed2599cb070e965dacac2dd1f1706975b
    sleep 10


    pkill -f Peerster
    exit
fi
