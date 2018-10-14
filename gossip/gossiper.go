package gossip

import (
    "fmt"
    "log"
    "net"
    "strings"
    "bytes"
    "github.com/dedis/protobuf"
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/collections"
)

type Gossiper struct {
    address *net.UDPAddr
    conn *net.UDPConn
    Name string
    peers []string
    simple bool
}

func NewGossiper(address, name string, peers []string, simple bool) *Gossiper {
    udpAddr, err := net.ResolveUDPAddr("udp4", address)
    if err != nil {
        log.Fatal(err)
    }

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
        packetBuffer := make([]byte, 2*1024)
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
                g.printReceivedPacket("peer", &gp)

                // Change the relay peer field to this node address
                receivedFrom := gp.Simple.RelayPeerAddr
                gp.Simple.RelayPeerAddr = g.address.String()

                // Broadcast the message to every peer except the one the message was received from
                go g.sendPacket(&gp, collections.Filter(g.peers, func(p string) bool{
                    return p != receivedFrom
                }))

            default:
                fmt.Println("WARNING: Unoknown message type")
        }

    }
}

func (g *Gossiper) ListenClient(uiPort string) {
    udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:" + uiPort)
    if err != nil {
        fmt.Println(err)
    }

    conn, err := net.ListenUDP("udp4", udpAddr)
    if err != nil {
        fmt.Println(err)
    }

    packetBuffer := make([]byte, 1024)

    for {
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
            fmt.Println("Not implemented!")
            // TODO
        }
    }
}

func (g *Gossiper) SendSimpleMessage(contents string) {
    sm := model.SimpleMessage{
        OriginalName: g.Name,
        RelayPeerAddr: g.address.String(),
        Contents: contents,
    }

    gossipPacket := model.GossipPacket{Simple: &sm}

    g.printReceivedPacket("client", &gossipPacket)

    g.sendPacket(&gossipPacket, g.peers)
}

func (g *Gossiper) sendPacket(pkt *model.GossipPacket, peersAddr []string) {
    packetBytes, err := protobuf.Encode(pkt)
    if err != nil {
        fmt.Println(err)
        return
    }

    for i := 0; i < len(peersAddr); i++ {
        addr, _ := net.ResolveUDPAddr("udp", peersAddr[i])
        g.conn.WriteToUDP(packetBytes, addr)
    }
}

func (g *Gossiper) printReceivedPacket(mode string, pkt *model.GossipPacket) {
    packetToString := pkt.String(mode)
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
