package webserver

import (
    "log"
    "fmt"
    "net/http"
    "github.com/pablo11/Peerster/gossip"
    "github.com/pablo11/Peerster/webserver/api"
    "github.com/gorilla/mux"
)


func CreateAndRun(g *gossip.Gossiper, webserverPort string) {
    fmt.Println("\033[0;32mWebserver listening on localhost:" + webserverPort + "\033[0m")
    fmt.Println()

    r := mux.NewRouter()
    a := api.NewApiHandler(g)

    // Get the JSON formatted list of messages
    r.HandleFunc("/api/messages", a.GetMessages).Methods("GET")

    // Send a new public (broadcast) message
    r.HandleFunc("/api/sendPublicMessage", a.SendPublicMessage).Methods("POST")

    // Get the list of origins known to this peer
    r.HandleFunc("/api/origins", a.GetOrigins).Methods("GET")

    // Send a new private message
    r.HandleFunc("/api/sendPrivateMessage", a.SendPrivateMessage).Methods("POST")

    // Get the list of known nodes
    r.HandleFunc("/api/nodes", a.GetNodes).Methods("GET")

    // Add a new node to the list of known nodes
    r.HandleFunc("/api/node", a.AddNode).Methods("POST")

    // Get the peer id
    r.HandleFunc("/api/id", a.GetId).Methods("GET")

    // Upload a file
    r.HandleFunc("/api/uploadFile", a.UploadFile).Methods("POST")

    // Request a file
    r.HandleFunc("/api/requestFile", a.RequestFile).Methods("POST")

    // List available files
    r.HandleFunc("/api/listFiles", a.ListFiles).Methods("GET")

    // Download file
    r.HandleFunc("/api/downloadFile", a.DownloadFile).Methods("GET")

    // Get the html index page
    r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webserver/gui/"))))

    http.Handle("/", r)

    log.Fatal(http.ListenAndServe(":" + webserverPort, nil))
}
