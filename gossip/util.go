package gossip

import (
    "fmt"
    "net"
    "crypto/sha256"
)

func resolveAddress(addr string) *net.UDPAddr {
    udpAddr, err := net.ResolveUDPAddr("udp4", addr)
    if err != nil {
        fmt.Println(err)
        return nil
    }
    return udpAddr
}

func hash(toHash []byte) []byte {
    h := sha256.New()
    h.Write(toHash)
    return h.Sum(nil)
}
