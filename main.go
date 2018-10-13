package main

import (
    "fmt"
    "flag"
    "strings"
    "github.com/pablo11/Peerster/gossip"
)

func main() {

    uiPort := flag.String("UIPort", "8080", "Port for the UI client (default \"8080\")")
    gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
    name := flag.String("name", "cryptop", "Name of the gossiper")
    peersParam := flag.String("peers", "", "Comma separated list of peers of the form ip:port")
    simple := flag.Bool("simple", false, "Run gossiper in simple broadcast mode")

    flag.Parse()

    // Prepare peers
    peers := strings.Split(*peersParam, ",")
    if *peersParam == "" {
        peers = make([]string, 0)
    }

    if *simple {
        g := gossip.NewGossiper(*gossipAddr, *name, peers)

        go g.ListenPeers()
        go g.ListenClient(*uiPort)

        for {}
    } else {
        fmt.Println("Not implemented yet. Please provide the -simple flag")
    }
}
