package src

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
)

const (
	SERVER_HOST = "0.0.0.0"
	SERVER_PORT = "8080"
	SERVER_TYPE = "tcp"
)

// peersHandler godoc
// @Summary      List connected peers
// @Description  Returns a list of currently connected TCP peers
// @Tags         peers
// @Produce      json
// @Success      200 {object} []Peer
// @Router       /peers [get]
func (s *Server) PeersHandler(w http.ResponseWriter, r *http.Request) {
	peers := s.GetPeers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

type Server struct {
	peers []*Peer
	mu    sync.Mutex // to protect concurrent access to peers
}

func (s *Server) AddPeer(p *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers = append(s.peers, p)
}

func (s *Server) RemovePeer(conn *net.TCPConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.peers {
		if p.Conn == conn {
			s.peers = append(s.peers[:i], s.peers[i+1:]...)
			return
		}
	}
}

func (s *Server) GetPeers() []Peer {
	s.mu.Lock()
	defer s.mu.Unlock()

	peersCopy := make([]Peer, len(s.peers))
	for i, p := range s.peers {
		peersCopy[i] = *p
	}
	return peersCopy
}

// IsPeerConnected checks if peer with given address is already connected
func (s *Server) IsPeerConnected(address string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.peers {
		if p.Address == address {
			return true
		}
	}
	return false
}

// ConnectToPeer attempts to establish a connection to a peer
func (s *Server) ConnectToPeer(address string, port int) {
	peerAddr := fmt.Sprintf("%s:%d", address, port)

	// Check if already connected
	if s.IsPeerConnected(peerAddr) {
		log.Printf("Already connected to peer at %s", peerAddr)
		return
	}

	log.Printf("Attempting to connect to peer at %s", peerAddr)

	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		log.Printf("Failed to connect to peer at %s: %v", peerAddr, err)
		return
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		log.Printf("Connection is not TCP")
		conn.Close()
		return
	}

	peer := NewPeer(tcpConn)
	s.AddPeer(peer)

	// Send initial handshake message
	initialMsg := fmt.Sprintf("Hello from %s", GetClientName())
	tcpConn.Write([]byte(initialMsg))

	// Handle peer in separate goroutine
	go peer.Handle(s)
}

func NewServer() *Server {
	return &Server{
		peers: make([]*Peer, 0, 32), // initial capacity 32
	}
}

func (s *Server) Start() {
	listen, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer listen.Close()

	log.Printf("Started listening server at %s:%s", SERVER_HOST, SERVER_PORT)

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			log.Printf("Connection is not TCP")
			conn.Close()
			continue
		}

		peer := NewPeer(tcpConn)
		s.AddPeer(peer)
		go peer.Handle(s)
	}
}
