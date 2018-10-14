# Peerster
---

## About
This is a project for the course "Decentralized Systems Engineering" (CS438) at EPFL, Fall 2018.

Author: Pablo Pfister <pablo.pfister@epfl.ch>


## Contents

#### Homework 1:
- Simple messages
- Rumor mongering and anti-entropy
- Simple GUI

## Run it
#### The gossiper
Navigate to the project directory in a terminal and type `go build`. Then type `./Peerster` to launch the gossiper (see homework 1 handout for the options of this command).

#### The client
Navigate to the `/client` project's subdirectory in a terminal and type `go build`. Then type `./client -UIPort=XXXX -msg=YYYYYY` to launch the client, where `XXXX` is to be replaced with the port your gossiper is listening for the client and `YYYYYY` is to be replaced with the message you want to send.

#### The GUI
To simplify the usage of this software, one unique command allows to launch a gossiper and a webserver serving the GUI.

Navigate to the `/webserver` project's subdirectory in a terminal and type `go build`. Then type `./webserver` to launch the client, 



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
