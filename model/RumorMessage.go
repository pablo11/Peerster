package model

import (
	"strconv"
)

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

func (rm *RumorMessage) String(mode, relayAddr string) string {
	switch mode {
	case "received":
		idStr := strconv.FormatUint(uint64(rm.ID), 10)
		return "RUMOR origin " + rm.Origin + " from " + relayAddr + " ID " + idStr + " contents " + rm.Text

	case "mongering":
		return "MONGERING with " + relayAddr

	default:
		return ""
	}
}

func (rm *RumorMessage) ToJSON() string {
	return `{"from":"` + rm.Origin + `","msg":"` + rm.Text + `"}`
}
