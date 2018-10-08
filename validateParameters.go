package main

import (
    "errors"
    "strconv"
)
/*
uiPort := flag.String("UIPort", "8080", "Port for the UI client (default \"8080\")")
gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
name := flag.String("name", "cryptop", "Name of the gossiper")
peers := flag.String("peers", "8080", "Comma separated list of peers of the form ip:port")
simple := flag.Bool("simple", false, "Run gossiper in simple broadcast mode")
*/
func validateUiPort(port string) error {
    portIntVal, err := strconv.ParseInt(port, 10, 0)
    if err != nil || portIntVal < 0 || portIntVal > 65535 {
        return errors.New("The port you provided is not valid.")
    }

    if portIntVal < 1024 {
        return errors.New("The port you provided is reserved (port numbers between 0 and 1024 are reserved).")
    }

    return nil
}

func validateGossipAddr(addr string) error {
    /*
    regex := "\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b:[0-65535]"

    match, _ := regexp.MatchString(regex, addr)


    fmt.Println(match)

    portIntVal, err := strconv.ParseInt(port, 10, 0)
    if err != nil || portIntVal < 0 || portIntVal > 65535 {
        return errors.New("The port you provided is not valid.")
    }

    if portIntVal < 1024 {
        return errors.New("The port you provided is reserved (port numbers between 0 and 1024 are reserved).")
    }
    */
    return nil
}
