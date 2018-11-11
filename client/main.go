package main

import (
    "fmt"
    "flag"
    "net"
    "github.com/dedis/protobuf"
    "github.com/pablo11/Peerster/model"
)

func main() {
    // Definition of the cli flags
    uiPort := flag.String("UIPort", "8080", "Port for the UI client (default \"8080\")")
    dest := flag.String("dest", "", "Destination for the private message")
    file := flag.String("file", "", "File to be indexed by the gossiper")
    msg := flag.String("msg", "", "Message to be sent")

    flag.Parse()

    if *msg == "" {
        fmt.Println("Please specify a message")
        return
    }

    sendPacket(*msg, *dest, *uiPort)
}

func sendPacket(msg, dest, uiPort string) {
    cm := &model.ClientMessage{
        Text: msg,
        Dest: dest,
    }

    packetBytes, err := protobuf.Encode(cm)
    if err != nil {
        fmt.Println(err)
        return
    }

    conn, e := net.Dial("udp", "127.0.0.1:" + uiPort)
	defer conn.Close()
	if e != nil {
		fmt.Println(e)
	}
	conn.Write(packetBytes)
}
