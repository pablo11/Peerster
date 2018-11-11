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
    request := flag.String("request", "", "Request a chunk or metafile of this hash")

    flag.Parse()


    // Ask to index file
    if *file != "" && *request == "" {
        //indexFile("../_SharedFiles/" + *file)

        cm := &model.ClientMessage{
            Type: "indexFile",
            File: *file,
        }
        sendPacket(cm, *uiPort)
        return
    }

    // Ask to download file
    if *file != "" && *request != "" && *dest != "" {
        cm := &model.ClientMessage{
            Type: "downloadFile",
            File: *file,
            Request: *request,
            Dest: *dest,
        }
        sendPacket(cm, *uiPort)
        return
    }

    // Send message
    if *msg != "" {
        cm := &model.ClientMessage{
            Type: "msg",
            Text: *msg,
            Dest: *dest,
        }
        sendPacket(cm, *uiPort)
        return
    }

    fmt.Println("Please provide some parameters")
}

func sendPacket(cm *model.ClientMessage, uiPort string) {
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
