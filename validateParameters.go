package main

import "errors"
/*
uiPort := flag.String("UIPort", "8080", "Port for the UI client (default \"8080\")")
gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
name := flag.String("name", "cryptop", "Name of the gossiper")
peers := flag.String("peers", "8080", "Comma separated list of peers of the form ip:port")
simple := flag.Bool("simple", false, "Run gossiper in simple broadcast mode")
*/
func validateUiPort(port string) error {
    return errors.New("can't work with 42")
}
