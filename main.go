package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
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
	conn *net.TCPConn
	uid  string
}

func newPeer(conn *net.TCPConn, uid string) *Peer {
	return &Peer{uid: conn.RemoteAddr().String(), conn: conn}
}

func (p *Peer) Handle() {

	var conn *net.TCPConn = p.conn

	log.Printf("Handling for sender: %s over %s", conn.RemoteAddr().String(), conn.RemoteAddr().Network())

	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Connection closed: %v", err)
			return
		}

		msg := string(buffer[:n])
		log.Printf("[RECEIVED] %s", msg)

		time := time.Now().Format(time.ANSIC)
		responseStr := fmt.Sprintf("Echo: %v @ %v", msg, time)
		conn.Write([]byte(responseStr))
	}

}

type Server struct {
	peers []*Peer
}

type Client struct {
}

func InitClient() Client {
	return Client{}
}

func (c *Client) Message(message string, address *net.TCPAddr) {
	conn, err := net.DialTCP(SERVER_TYPE, nil, address)
	if err != nil {
		log.Printf("Dial failed: %s", err.Error())
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message))
	if err != nil {
		log.Printf("Write data failed: %s", err.Error())
		return
	}
	log.Printf("Sent message: %s", message)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Read response failed: %s", err.Error())
		return
	}

	response := string(buffer[:n])
	log.Printf("Received response: %s", response)
}

func (s Server) Start() {

	listen, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	log.Printf("Started listening server at %s:%s", SERVER_HOST, SERVER_PORT)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer listen.Close()

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

		peer := newPeer(tcpConn, "placeholder")
		s.peers = append(s.peers, peer)
		go peer.Handle()

	}

}

func newServer() Server {
	return Server{peers: make([]*Peer, 32)}
}

func main() {
	// Run the TCP server in the background
	go func() {
		server := newServer()
		server.Start()
	}()

	// Wait before the client tries to connect
	time.Sleep(1 * time.Second)

	// Create the TCP client connection
	tcpServer, err := net.ResolveTCPAddr(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		log.Printf("Error resolving address: %s", err.Error())
		return
	}
	client := InitClient()
	client.Message("my message!", tcpServer)

	// Start the HTTP server for the API
	http.HandleFunc("/ping", pingHandler)

	// Serve the Swagger documentation JSON/YAML
	http.Handle("/docs/", http.StripPrefix("/docs", http.FileServer(http.Dir("./docs"))))

	// Start HTTP server on port 8081
	log.Println("HTTP API listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))

	// Block the main goroutine forever
	select {}
}
