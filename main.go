package main

import (
    "fmt"
    "flag"
    "net"
)

func main() {

    // Definition of the cli flags
    uiPort := flag.String("UIPort", "8080", "Port for the UI client (default \"8080\")")
    gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
    name := flag.String("name", "cryptop", "Name of the gossiper")
    peers := flag.String("peers", "8080", "Comma separated list of peers of the form ip:port")
    simple := flag.Bool("simple", false, "Run gossiper in simple broadcast mode")

    // TODO: validate flags
    err := validateUiPort(*uiPort)
    if err != nil {
        fmt.Println(err)
    }

    simpleMessage := SimpleMessage{
        OriginalName: "",
        RelayPeerAddr: "",
        Contents: "",
    }
    packetToSend := GossipPacket{Simple: &simpleMessage}


    fmt.Println("!oG ,olleH!!!")

    go listenOnPort("8080", func(conn net.Conn) {
        fmt.Println(conn)
    })
}

func listenOnPort(port string, handler func(net.Conn)) {
    ln, err := net.Listen("udp", ":" + port)
    if err != nil {
    	// handle error
        fmt.Println("Network connection error")
        fmt.Print(err)
        return
    }
    for {
    	conn, err := ln.Accept()
    	if err != nil {
    		// handle error
    	}
    	go handler(conn)


        // TODO: remove the return to have the process work forever
        return
    }
}

func handleListeningConnection(conn net.Conn) {


}
