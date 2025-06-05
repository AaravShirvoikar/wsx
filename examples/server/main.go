package main

import (
	"fmt"
	"log"

	"github.com/AaravShirvoikar/wsx"
)

func main() {
	addr := "localhost:9001"
	server := wsx.NewWebSocketServer(addr)

	fmt.Printf("Server listening on %v\n", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
