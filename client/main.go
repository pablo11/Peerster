package main

import (
    "fmt"
    "flag"
    "net"
    "strings"
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
    keywords := flag.String("keywords", "", "Keywords for the file search")
    budget := flag.Int("budget", 0, "Budget for the file search")
    asset := flag.String("asset", "", "Name of the asset to create or transact")
    amount := flag.Int("amount", 0, "Amount to send in an asset transaction")
    identity := flag.String("identity", "", "Identity on the blockchain")
	question := flag.String("question", "", "Question for vote")
	assetVote := flag.String("assetVote", "", "Asset on which vote is done")
    sign := flag.Bool("sign", false, "Send signed content")
    encrypt := flag.Bool("encrypt", false, "Send encrypted content")

    flag.Parse()

    // Ask to insert an identity on the blockchain
    if *identity != "" {
        cm := &model.ClientMessage{
            Type: "identity",
            Identity: *identity,
        }
        sendPacket(cm, *uiPort)
        return
    }

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
            Sign: *sign,
            Encrypt: *encrypt,
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
            Sign: *sign,
            Encrypt: *encrypt,
        }
        sendPacket(cm, *uiPort)
        return
    }

    // Send search request
    if *keywords != "" {
        cm := &model.ClientMessage{
            Type: "searchFile",
            Keywords: strings.Split(*keywords, ","),
            Budget: uint64(*budget),
        }
        sendPacket(cm, *uiPort)
        return
    }

    // Send/Create an asset
    if *asset != "" && *dest != "" {
        if uint64(*amount) <= 0 {
            fmt.Println("Invalid amount of asset (it must be > 0)")
            return
        } else {
            cm := &model.ClientMessage{
                Type: "shareTx",
                Dest: *dest,
                Asset: *asset,
                Amount: uint64(*amount),
            }
            sendPacket(cm, *uiPort)
            return
        }
    }

    // Launch/Answer a poll
	if *question != "" && *assetVote != "" {
		cm := &model.ClientMessage{
            Type: "vote",
            Text: *question,
            Asset: *assetVote,
        }
        sendPacket(cm, *uiPort)
        return
	}

    fmt.Println("Please provide some valid parameters")
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
