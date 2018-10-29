package model

import (
    "strconv"
)

type PrivateMessage struct {
    Origin string
    ID uint32
    Text string
    Dest string
    HopLimit uint32
}

func NewPrivateMessage(origin, text, dest string) *PrivateMessage {
    return &PrivateMessage{
        Origin: origin,
        ID: 0,
        Text: text,
        Dest: dest,
        HopLimit: 10,
    }
}

func (pm *PrivateMessage) String() string {
    hopLimitStr := strconv.FormatUint(uint64(pm.HopLimit), 10)
    return "PRIVATE origin " + pm.Origin + " hop-limit " + hopLimitStr + " contents " + pm.Text
}
