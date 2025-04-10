package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Kaya-Sem/ouroboros/src"
)

func main() {
	log.SetFlags(0)
	log.Printf("Client: %s", src.GetClientName())
	src.InitializeInterfaces()

	messageBus := src.NewMessageBus()

	go messageBus.Start()

	messageBus.AnnouncePresence()

	time.Sleep(1 * time.Second)

	http.HandleFunc("/peers", messageBus.PeersHandler)

	http.Handle("/docs/", http.StripPrefix("/docs", http.FileServer(http.Dir("./docs"))))

	log.Printf("HTTP API listening on :%s", messageBus.HttpPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", messageBus.HttpPort), nil))
}
