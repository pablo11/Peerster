package main

import (
    "fmt"
    "flag"
    "net"
)

func main() {
    // Definition of the cli flags
    uiPort := flag.String("UIPort", "8080", "Port for the UI client (default \"8080\")")
    msg := flag.String("msg", "", "Message to be sent")

    flag.Parse()

    if *msg == "" {
        fmt.Println("Please specify a message")
        return
    }

    sendMessage(*msg, *uiPort)
}

func sendMessage(msg string, uiPort string) {
	conn, e := net.Dial("udp", "127.0.0.1:" + uiPort)
	defer conn.Close()
	if e != nil {
		fmt.Println(e)
	}
	conn.Write([]byte(msg))
}
