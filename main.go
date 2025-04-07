package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

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
	go func() {
		server := newServer()
		server.Start()
	}()

	time.Sleep(1 * time.Second)

	tcpServer, err := net.ResolveTCPAddr(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		log.Printf("Error resolving address: %s", err.Error())
		return
	}
	client := InitClient()
	client.Message("my message!", tcpServer)

	fmt.Println("Server is running. Press Ctrl+C to exit.")
	select {} // This blocks forever until program is interrupted
}
