package gossip

import (
    "fmt"
    "log"
    "net"
    "strings"
    //"bytes"
    "sync"
    "math/rand"
    "time"
    "github.com/dedis/protobuf"
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/collections"
)

const DEBUG bool = false

const PACKET_BUFFER_LEN int = 1024
const ACK_STATUS_WAIT_TIME time.Duration = 1 // Number of seconds to wait
const ANTI_ENTROPY_PERIOD time.Duration = 2 // Number of seconds to wait

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
    mutex sync.Mutex
    /*
    peersMutex sync.Mutex
    nextMessageIdMutex sync.Mutex
    statusMutex sync.Mutex
    messagesMutex sync.Mutex
    */

    // Channel for wait for acknowledgement event
    waitStatusChannel map[string]chan bool

    // List of every message received to send to the web interface
    allMessages []*model.RumorMessage

    // Routing table Origin->ip:port
    routingTable map[string]string

    rtimer time.Duration

    fileSharing *FileSharing
}

func NewGossiper(address, name string, peers []string, rtimer int, simple bool) *Gossiper {
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

        mutex: sync.Mutex{},
        /*
        peersMutex: sync.Mutex{},
        nextMessageIdMutex: sync.Mutex{},
        statusMutex: sync.Mutex{},
        messagesMutex: sync.Mutex{},
        */
        waitStatusChannel: make(map[string]chan bool),
        //allMessages: make([]*model.RumorMessage),
        routingTable: make(map[string]string),
        rtimer: time.Duration(rtimer),
        fileSharing: NewFileSharing(),
    }
}

func (g *Gossiper) Run(uiPort string) {
    fmt.Println("\033[0;32mGossiper " + g.Name + " started on " + g.address.String() + "\033[0m")
    fmt.Println()

    g.fileSharing.SetGossiper(g)

    go g.listenPeers()
    go g.listenClient(uiPort)
    if (!g.simple) {
        go g.startAntiEntropy()
        go g.sendRouteRumorMessage(true)
        go g.startRouteRumoring()
    }
}

func (g *Gossiper) GetAddress() string {
    return g.address.String()
}

func (g *Gossiper) GetPeers() []string {
    return g.peers
}

func (g *Gossiper) GetOrigins() []string {
    return collections.MapKeys(g.routingTable)
}

func (g *Gossiper) GetAllMessages() []*model.RumorMessage {
    return g.allMessages
}

func (g *Gossiper) listenPeers() {
    packetBuffer := make([]byte, 9 * PACKET_BUFFER_LEN)
    for {
        bytesRead, fromAddr, err := g.conn.ReadFrom(packetBuffer)
        if err != nil {
            fmt.Println(err)
            continue
        }

        // Decode the message
        gp := model.GossipPacket{}
        err = protobuf.Decode(packetBuffer[:bytesRead], &gp)
        if err != nil {
            //fmt.Println("ERROR:", err)
            err = nil
        }

        // Store addr in the list of peers if not already present
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

                isRouteRumor := gp.Rumor.Text == ""

                if !isRouteRumor {
                    g.printGossipPacket("received", fromAddr.String(), &gp)
                }

                g.updateRoutingTable(gp.Rumor, fromAddr.String())

                // If the message is the next one expected, store it
                if !isRouteRumor && gp.Rumor.ID == g.getVectorClock(gp.Rumor.Origin) {
                    g.incrementVectorClock(gp.Rumor.Origin)
                    g.storeMessage(gp.Rumor)
                    g.sendRumorMessage(gp.Rumor, true, fromAddr.String())
                }

                // Send status message to the peer the rumor message was received from
                g.sendStatusMessage(fromAddr.String())

            case gp.Status != nil:
                if (!DEBUG) {
                    g.printGossipPacket("", fromAddr.String(), &gp)
                }

                g.compareVectorClocks(gp.Status, fromAddr.String())

            case gp.Private != nil:
                if gp.Private.Destination == g.Name {
                    // If the private message is for this node, display it
                    g.printGossipPacket("", fromAddr.String(), &gp)
                } else {
                    // Forward the message and decrease the HopLimit
                    pm := gp.Private
                    fmt.Println("🧠 Forwarding private msg dest " + pm.Destination)
                    if pm.HopLimit > 1 {
                        pm.HopLimit -= 1
                        g.SendPrivateMessage(pm)
                    }
                }

            case gp.DataReply != nil:
                go g.fileSharing.HandleDataReply(gp.DataReply)

            case gp.DataRequest != nil:
                fmt.Println("❤️")
                go g.fileSharing.HandleDataRequest(gp.DataRequest)

            default:
                fmt.Println("WARNING: Unoknown message type")
        }
    }
}

func (g *Gossiper) compareVectorClocks(sp *model.StatusPacket, fromAddr string) {
    // Prepare g.status to compare vactors clocks
    tmpStatus := make(map[string]bool)
    for key, _ := range g.status {
        tmpStatus[key] = false
    }

    // Compare the two vector clocks
    for i := 0; i < len(sp.Want); i++ {
        otherStatusPeer := sp.Want[i]

        statusPeer, exists := g.status[otherStatusPeer.Identifier]
        if exists {
            tmpStatus[otherStatusPeer.Identifier] = true
            if otherStatusPeer.NextID > statusPeer.NextID {
                // The other peer has something more, so send StatusPacket
                g.sendStatusMessage(fromAddr)

                // Don't flip the coin and stop timer
                g.getChannelForPeer(fromAddr) <- false
                return
            } else if otherStatusPeer.NextID < statusPeer.NextID {
                // The gossiper has something more, so send rumor of this thing
                rm := g.messages[otherStatusPeer.Identifier][otherStatusPeer.NextID - 1]
                g.sendRumorMessage(rm, false, fromAddr)

                // Don't flip the coin and stop timer
                g.getChannelForPeer(fromAddr) <- false
                return
            }
        } else {
            // The other peer has something more, so send status
            g.sendStatusMessage(fromAddr)

            // Don't flip the coin and stop timer
            g.getChannelForPeer(fromAddr) <- false
            return
        }
    }

    if len(sp.Want) == len(g.status) {
        // The two vectors are the same -> we are in sync with the peer
        if (!DEBUG) {
            fmt.Println("IN SYNC WITH " + fromAddr)
            fmt.Println()
        }

        // Flip the coin and stop timer
        g.getChannelForPeer(fromAddr) <- true
        return
    } else {
        // The peer vector cannot be longer than the gossiper vector clock, otherwise we don't get here
        // Find the first message from tmpStatus to send
        for key, isVisited := range tmpStatus {
            if !isVisited && len(g.messages[key]) > 0 {
                rm := g.messages[key][0]
                g.sendRumorMessage(rm, false, fromAddr)

                // Don't flip the coin and stop timer
                g.getChannelForPeer(fromAddr) <- false
                return
            }
        }
    }

    tmpStatus = nil
}

func (g *Gossiper) listenClient(uiPort string) {
    udpAddr := resolveAddress("127.0.0.1:" + uiPort)
    conn, err := net.ListenUDP("udp4", udpAddr)
    if err != nil {
        fmt.Println(err)
    }

    packetBuffer := make([]byte, 9 * PACKET_BUFFER_LEN)
    for {
        bytesRead, _, err := conn.ReadFromUDP(packetBuffer)
        if err != nil {
            fmt.Println(err)
            continue
        }

        // Decode the message
        cm := model.ClientMessage{}
        err = protobuf.Decode(packetBuffer[:bytesRead], &cm)
        if err != nil {
            //fmt.Println("ERROR:", err)
            err = nil
        }

        switch cm.Type {
            case "msg":
                fmt.Println(cm.String())
                fmt.Println()

                if cm.Dest == "" {
                    go g.SendPublicMessage(cm.Text)
                } else {
                    pm := model.NewPrivateMessage(g.Name, cm.Text, cm.Dest)
                    go g.SendPrivateMessage(pm)
                }

            case "indexFile":
                fmt.Println("♻️ INDEXING FILE " + cm.File)
                fmt.Println()
                go g.fileSharing.IndexFile(cm.File)

            case "downloadFile":
                fmt.Println("✅ START DOWNLOADING FILE " + cm.File + " - " + time.Now().String())
                fmt.Println()
                go g.fileSharing.RequestFile(cm.File, cm.Dest, cm.Request)
        }
    }
}

func (g *Gossiper) SendPublicMessage(contents string) {
    if g.simple {
        go g.sendSimpleMessage(contents)
    } else {
        // Build RumorMessage
        rm := model.RumorMessage{
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
        g.storeMessage(&rm)

        // Rumor RumorMessage
        g.sendRumorMessage(&rm, true, "")
    }
}

func (g *Gossiper) startAntiEntropy() {
    for {
        time.Sleep(ANTI_ENTROPY_PERIOD * time.Second)
        if len(g.peers) > 0 {
            randomPeer := g.peers[rand.Intn(len(g.peers))]
            g.sendStatusMessage(randomPeer)
        }
    }
}

func (g *Gossiper) startRouteRumoring() {
    if g.rtimer == 0 {
        return
    }

    for {
        time.Sleep(g.rtimer * time.Second)
        go g.sendRouteRumorMessage(false)
    }
}

func (g *Gossiper) sendSimpleMessage(contents string) {
    sm := model.SimpleMessage{
        OriginalName: g.Name,
        RelayPeerAddr: g.address.String(),
        Contents: contents,
    }

    gp := model.GossipPacket{Simple: &sm}

    go g.sendGossipPacket(&gp, g.peers)
}

// If random is true, addr is used as "not send to this one"
// If random is false, the rumor message is sent to addr
func (g *Gossiper) sendRumorMessage(rm *model.RumorMessage, random bool, addr string) {
    peer := addr
    if random {
        // Create the list of available peers by removing the sender
        availablePeers := collections.Filter(g.peers, func(p string) bool{
            return p != addr
        })

        // Check if the list of available peers is not empty
        if len(availablePeers) < 1 {
            return
        }

        // Select a random peer
        peer = availablePeers[rand.Intn(len(availablePeers))]
    }

    // Send him the packet
    gp := model.GossipPacket{Rumor: rm}

    g.printGossipPacket("mongering", peer, &gp)

    go g.sendGossipPacket(&gp, []string{peer})

    // Wait for acknowledgement
    go g.waitStatusAcknowledgement(addr, rm)
}

func (g *Gossiper) SendPrivateMessage(pm *model.PrivateMessage) {
    destPeer := g.GetNextHopForDest(pm.Destination)
    fmt.Println("Sending PRIVATE message 🍿 1 : " + pm.Text + " to " + destPeer)
    if destPeer == "" {
        return
    }

    gp := model.GossipPacket{Private: pm}

    go g.sendGossipPacket(&gp, []string{destPeer})
}

func (g *Gossiper) GetNextHopForDest(dest string) string {
    destPeer, destExists := g.routingTable[dest]
    if !destExists {
        fmt.Println("🤬 Node " + dest + " not in the routing table")
        return ""
    }
    return destPeer
}

func (g *Gossiper) sendRouteRumorMessage(broadcast bool) {
    rm := model.RumorMessage{
        Origin: g.Name,
        ID: g.nextMessageId,
        Text: "",
    }

    gp := model.GossipPacket{Rumor: &rm}

    if broadcast {
        go g.sendGossipPacket(&gp, g.peers)
    } else {
        if (len(g.peers) > 0) {
            randomPeer := g.peers[rand.Intn(len(g.peers))]
            go g.sendGossipPacket(&gp, []string{randomPeer})
        }
    }
}

func (g *Gossiper) waitStatusAcknowledgement(fromAddr string, rm *model.RumorMessage) {
    ticker := time.NewTicker(ACK_STATUS_WAIT_TIME * time.Second)
    defer ticker.Stop()

    channel := g.getChannelForPeer(fromAddr)

    select {
    case flipCoin := <-channel:
        ticker.Stop()

        // If we get ther, it means that the StatusPacket did not acknowledge the RumorMessage
        if flipCoin {
            g.flipCoin(rm)
        }

        g.removeChannelForPeer(fromAddr)

    case <-ticker.C:
        ticker.Stop()
        g.flipCoin(rm)
        g.removeChannelForPeer(fromAddr)
    }
}

func (g *Gossiper) getChannelForPeer(addr string) chan bool {
    _, channelExists := g.waitStatusChannel[addr]
    if !channelExists {
        g.waitStatusChannel[addr] = make(chan bool, 8)
    }
    return g.waitStatusChannel[addr]
}

func (g *Gossiper) removeChannelForPeer(addr string) {
    _, channelExists := g.waitStatusChannel[addr]
    if channelExists {
        g.waitStatusChannel[addr] = nil
        delete(g.waitStatusChannel, addr)
    }
}

func (g *Gossiper) flipCoin(rm *model.RumorMessage) {
    if rand.Int() % 2 == 0 {
        if len(g.peers) > 0 {
            randomPeer := g.peers[rand.Intn(len(g.peers))]
            g.sendRumorMessage(rm, false, randomPeer)
            fmt.Println("FLIPPED COIN sending rumor to " + randomPeer)
        }
    }
}

func (g *Gossiper) sendStatusMessage(toPeer string) {
    // Prepare the list of wanted messages
    wantedList := make([]model.PeerStatus, 0)
    for _, vc := range g.status {
        wantedList = append(wantedList, *vc)
    }

    sp := model.StatusPacket{Want: wantedList}
    gp := model.GossipPacket{Status: &sp}

    go g.sendGossipPacket(&gp, []string{toPeer})
}

func (g *Gossiper) sendGossipPacket(gp *model.GossipPacket, peersAddr []string) {
    packetBytes, err := protobuf.Encode(gp)
    if err != nil {
        fmt.Println(err)
        err = nil
        return
    }

    for i := 0; i < len(peersAddr); i++ {
        addr := resolveAddress(peersAddr[i])
        if (addr != nil) {
            if _, err2 := g.conn.WriteToUDP(packetBytes, addr); err2 != nil {
                fmt.Println(err2)
            }
        }
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

    // Add message also to allMessages for webserver
    g.allMessages = append(g.allMessages, rm)
}

func (g *Gossiper) updateRoutingTable(rm *model.RumorMessage, fromAddr string) {
    if vector, isPresent := g.status[rm.Origin]; isPresent && rm.ID < vector.NextID {
        return
    }

    if g.routingTable[rm.Origin] != fromAddr {
        g.routingTable[rm.Origin] = fromAddr
        fmt.Println("DSDV " + rm.Origin + " " + fromAddr)
        fmt.Println()
    }
}


/* HELPERS */

func resolveAddress(addr string) *net.UDPAddr {
    udpAddr, err := net.ResolveUDPAddr("udp4", addr)
    if err != nil {
        fmt.Println(err)
        return nil
    }
    return udpAddr
}
