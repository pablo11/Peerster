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

#### Homework 2:
- Routing
- Private messages
- File sharing (indexing, requesting, sending and downloading)


---
## Run it
#### The gossiper
Navigate to the project directory in a terminal and type `go build`. Then type `./Peerster` to launch the gossiper (see homework 1 handout for the options of this command).

#### The client
The client allows multiple interactions:
- Sending a broadcast message
- Sending a private message to a peer
- Indexing a file (the file must be in the \_SharedFiles folder)
- Requesting a file to another peer

Navigate to the `/client` project's subdirectory in a terminal and type `go build`.

The four features described above can be used with the following commands:
`./client -UIPort=XXXX -msg=YYYYYY`
`./client -UIPort=XXXX -msg=YYYYYY -dest=peerName`
`./client -UIPort=XXXX -file=filename`
`./client -UIPort=XXXX -file=filename -dest=peerName -reqest=hashOfTheRequestedChunkOrMetafile`

#### The GUI
The GUI is served by default by this implementation of Peerster on startup.
To see the GUI simply open a browser window and go at `127.0.0.1:UIPort`, where `UIPort` is the UIPort option (default 8080).


---
## Examples

#### The gossiper
`go build`
`./Peerster -rtimer=2 -UIPort=10001 -gossipAddr=127.0.0.1:5001 -peers=127.0.0.1:5002 -name=nodeA -noGUI`
`./Peerster -rtimer=2 -UIPort=10002 -gossipAddr=127.0.0.1:5002 -peers=127.0.0.1:5001 -name=nodeB -noGUI`

Chain of 3 peers A<->B<->C
`./Peerster -rtimer=2 -name=nodeA -UIPort=10001 -gossipAddr=127.0.0.1:5001 -peers=127.0.0.1:5002`
`./Peerster -rtimer=2 -name=nodeB -UIPort=10002 -gossipAddr=127.0.0.1:5002 -peers=127.0.0.1:5001,127.0.0.1:5003`
`./Peerster -rtimer=2 -name=nodeC -UIPort=10003 -gossipAddr=127.0.0.1:5003 -peers=127.0.0.1:5002`

#### The client
`./client -UIPort=10001 -msg=hello`
`./client -UIPort=10002 -file=two.txt -dest=nodeA -request=3bbe464d4f594b30e823451fff26198d865fb256b041a1b1f114d400ff94a70c`
`./client -UIPort=10001 -file=2chunks.txt`

# TODO
- Add to the GUI the support for file upload and download
- Add checks for to big filesize in the indexing
- Add the timeout of 5 sec if a data request is not replied and send the request again
- Add support for huge files (don't know how to doit but maybe storing on file chunks and metafiles instead of in memory)
- Remove chunks after file reconstruction
