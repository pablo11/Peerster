package main

import (
    "log"
    "net/http"
    "github.com/pablo11/Peerster/gossip"
    "github.com/pablo11/Peerster/webserver/api"
    "github.com/gorilla/mux"
)


func createWebserverAndRun(g *gossip.Gossiper, webserverPort string) {
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

    log.Fatal(http.ListenAndServe(":" + webserverPort, nil))
}
