package gossip

import (
    "fmt"
    "log"
    "net"
    "strings"
    "bytes"
    "sync"
    "math/rand"
    "time"
    "github.com/dedis/protobuf"
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/collections"
)

const PACKET_BUFFER_LEN int = 1024
const ANTI_ENTROPY_PERIOD time.Duration = 1 // Number of seconds to wait

type Gossiper struct {
    address *net.UDPAddr
    conn *net.UDPConn
    Name string
    peers []string
    simple bool

    // For rumoring
    nextMessageId uint32
    status map[string]*model.PeerStatus
    messages map[string][]*model.RumorMessage

    // Mutex to lock structures on modification
    peersMutex sync.Mutex
    nextMessageIdMutex sync.Mutex
    statusMutex sync.Mutex
    messagesMutex sync.Mutex
}

func NewGossiper(address, name string, peers []string, simple bool) *Gossiper {
    udpAddr := resolveAddress(address)
    udpConn, err := net.ListenUDP("udp4", udpAddr)
    if err != nil {
        log.Fatal(err)
    }

    return &Gossiper{
        address: udpAddr,
        conn: udpConn,
        Name: name,
        peers: peers,
        simple: simple,
        nextMessageId: 1,
        status: make(map[string]*model.PeerStatus),
        messages: make(map[string][]*model.RumorMessage),
        peersMutex: sync.Mutex{},
        nextMessageIdMutex: sync.Mutex{},
        statusMutex: sync.Mutex{},
        messagesMutex: sync.Mutex{},
    }
}

func (g *Gossiper) GetAddress() string {
    return g.address.String()
}

func (g *Gossiper) GetPeers() []string {
    return g.peers
}

func (g *Gossiper) ListenPeers() {
    for {
        packetBuffer := make([]byte, 2 * PACKET_BUFFER_LEN)
        _, fromAddr, err := g.conn.ReadFrom(packetBuffer)
        if err != nil {
            fmt.Println(err)
            continue
        }

        // Decode the message
        gp := model.GossipPacket{}
        err = protobuf.Decode(packetBuffer, &gp)
        if err != nil {
            //fmt.Println("ERROR:", err)
        }

        // Store addr in the list of peers if not already present
        // Could get address from: _, addr, _ := g.conn.ReadFrom(packetBuffer)
        g.AddPeer(fromAddr.String())

        switch {
            case gp.Simple != nil:
                g.printGossipPacket("peer", "", &gp)

                // Change the relay peer field to this node address
                receivedFrom := gp.Simple.RelayPeerAddr
                gp.Simple.RelayPeerAddr = g.address.String()

                // Broadcast the message to every peer except the one the message was received from
                go g.sendGossipPacket(&gp, collections.Filter(g.peers, func(p string) bool{
                    return p != receivedFrom
                }))

            case gp.Rumor != nil:
                g.printGossipPacket("received", fromAddr.String(), &gp)

                // If the message is the next one expected, store it
                if gp.Rumor.ID == g.getVectorClock(gp.Rumor.Origin) {
                    g.incrementVectorClock(gp.Rumor.Origin)
                    g.storeMessage(gp.Rumor)
                    g.sendRumorMessage(gp.Rumor, true, "")
                }

                // Send status message to the peer the rumor message was received from
                g.sendStatusMessage(fromAddr.String())

            case gp.Status != nil:
                g.printGossipPacket("", fromAddr.String(), &gp)

                sm := gp.Status

                // Check if the other peer has some messages that this gossiper doesn't have.
                // In this case send him a StatusMessage to ask those messages
                inSync := g.getOutOfSyncMessages(fromAddr.String(), sm.Want)
                if inSync {
                    fmt.Println("IN SYNC with " + fromAddr.String())
                    fmt.Println()
                }

                // Send to the other peer messages that he doesn't have
                go g.sendMissingMessagesToPeer(fromAddr.String(), sm.Want)

                // Notify "wait for acknowledgement" if a status is received from a peer we are waiting for
                // TODO

                // Flip coint to know if continouing the rumoring

            default:
                fmt.Println("WARNING: Unoknown message type")
        }

    }
}

func (g *Gossiper) ListenClient(uiPort string) {
    udpAddr := resolveAddress("127.0.0.1:" + uiPort)
    conn, err := net.ListenUDP("udp4", udpAddr)
    if err != nil {
        fmt.Println(err)
    }

    packetBuffer := make([]byte, PACKET_BUFFER_LEN)

    for {
        // TODO: clean packetBuffer
        _, _, err := conn.ReadFromUDP(packetBuffer)
        if err != nil {
            fmt.Println(err)
            continue
        }

        // Prepare contents removing unused bytes
        contents := string(bytes.Trim(packetBuffer, "\x00"))

        if g.simple {
            go g.SendSimpleMessage(contents)
        } else {
            // Build RumorMessage
            rm := &model.RumorMessage{
                Origin: g.Name,
                ID: g.nextMessageId,
                Text: contents,
            }

            // Increment messageId
            g.nextMessageId += 1

            // Add node (self) to the status and messages maps if not already there
            g.addNewNode(g.Name)

            // Increment vector clock ID for node
            g.incrementVectorClock(g.Name)

            // Store message
            g.storeMessage(rm)

            // Rumor RumorMessage
            g.sendRumorMessage(rm, true, "")
        }
    }
}

func (g *Gossiper) StartAntiEntropy() {
    for {
        time.Sleep(ANTI_ENTROPY_PERIOD * time.Second)
        if len(g.peers) > 0 {
            randomPeer := g.peers[rand.Intn(len(g.peers))]
            g.sendStatusMessage(randomPeer)
            fmt.Println("ANTI ENTROPY")
            fmt.Println()
        }
    }
}

func (g *Gossiper) SendSimpleMessage(contents string) {
    sm := model.SimpleMessage{
        OriginalName: g.Name,
        RelayPeerAddr: g.address.String(),
        Contents: contents,
    }

    gp := model.GossipPacket{Simple: &sm}

    g.printGossipPacket("client", "", &gp)

    g.sendGossipPacket(&gp, g.peers)
}

func (g *Gossiper) sendRumorMessage(rm *model.RumorMessage, random bool, addr string) {
    // Create the list of available peers by removing the sender
    availablePeers := collections.Filter(g.peers, func(p string) bool{
        return p != rm.Origin
    })

    // Check if the list of available peers is not empty
    if len(availablePeers) < 1 {
        return
    }

    peer := addr
    if random {
        // Select a random peer
        peer = availablePeers[rand.Intn(len(availablePeers))]
    }

    // Send him the packet
    gp := model.GossipPacket{Rumor: rm}
    g.sendGossipPacket(&gp, []string{peer})

    // TODO: Wait for acknowledgement

    g.printGossipPacket("mongering", peer, &gp)
}

func (g *Gossiper) sendStatusMessage(toPeer string) {
    // Prepare the list of wanted messages
    wantedList := make([]model.PeerStatus, 0)
    for _, vc := range g.status {
        wantedList = append(wantedList, *vc)
    }

    sp := model.StatusPacket{Want: wantedList}
    gp := model.GossipPacket{Status: &sp}

    g.sendGossipPacket(&gp, []string{toPeer})
}

func (g *Gossiper) sendGossipPacket(gp *model.GossipPacket, peersAddr []string) {
    packetBytes, err := protobuf.Encode(gp)
    if err != nil {
        fmt.Println(err)
        return
    }

    for i := 0; i < len(peersAddr); i++ {
        addr := resolveAddress(peersAddr[i])
        g.conn.WriteToUDP(packetBytes, addr)
    }
}

func (g *Gossiper) printGossipPacket(mode, relayAddr string, gp *model.GossipPacket) {
    packetToString := gp.String(mode, relayAddr)
    allPeersToString := "PEERS " + strings.Join(g.peers, ",")

    fmt.Println(packetToString)
    fmt.Println(allPeersToString)
    fmt.Println()
}

func (g *Gossiper) AddPeer(peer string) {
    // Don't add yourself
    if peer == g.address.String() {
        return
    }

    // Check if already present
    for _, a := range g.peers {
        if a == peer {
            return
        }
    }

    // Add the peer given that it isn't already in the list nor is yourself
    g.peers = append(g.peers, peer)
}

// Adds the node with name "origin" to the status and messages maps if not already present
func (g *Gossiper) addNewNode(origin string) {
    _, isInStatusMap := g.status[origin]
    _, isInMessagesMap := g.messages[origin]

    if !isInStatusMap || !isInMessagesMap {
        g.status[origin] = &model.PeerStatus{
            Identifier: origin,
            NextID: 1,
        }
        g.messages[origin] = make([]*model.RumorMessage, 0, 1024)
    }
}

func (g *Gossiper) getVectorClock(origin string) uint32 {
    g.addNewNode(origin)
    return g.status[origin].NextID
}

func (g *Gossiper) incrementVectorClock(origin string) {
	g.status[origin].NextID += 1
}

func (g *Gossiper) storeMessage(rm *model.RumorMessage) {
	g.messages[rm.Origin] = append(g.messages[rm.Origin], rm)
}

// Checks if otherStatus has something more than the current gossiper.
// If it is the case, the gossiper sends a StatusMessage to this peer.
// It returns true if the messages are in sync, false otherwise
func (g *Gossiper) getOutOfSyncMessages(peerAddr string, peerVectorClock []model.PeerStatus) bool {
    for i := 0; i < len(peerVectorClock); i++ {
        // Check if the origin is in the status map and the last message is in sync
        originStatus, isInStatusMap := g.status[peerVectorClock[i].Identifier]
        if !isInStatusMap || originStatus.NextID < peerVectorClock[i].NextID {
            g.sendStatusMessage(peerAddr)
            return false
        }
    }
    return true
}

func (g *Gossiper) sendMissingMessagesToPeer(peerAddr string, peerVectorClock []model.PeerStatus) {
    // Transform peerVectorClock in mapping from origin to PeerStatus.NextID
    peerStatus := make(map[string]uint32)
    for i := 0; i < len(peerVectorClock); i++ {
        peerStatus[peerVectorClock[i].Identifier] = peerVectorClock[i].NextID
    }

    // Find what the gossiper has but not the other peer
    for origin, vc := range g.status {
        nextId, isInStatusMap := peerStatus[origin]
        if !isInStatusMap || vc.NextID > nextId {
            // Send as RumorMessage all messages that the peer doen't have
            for msgId := int(nextId); msgId < len(g.messages[origin]); msgId++ {
                g.sendRumorMessage(g.messages[origin][msgId], false, peerAddr)
            }
        }
    }
}


/* HELPERS */

func resolveAddress(addr string) *net.UDPAddr {
    udpAddr, err := net.ResolveUDPAddr("udp4", addr)
    if err != nil {
        fmt.Println(err)
    }
    return udpAddr
}
