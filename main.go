package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// @title TCP Client API
// @version 1.0
// @description A simple HTTP interface to your TCP client.
// @host localhost:8081
// @BasePath /

// ping godoc
// @Summary      Ping the service
// @Description  Responds with pong
// @Tags         health
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /ping [get]
func pingHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
}

const (
	SERVER_HOST = "0.0.0.0"
	SERVER_PORT = "8080"
	SERVER_TYPE = "tcp"
)

type Peer struct {
	conn    *net.TCPConn
	UID     string    `json:"uid"`
	Address string    `json:"address"`
	Since   time.Time `json:"connected_since"`
}

func newPeer(conn *net.TCPConn) *Peer {
	return &Peer{
		conn:    conn,
		UID:     conn.RemoteAddr().String(),
		Address: conn.RemoteAddr().String(),
		Since:   time.Now(),
	}
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
		if p.conn == conn {
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

// peersHandler godoc
// @Summary      List connected peers
// @Description  Returns a list of currently connected TCP peers
// @Tags         peers
// @Produce      json
// @Success      200 {object} []Peer
// @Router       /peers [get]
func (s *Server) peersHandler(w http.ResponseWriter, r *http.Request) {
	peers := s.GetPeers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

func (p *Peer) Handle(server *Server) {
	defer func() {
		p.conn.Close()
		server.RemovePeer(p.conn)
	}()

	log.Printf("Handling connection from: %s", p.Address)

	buffer := make([]byte, 1024)

	for {
		n, err := p.conn.Read(buffer)
		if err != nil {
			log.Printf("Connection closed (%s): %v", p.Address, err)
			return
		}

		msg := string(buffer[:n])
		log.Printf("[RECEIVED from %s] %s", p.Address, msg)

		time := time.Now().Format(time.ANSIC)
		responseStr := fmt.Sprintf("Echo: %v @ %v", msg, time)
		p.conn.Write([]byte(responseStr))
	}
}

func newServer() *Server {
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

		peer := newPeer(tcpConn)
		s.AddPeer(peer)
		go peer.Handle(s)
	}
}

func main() {
	server := newServer()

	go server.Start()

	time.Sleep(1 * time.Second)

	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/peers", server.peersHandler) // Use the server's method

	// qerve Swagger documentation
	http.Handle("/docs/", http.StripPrefix("/docs", http.FileServer(http.Dir("./docs"))))

	log.Println("HTTP API listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
