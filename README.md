# Peerster

This is a project for the course "Decentralized Systems Engineering" (CS438) at EPFL, Fall 2018.

#### Authors:  
Pablo Pfister <pablo.pfister@epfl.ch>  
Riccardo Conti <riccardo.conti@epfl.ch>  
Raphael Madillo <raphael.madillo@epfl.ch>

---

## About

### Gossiping in Peerster
Gossip protocols are distributed exchange protocols for ​robust information exchange​, typically deployed on ​dynamic network topologies​, e.g, because nodes can join and leave the network, they are mobile, their connectivity varies, etc. Examples of applications are ad-hoc communication between self-driving cars, peer-to-peer networks that broadcast a TV
program, sensor nodes that detect fire hazard in remote areas. The way gossip protocols spread information resembles gossipping in real life: a rumor may be heard by many people, although they don’t hear it directly from the rumor initiator.
When a node joins a gossip protocol, it has the contact information (e.g., network address) of a few nodes it can send messages to. Additionally, when a node receives a message, it learns the address of the sender. As an example, node C learns the address of node A when it receives the message from A.

### Peerster Design
Each node in Peerster acts as a ​gossiper​, as depicted above, but also ​exposes an API to clients to allow them to send messages, list received messages etc. The client could, in principle, run either locally, on the same machine, or remotely - for this assignment, however, we consider local clients. The gossiper communicates with other peers on the gossipPort​, and with clients on the ​UIPort​.

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
- GUI allowing all interactions

#### Homework 3:
- File search by keywords
- GUI allowing file search, file download
- Blockchain with mining and fork handling to guarantee unique filenames

#### Project:
- identities on the blockchain
- Sign and encrypt messages
- Assets on the blockchain
- Sending shares of assets
- Propose votes and vote for assets

---
## Run it
### The gossiper
Navigate to the project directory in a terminal and type `go build`. Then type `./Peerster` with the following options to launch the gossiper.
- `-UIPort=XXXX`: Port for the UI client (default 8080)
- `-gossipAddr=ip:port`: ip:port for the gossiper (default 127.0.0.1:5000)
- `-name=XXXX`: Name of the gossiper
- `-peers=ip:port,ip:port,...`: Comma separated list of peers of the form ip:port
- `-rtimer=X`: Route rumors sending period in seconds, 0 to disable
- `-simple`: Run gossiper in simple broadcast mode is present
- `-noGUI`: If this flag is present, don't run the webserver serving the GUI

### The client
The client allows multiple interactions:
##### Messages
- Sending a broadcast message: `./client -UIPort=XXXX -msg=YYYYYY`
- Sending a private message to a peer: `./client -UIPort=XXXX -msg=YYYYYY -dest=peerName`
- Sending an encrypted private message to a peer: `./client -UIPort=XXXX -msg=YYYYYY -dest=peerName -encrypt`

##### File Sharing
- Indexing a file (the file must be in the \_SharedFiles folder): `./client -UIPort=XXXX -file=filename`
- Requesting a file to another peer: `./client -UIPort=XXXX -file=filename -dest=peerName -reqest=hashOfTheRequestedChunkOrMetafile`
- Search files in the network by providing some keywords and optionally a budget: `./client -UIPort=XXXX -keywords=key1,key2 [-budget=4]`

##### Blockchain
- Inserting a new identity in the blockchain: `./client -UIPort=XXXX -identity=YYYYYY`
- Create a new asset (being nodeA): `./client -UIPort=XXXX -asset=AssetName -amount=100 -dest=nodeA`
- Send asset's shares: `./client -UIPort=XXXX -asset=AssetName -amount=50 -dest=nodeB`
- Propose a vote regarding an asset that you own: `./client -UIPort=XXXX -question="Should fire the CEO?" -assetVote=AssetName`
- Vote on a proposed vote: `./client -UIPort=XXXX -question="Should fire the CEO?" -assetVote=AssetName -origin=nodeA -answer=true/false`

Navigate to the `/client` project's subdirectory in a terminal and type `go build`.

### The GUI
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
