package util

import (
	"crypto/sha256"
	"fmt"
	"net"
)

func ResolveAddress(addr string) *net.UDPAddr {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return udpAddr
}

func Hash(toHash []byte) []byte {
	h := sha256.New()
	h.Write(toHash)
	return h.Sum(nil)
}
