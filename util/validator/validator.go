package validator

import (
    "strconv"
    "strings"
)

func IsUiPort(port string) bool {
    if !IsIntBoundedBy(port, 0, 65535) {
        //fmt.Println("The port you provided is not valid.")
        return false
    }

    if IsIntBoundedBy(port, 0, 1024) {
        //fmt.Println("The port you provided is reserved (port numbers between 0 and 1024 are reserved).")
        return false
    }
    return true
}

// Validate strings of the form ip:port
func IsGossipAddr(addr string) bool {
    parts := strings.Split(addr, ":")
    if len(parts) != 2 {
        return false
    }

    // Check IP
    ipParts := strings.Split(parts[0], ".")
    if len(ipParts) != 4 {
        return false
    }

    for i := 0; i < 4; i++ {
        if !IsIntBoundedBy(ipParts[i], 0, 255) {
            return false
        }
    }

    // Check port
    if !IsIntBoundedBy(parts[1], 0, 65535) {
        return false
    }

    return true
}

func IsIntBoundedBy(val string, lowerBound, upperBound int64) bool {
    intVal, err := strconv.ParseInt(val, 10, 0)
    return !(err != nil || intVal < lowerBound || intVal > upperBound)
}
