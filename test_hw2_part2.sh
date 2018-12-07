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

    # Test small file
    FILENAME="2chunks.txt"
    METAHASH="3bbe464d4f594b30e823451fff26198d865fb256b041a1b1f114d400ff94a70c"
    RECONSTRUCTED_FILENAME=$(date +"%T")

    ./client/client -UIPort=10000 -file=$FILENAME
    sleep 4
    ./client/client -UIPort=10001 -file=$RECONSTRUCTED_FILENAME -dest=Node0 -request=$METAHASH
    sleep 8

    require_text_in_file "METAHASH: $METAHASH" "testOutputs/Node0.out" "1"
    require_text_in_file "DOWNLOADING metafile of $RECONSTRUCTED_FILENAME from Node0" "testOutputs/Node1.out" "2"
    require_text_in_file "DOWNLOADING $RECONSTRUCTED_FILENAME chunk [0-9] from Node0" "testOutputs/Node1.out" "3"
    require_text_in_file "RECONSTRUCTED file $RECONSTRUCTED_FILENAME" "testOutputs/Node1.out" "4"

    # Test large file
    FILENAME2="256chunks.txt"
    METAHASH2="4d0fbe1a000e3e579a25c19ab7f86eb894c4d21ee28524d818fb0ab52b9b63ec"
    RECONSTRUCTED_FILENAME2="$RECONSTRUCTED_FILENAME-large"

    ./client/client -UIPort=10000 -file=$FILENAME2
    sleep 4
    ./client/client -UIPort=10001 -file=$RECONSTRUCTED_FILENAME2 -dest=Node0 -request=$METAHASH2
    sleep 8

    require_text_in_file "METAHASH: $METAHASH2" "testOutputs/Node0.out" "5"
    require_text_in_file "DOWNLOADING metafile of $RECONSTRUCTED_FILENAME2 from Node0" "testOutputs/Node1.out" "6"
    require_text_in_file "DOWNLOADING $RECONSTRUCTED_FILENAME2 chunk [0-9] from Node0" "testOutputs/Node1.out" "7"
    require_text_in_file "RECONSTRUCTED file $RECONSTRUCTED_FILENAME2" "testOutputs/Node1.out" "8"


    pkill -f Peerster
    exit
fi
