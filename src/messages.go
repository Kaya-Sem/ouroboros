package src

import (
	"encoding/gob"
	"log"
	"net"
	"time"
)

type MessageType int

const (
	Unknown MessageType = iota

	GetClientInfo
	GetClientInfoResponse

	GetPeersRequest
	GetPeersReply
)

// TODO: to be called before init servers etc
func InitializeInterfaces() {
	log.Println("Initialising gob data interfaces")
	gob.Register(ClientInfoData{})
}

type Message struct {
	Type MessageType
	Data interface{}
}

type ClientInfoData struct {
	Name        string
	OnlineSince time.Time
}

type HandlerFunc func(msg Message, conn net.Conn)

var handlers = map[MessageType]HandlerFunc{
	GetClientInfo: handleGetClientInfo,
}

func handleGetClientInfo(msg Message, conn net.Conn) {
	response := Message{
		Type: GetClientInfoResponse,
		Data: ClientInfoData{Name: "My PC", OnlineSince: time.Now()},
	}

	gob.NewEncoder(conn).Encode(response)
}
