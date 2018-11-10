# Peerster
---

## About
This is a project for the course "Decentralized Systems Engineering" (CS438) at EPFL, Fall 2018.

Author: Pablo Pfister <pablo.pfister@epfl.ch>

---
## Contents

#### Homework 1:
- Simple messages
- Rumor mongering and anti-entropy
- Simple GUI

---
## Run it
#### The gossiper
Navigate to the project directory in a terminal and type `go build`. Then type `./Peerster` to launch the gossiper (see homework 1 handout for the options of this command).

#### The client
Navigate to the `/client` project's subdirectory in a terminal and type `go build`. Then type `./client -UIPort=XXXX -msg=YYYYYY` to launch the client, where `XXXX` is to be replaced with the port your gossiper is listening for the client and `YYYYYY` is to be replaced with the message you want to send.

#### The GUI
The GUI is served by default by this implementation of Peerster on startup.
To see the GUI simply open a browser window and go at `127.0.0.1:UIPort`, where `UIPort` is the UIPort option (default 8080).


---
## Examples

#### The gossiper
`go build`
`./Peerster -UIPort=10001 -gossipAddr=127.0.0.1:5001 -peers=127.0.0.1:5002 -name=nodeA`
`./Peerster -UIPort=10002 -gossipAddr=127.0.0.1:5002 -peers=127.0.0.1:5001 -name=nodeB`

#### The client
`./client -UIPort=10001 -msg=hello`
