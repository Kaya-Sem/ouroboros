package src

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

const (
	SERVER_HOST = "0.0.0.0"
	SERVER_PORT = "8080"
)

type MessageBus struct {
	peers    map[string]*Peer // Key is remote address
	udpConn  *net.UDPConn
	mu       sync.Mutex // to protect concurrent access to peers
	handlers map[MessageType]HandlerFunc
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewMessageBus() *MessageBus {
	InitializeInterfaces()
	return &MessageBus{
		peers: make(map[string]*Peer),
		handlers: map[MessageType]HandlerFunc{
			ClientDiscoveryAnnouncement: handleClientDiscoveryAnnouncement,
		},
		stopChan: make(chan struct{}),
	}
}

// Start initializes the UDP server
func (s *MessageBus) Start() error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", SERVER_HOST, SERVER_PORT))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	s.udpConn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}
	defer s.udpConn.Close()

	log.Printf("Started UDP server at %s:%s", SERVER_HOST, SERVER_PORT)

	s.wg.Add(1)
	go s.handleUDPMessages()

	// Wait for shutdown signal
	<-s.stopChan
	s.wg.Wait()
	return nil
}

// Stop shuts down the server gracefully
func (s *MessageBus) Stop() {
	close(s.stopChan)
	if s.udpConn != nil {
		s.udpConn.Close()
	}
	s.wg.Wait()
}

// handleUDPMessages processes incoming UDP messages
func (s *MessageBus) handleUDPMessages() {
	defer s.wg.Done()

	buf := make([]byte, 2048)

	for {
		select {
		case <-s.stopChan:
			return
		default:
			n, addr, err := s.udpConn.ReadFromUDP(buf)
			if err != nil {
				if _, ok := err.(net.Error); ok {
					log.Printf("UDP read error: %v", err)
					continue
				}
				break
			}

			// Process the message
			go s.processMessage(buf[:n], addr)
		}
	}
}

// processMessage decodes and handles an incoming message
func (s *MessageBus) processMessage(data []byte, addr *net.UDPAddr) {
	var msg Message
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&msg); err != nil {
		log.Printf("Error decoding message from %s: %v", addr.String(), err)
		return
	}

	if handler, ok := s.handlers[msg.Type]; ok {
		handler(msg, s)
	} else {
		log.Printf("No handler for message type %d from %s", msg.Type, addr.String())
	}
}

// SendMessage sends a message to a specific peer
func (s *MessageBus) SendMessage(p *Peer, msg Message) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(msg); err != nil {
		return fmt.Errorf("error encoding message: %w", err)
	}

	msg.Sender = s.udpConn.LocalAddr().String()

	_, err := s.udpConn.WriteToUDP(buf.Bytes(), p.Address)
	return err
}

func (bus *MessageBus) BroadcastToPeers(msg Message) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(msg); err != nil {
		return fmt.Errorf("error encoding message: %w", err)
	}

	for _, peer := range bus.peers {
		if _, err := bus.udpConn.WriteToUDP(buf.Bytes(), peer.Address); err != nil {
			log.Printf("Error sending to %s: %v", peer.Address.String(), err)
		}
	}

	return nil
}

// PeersHandler godoc
// @Summary      List connected peers
// @Description  Returns a list of currently connected UDP peers
// @Tags         peers
// @Produce      json
// @Success      200 {object} []Peer
// @Router       /peers [get]
func (s *MessageBus) PeersHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peers := make([]Peer, 0, len(s.peers))
	for _, p := range s.peers {
		peers = append(peers, *p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

func (s *MessageBus) AddPeer(name string, conn *net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := NewPeer(name, conn)

	s.peers[p.Address.String()] = p
}

func (s *MessageBus) HasPeer(addr string) bool {
	_, exists := s.peers[addr]
	return exists
}

func (s *MessageBus) GetPeer(addres string) (*Peer, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peer, exists := s.peers[addres]
	return peer, exists
}

func handleClientDiscoveryAnnouncement(msg Message, bus *MessageBus) {
	discoveryData, ok := msg.Data.(ClientDiscoveryAnnouncementData)
	if !ok {
		log.Println("Error: Invalid data format for PeerDiscovery message")
		return
	}

	if !bus.HasPeer(msg.Sender) {
		udpAddr, err := net.ResolveUDPAddr("udp", msg.Sender)
		if err != nil {
			log.Printf("Could not resolve %s", msg.Sender)
			return
		}

		bus.AddPeer(discoveryData.ClientName, udpAddr)
	}

	peer, _ := bus.GetPeer(msg.Sender)

	response := Message{
		Type: ClientDiscoveryAnnouncementReply,
		Data: ClientDiscoveryAnnouncementData{GetClientName(), discoveryData.MessageID},
	}

	bus.SendMessage(peer, response)

	log.Printf("New peer %s with adress %s. \n", discoveryData.ClientName, msg.Sender)
}

func (s *MessageBus) AnnouncePresence() {
	localIP, err := getLocalIP()
	if err != nil {
		log.Printf("Failed to determine local IP: %v", err)
		localIP = "unknown"
	}

	msg := Message{
		Type: ClientDiscoveryAnnouncement,
		Data: ClientDiscoveryAnnouncementData{
			ClientName: GetClientName(),
			MessageID:  69, // TODO: Replace with proper ID gen
		},
		Sender: fmt.Sprintf("%s:%s", localIP, SERVER_PORT),
	}

	err = s.BroadcastNetwork(msg, 8080) // TODO: don't hardcode
	if err != nil {
		log.Println(err.Error())
	}
}

// BroadcastNetwork sends a message to the entire network (including peers we don't know yet)
func (s *MessageBus) BroadcastNetwork(msg Message, port int) error {
	// Encode the message
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(msg); err != nil {
		return fmt.Errorf("network broadcast encode error: %w", err)
	}
	msgData := buf.Bytes()

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("get interfaces error: %w", err)
	}

	// Broadcast on each eligible interface
	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}

			// Create broadcast address for this network
			broadcastIP := net.IP(make([]byte, 4))
			for i := range ipNet.IP.To4() {
				broadcastIP[i] = ipNet.IP.To4()[i] | ^ipNet.Mask[i]
			}
			broadcastAddr := &net.UDPAddr{
				IP:   broadcastIP,
				Port: port,
			}

			// Send broadcast packet
			_, err := s.udpConn.WriteToUDP(msgData, broadcastAddr)
			if err != nil {
				log.Printf("network broadcast failed on %s: %v", iface.Name, err)
			}
		}
	}
	log.Println("Succesfully broadcasted to the network")
	return nil
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80") // Use a public IP (won't actually connect)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
