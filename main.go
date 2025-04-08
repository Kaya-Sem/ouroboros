package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Kaya-Sem/ouroboros/src"
)

func main() {
	log.SetFlags(0)
	log.Printf("Client: %s", src.GetClientName())
	src.InitializeInterfaces()

	server := src.NewServer()

	go server.Start()

	time.Sleep(1 * time.Second)

	http.HandleFunc("/peers", server.PeersHandler)

	// Serve Swagger documentation
	http.Handle("/docs/", http.StripPrefix("/docs", http.FileServer(http.Dir("./docs"))))

	log.Println("HTTP API listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
