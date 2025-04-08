package src

import (
	"encoding/gob"
	"log"
	"net"
	"time"
)

type Peer struct {
	Conn    *net.TCPConn `swaggerignore:"true"`
	UID     string       `json:"uid"`
	Address string       `json:"address"`
	Since   time.Time    `json:"connected_since"`
}

// sendMessage sends a structured message to the peer using gob encoding
func (p *Peer) SendMessage(m Message) error {
	p.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	defer p.Conn.SetWriteDeadline(time.Time{})

	enc := gob.NewEncoder(p.Conn)
	if err := enc.Encode(m); err != nil {
		log.Printf("Failed to send message to %s: %v", p.Address, err)
		return err
	}
	return nil
}

func NewPeer(conn *net.TCPConn) *Peer {
	return &Peer{
		Conn:    conn,
		UID:     conn.RemoteAddr().String(),
		Address: conn.RemoteAddr().String(),
		Since:   time.Now(),
	}
}

func (p *Peer) Handle(server *Server) {
	defer func() {
		p.Conn.Close()
		server.RemovePeer(p.Conn)
	}()

	log.Printf("Handling connection from peer with address %s", p.Address)

	buffer := make([]byte, 1024)

	for {
		n, err := p.Conn.Read(buffer)
		if err != nil {
			log.Printf("Connection closed (%s): %v", p.Address, err)
			return
		}

		msg := string(buffer[:n])
		log.Printf("[RECEIVED from %s] %s", p.Address, msg)

		// time := time.Now().Format(time.ANSIC)
		// responseStr := fmt.Sprintf("Echo: %v @ %v", msg, time)
		// p.conn.Write([]byte(responseStr))
	}
}
