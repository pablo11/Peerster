package main

import (
    "fmt"
    "log"
    "flag"
    "strings"
    "net/http"
    "github.com/pablo11/Peerster/gossip"
    "github.com/pablo11/Peerster/webserver/api"
    //"github.com/gorilla/handlers"
    "github.com/gorilla/mux"
)

var WEBSERVER_PORT string

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

    WEBSERVER_PORT = *uiPort

    if *simple {
        g := gossip.NewGossiper(*gossipAddr, *name, peers)

        go g.ListenPeers()
        go g.ListenClient(*uiPort)

        go createWebserverAndRun(g)

        for {}
    } else {
        fmt.Println("Not implemented yet. Please provide the -simple flag")
    }
}

func createWebserverAndRun(g *gossip.Gossiper) {
    r := mux.NewRouter()
    a := api.NewApiHandler(g)

    // Get the JSON formatted list of messages
    r.HandleFunc("/api/message", a.GetMessages).Methods("GET")

    // Send a new message
    r.HandleFunc("/api/message", a.SendMessage).Methods("POST")

    // Get the list of known nodes
    r.HandleFunc("/api/node", a.GetNodes).Methods("GET")

    // Add a new node to the list of known nodes
    r.HandleFunc("/api/node", a.AddNode).Methods("POST")

    // Get the peer id
    r.HandleFunc("/api/id", a.GetId).Methods("GET")

    // Get the html index page
    r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("gui/"))))

    http.Handle("/", r)

    log.Fatal(http.ListenAndServe(":" + WEBSERVER_PORT, nil))
}
