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
    msg := flag.String("msg", "", "Message to be sent")
    dest := flag.String("dest", "", "Destination for the private message")

    flag.Parse()

    if *msg == "" {
        fmt.Println("Please specify a message")
        return
    }

    //sendMessage(*msg, *uiPort)
    sendPacket(*msg, *dest, *uiPort)
}

/*
func sendMessage(msg, uiPort string) {
	conn, e := net.Dial("udp", "127.0.0.1:" + uiPort)
	defer conn.Close()
	if e != nil {
		fmt.Println(e)
	}
	conn.Write([]byte(msg))
}
*/

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
