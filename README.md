# Peerster

## CS438 - ecentralized Systems Engineering

Fall 2018

Authors: Pablo Pfister

### Run it

go build
./Peerster -UIPort=10001 -gossipAddr=127.0.0.1:5001 -peers=127.0.0.1:5002 -name=nodeA -simple
./Peerster -UIPort=10002 -gossipAddr=127.0.0.1:5002 -peers=127.0.0.1:5001 -name=nodeB -simple

or

go install
Peerster

### Test

Just write in the terminal:
echo -n "hello" >/dev/udp/localhost/5000

If Peerster is running it will handle it



### Run the webserver
./webserver -UIPort=8080 -gossipAddr=127.0.0.1:5001 -peers=127.0.0.1:5002 -name=nodeA -simple
./webserver -UIPort=8081 -gossipAddr=127.0.0.1:5002 -peers=127.0.0.1:5001 -name=nodeB -simple


### Run the client
./client -UIPort=10001 -msg=hello
./client -UIPort=10002 -msg=hello
