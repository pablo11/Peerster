package main

import (
    "fmt"
    "flag"
    "strings"
    "github.com/pablo11/Peerster/server"
)

func main() {

    uiPort := flag.String("UIPort", "8080", "Port for the UI client (default \"8080\")")
    gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
    name := flag.String("name", "cryptop", "Name of the gossiper")
    peersParam := flag.String("peers", "", "Comma separated list of peers of the form ip:port")
    simple := flag.Bool("simple", false, "Run gossiper in simple broadcast mode")

    flag.Parse()

    peers := strings.Split(*peersParam, ",")
    if *peersParam == "" {
        peers = make([]string, 0)
    }

    /*simpleMessage := SimpleMessage{
        OriginalName: "",
        RelayPeerAddr: "",
        Contents: "",
    }
    packetToSend := GossipPacket{Simple: &simpleMessage}
*/

    if *simple {
        g := server.NewGossiper(*gossipAddr, *name, peers)

        go g.ListenPeers()
        go g.ListenClient(*uiPort)

        for {}
    } else {
        fmt.Println("Not implemented yet. Please provide the -simple flag")
    }
}

/*
func setupFlags() error {
    uiPort := flag.String("UIPort", "8080", "Port for the UI client (default \"8080\")")
    gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
    name := flag.String("name", "cryptop", "Name of the gossiper")
    peersParam := flag.String("peers", "", "Comma separated list of peers of the form ip:port")
    simple := flag.Bool("simple", false, "Run gossiper in simple broadcast mode")

    flag.Parse()

    // Validate uiPort
    err := validateUiPort(*uiPort)
    if err != nil {
        return err
    }

    // TODO: validate flags

    UI_PORT = *uiPort
    GOSSIP_ADDR = *gossipAddr
    NAME = *name
    PEERS = strings.Split(*peersParam, ",")
    SIMPLE_MODE = *simple

    return nil
}
*/
