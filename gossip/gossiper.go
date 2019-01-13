package gossip

import (
    "fmt"
    "log"
    "net"
    "strings"
    "sync"
    "math/rand"
    "time"
    "crypto/rsa"
    "github.com/dedis/protobuf"
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/util"
    "github.com/pablo11/Peerster/util/collections"
)

const (
    DEBUG bool = true
    PACKET_BUFFER_LEN int = 1024
    ACK_STATUS_WAIT_TIME time.Duration = 1 // Number of seconds to wait for a reply to a status message
    ANTI_ENTROPY_PERIOD time.Duration = 2
    SEARCH_REQUEST_DUPLICATE_PERIOD time.Duration = 500 // Milliseconds to wait before considering new SearchRequest as not duplicate
    SEARCH_REQUEST_BUDGET_DOUBLING_PERIOD time.Duration = 1
    MAX_SEARCH_BUDGET uint64 = 32
    SEARCH_REQUEST_MATCH_THRESHOLD int = 2
    GENESIS_BLOCK_WAIT_TIME time.Duration = 5
)

type Gossiper struct {
    address *net.UDPAddr
    conn *net.UDPConn
    Name string
    peers []string
    simple bool
    rtimer time.Duration
    nextMessageId uint32

    FileSharing *FileSharing

    status map[string]*model.PeerStatus
    statusMutex sync.Mutex

    messages map[string][]*model.RumorMessage
    messagesMutex sync.Mutex

    // Channel for wait for acknowledgement event
    waitStatusChannel map[string]chan bool
    waitStatusChannelMutex sync.Mutex

    // List of every message received to send to the web interface
    allMessages []*model.RumorMessage
    allMessagesMutex sync.Mutex

    // Routing table Origin->ip:port
    routingTable map[string]string
    routingTableMutex sync.Mutex

    // Array containing SearchRequest uid received in the last 0.5 seconds
    processingSearchRequests map[string]bool
    processingSearchRequestsMutex sync.Mutex

    // Keep track of current SearchRequests: mapping from keywords to ActiveSearch
    activeSearchRequests map[string]*model.ActiveSearch
    activeSearchRequestsMutex sync.Mutex

    FullMatches []*model.FileMatch
    FullMatchesMutex sync.Mutex

    Blockchain *Blockchain
    PrivateKey *rsa.PrivateKey

	QuestionKey map[string]string
	QuestionKeyMutex sync.Mutex
}

func NewGossiper(address, name string, peers []string, rtimer int, simple bool) *Gossiper {
    udpAddr := util.ResolveAddress(address)
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
        rtimer: time.Duration(rtimer),
        nextMessageId: 1,
        FileSharing: NewFileSharing(),
        status: make(map[string]*model.PeerStatus),
        statusMutex: sync.Mutex{},
        messages: make(map[string][]*model.RumorMessage),
        messagesMutex: sync.Mutex{},
        waitStatusChannel: make(map[string]chan bool),
        waitStatusChannelMutex: sync.Mutex{},
        allMessages: make([]*model.RumorMessage, 0),
        allMessagesMutex: sync.Mutex{},
        routingTable: make(map[string]string),
        routingTableMutex: sync.Mutex{},
        processingSearchRequests: make(map[string]bool),
        processingSearchRequestsMutex: sync.Mutex{},
        activeSearchRequests: make(map[string]*model.ActiveSearch),
        activeSearchRequestsMutex: sync.Mutex{},
        FullMatches: make([]*model.FileMatch, 0),
        FullMatchesMutex: sync.Mutex{},

        Blockchain: NewBlockchain(),
        PrivateKey: NewPrivateKey(),

		QuestionKey: make(map[string]string),
		QuestionKeyMutex: sync.Mutex{},
    }
}

func (g *Gossiper) Run(uiPort string) {
    fmt.Println("\033[0;32mGossiper " + g.Name + " started on " + g.address.String() + "\033[0m")
    fmt.Println()

    g.FileSharing.SetGossiper(g)
    g.Blockchain.SetGossiper(g)

    go g.listenPeers()
    go g.listenClient(uiPort)
    if (!g.simple) {
        go g.startAntiEntropy()
        go g.SendPublicMessage("", false)
        go g.startRouteRumoring()
    }
    go g.Blockchain.StartMining()
}

func (g *Gossiper) GetAddress() string {
    return g.address.String()
}

func (g *Gossiper) GetPeers() []string {
    return g.peers
}

func (g *Gossiper) GetOrigins() []string {
    g.routingTableMutex.Lock()
    defer g.routingTableMutex.Unlock()
    return collections.MapKeys(g.routingTable)
}

func (g *Gossiper) GetAllMessages() []*model.RumorMessage {
    g.allMessagesMutex.Lock()
    defer g.allMessagesMutex.Unlock()
    return g.allMessages
}

func (g *Gossiper) GetFullMatches() []*model.FileMatch {
    return g.FullMatches
}

func (g *Gossiper) listenPeers() {
    packetBuffer := make([]byte, 9 * PACKET_BUFFER_LEN)
    bytesRead := 0
    var fromAddr net.Addr = nil
    var err error = nil
    defer g.conn.Close()

    for {
        bytesRead, fromAddr, err = g.conn.ReadFrom(packetBuffer)
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

        go g.handlePeerReceivedPacket(&gp, fromAddr.String())
    }
}

func (g *Gossiper) handlePeerReceivedPacket(gp *model.GossipPacket, fromAddrStr string) {
    switch {
        case gp.Simple != nil:
            g.HandlePktSimple(gp)

        case gp.Rumor != nil:
            g.HandlePktRumor(gp, fromAddrStr)

        case gp.Status != nil:
            g.HandlePktStatus(gp, fromAddrStr)

        case gp.Private != nil:
            g.HandlePktPrivate(gp, fromAddrStr)

        case gp.DataRequest != nil:
            g.FileSharing.HandleDataRequest(gp.DataRequest)

        case gp.DataReply != nil:
            g.FileSharing.HandleDataReply(gp.DataReply)

        case gp.SearchRequest != nil:
            g.HandlePktSearchRequest(gp)

        case gp.SearchReply != nil:
            g.HandlePktSearchReply(gp)

        case gp.TxPublish != nil:
            g.Blockchain.HandlePktTxPublish(gp)

        case gp.BlockPublish != nil:
            g.Blockchain.HandlePktBlockPublish(gp)

        default:
            fmt.Println("WARNING: Unoknown message type")
    }

    gp = nil
}

func (g *Gossiper) listenClient(uiPort string) {
    var err error = nil
    udpAddr := util.ResolveAddress("127.0.0.1:" + uiPort)
    conn, err := net.ListenUDP("udp4", udpAddr)
    if err != nil {
        fmt.Println(err)
    }
    defer conn.Close()

    packetBuffer := make([]byte, 9 * PACKET_BUFFER_LEN)
    bytesRead := 0

    for {
        bytesRead, _, err = conn.ReadFromUDP(packetBuffer)
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

        g.HandlePktClient(&cm)
    }
}

func (g *Gossiper) SendPublicMessage(contents string, storeForGUI bool) {
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
        g.storeMessage(&rm, storeForGUI)

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
        go g.SendPublicMessage("", false)
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

    if (!DEBUG) {
        g.printGossipPacket("mongering", peer, &gp)
    }

    go g.sendGossipPacket(&gp, []string{peer})

    // Wait for acknowledgement
    go g.waitStatusAcknowledgement(addr, rm)
}

func (g *Gossiper) SendPrivateMessage(pm *model.PrivateMessage) {
    destPeer := g.GetNextHopForDest(pm.Destination)
    if destPeer == "" {
        return
    }

    gp := model.GossipPacket{Private: pm}

    go g.sendGossipPacket(&gp, []string{destPeer})
}

func (g *Gossiper) GetNextHopForDest(dest string) string {
    g.routingTableMutex.Lock()
    destPeer, destExists := g.routingTable[dest]
    g.routingTableMutex.Unlock()
    if !destExists {
        fmt.Println("WARNING: Node " + dest + " not in the routing table")
        return ""
    }
    return destPeer
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
    g.waitStatusChannelMutex.Lock()
    defer g.waitStatusChannelMutex.Unlock()
    _, channelExists := g.waitStatusChannel[addr]
    if !channelExists {
        g.waitStatusChannel[addr] = make(chan bool, 8)
    }
    return g.waitStatusChannel[addr]
}

func (g *Gossiper) removeChannelForPeer(addr string) {
    g.waitStatusChannelMutex.Lock()
    defer g.waitStatusChannelMutex.Unlock()
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
            if (!DEBUG) {
                fmt.Println("FLIPPED COIN sending rumor to " + randomPeer)
            }
        }
    }
}

func (g *Gossiper) sendStatusMessage(toPeer string) {
    // Prepare the list of wanted messages
    wantedList := make([]model.PeerStatus, 0)
    g.statusMutex.Lock()
    for _, vc := range g.status {
        wantedList = append(wantedList, *vc)
    }
    g.statusMutex.Unlock()

    sp := model.StatusPacket{Want: wantedList}
    gp := model.GossipPacket{Status: &sp}

    go g.sendGossipPacket(&gp, []string{toPeer})
}

func (g *Gossiper) sendGossipPacket(gp *model.GossipPacket, peersAddr []string) {
    packetBytes, err := protobuf.Encode(gp)
    if err != nil {
        fmt.Println("DEBUG PROTOBUF")
        fmt.Println(err)
        err = nil
        return
    }

    for i := 0; i < len(peersAddr); i++ {
        addr := util.ResolveAddress(peersAddr[i])
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
    g.messagesMutex.Lock()
    _, isInMessagesMap := g.messages[origin]
    g.messagesMutex.Unlock()

    if !isInStatusMap || !isInMessagesMap {
        g.statusMutex.Lock()
        g.status[origin] = &model.PeerStatus{
            Identifier: origin,
            NextID: 1,
        }
        g.statusMutex.Unlock()
        g.messagesMutex.Lock()
        g.messages[origin] = make([]*model.RumorMessage, 0, 1024)
        g.messagesMutex.Unlock()
    }
}

func (g *Gossiper) getVectorClock(origin string) uint32 {
    g.addNewNode(origin)
    return g.status[origin].NextID
}

func (g *Gossiper) incrementVectorClock(origin string) {
    g.statusMutex.Lock()
	g.status[origin].NextID += 1
    g.statusMutex.Unlock()
}

func (g *Gossiper) storeMessage(rm *model.RumorMessage, storeForGUI bool) {
    g.messagesMutex.Lock()
	g.messages[rm.Origin] = append(g.messages[rm.Origin], rm)
    g.messagesMutex.Unlock()

    if storeForGUI {
        // Add message also to allMessages for webserver
        g.allMessagesMutex.Lock()
        g.allMessages = append(g.allMessages, rm)
        g.allMessagesMutex.Unlock()
    }
}

func (g *Gossiper) updateRoutingTable(rm *model.RumorMessage, fromAddr string) {
    if vector, isPresent := g.status[rm.Origin]; isPresent && rm.ID < vector.NextID {
        return
    }

    g.routingTableMutex.Lock()
    if g.routingTable[rm.Origin] != fromAddr {
        g.routingTable[rm.Origin] = fromAddr
        fmt.Println("DSDV " + rm.Origin + " " + fromAddr)
        fmt.Println()
    }
    g.routingTableMutex.Unlock()
}
