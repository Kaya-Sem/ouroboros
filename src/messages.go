package src

import (
	"encoding/gob"
	"log"
	"time"
)

type MessageType int

const (
	Unknown MessageType = iota

	PeerDiscovery
	PeerDiscoveryReply

	GetClientInfo
	GetClientInfoResponse

	GetPeersRequest
	GetPeersReply

	ClientDiscoveryAnnouncement
	ClientDiscoveryAnnouncementReply
)

/* should be called before any messages are processed or sent */
func InitializeInterfaces() {
	gob.Register(ClientInfoData{})
	gob.Register(PeerDiscoveryData{})
	gob.Register(PeerDiscoveryReplyData{})
	gob.Register(ClientDiscoveryAnnouncementData{})
}

type Message struct {
	Type   MessageType
	Data   interface{}
	Sender string
}

type ClientInfoData struct {
	Name        string
	OnlineSince time.Time
}

type ClientDiscoveryAnnouncementData struct {
	ClientName string
	MessageID  int
}

type PeerDiscoveryData struct {
	ClientName    string
	SenderAddress string
	ID            int
}

type PeerDiscoveryReplyData struct {
	SenderAddres string
	ID           int
}

type HandlerFunc func(msg Message, bus *MessageBus)

func handleGetClientInfo(msg Message) {
	response := Message{
		Type: GetClientInfoResponse,
		Data: ClientInfoData{Name: "My PC", OnlineSince: time.Now()},
	}

	_ = response

}

func handlePeerDiscovery(msg Message) {
	discoveryData, ok := msg.Data.(PeerDiscoveryData)
	if !ok {
		log.Println("Error: Invalid data format for PeerDiscovery message")
		return
	}

	response := Message{
		Type: PeerDiscoveryReply,
		Data: PeerDiscoveryReplyData{
			SenderAddres: "nonce", // Replace with actual address
			ID:           discoveryData.ID,
		},
	}

	_ = response

	// You'll need to send this response somewhere
	// For example, you might want to pass in the sender's connection or address
}
