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

type Server struct {
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
	// close listener
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		go handleRequest(conn)
		log.Printf("Sender: %s over %s", conn.RemoteAddr().String(), conn.RemoteAddr().Network())
	}
}

func InitServer() Server {
	return Server{}
}

func main() {
	go func() {
		server := InitServer()
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

func handleRequest(conn net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	time := time.Now().Format(time.ANSIC)
	responseStr := fmt.Sprintf("Your message is: %v. Received time: %v", string(buffer[:]), time)
	conn.Write([]byte(responseStr))

	// close conn
	conn.Close()
}
