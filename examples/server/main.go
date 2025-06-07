package main

import (
	"fmt"
	"log"

	"github.com/AaravShirvoikar/wsx"
)

func main() {
	addr := "127.0.0.1:9001"
	server := wsx.NewWebSocketServer(addr, EchoHandler)

	fmt.Printf("Server listening on %v\n", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func EchoHandler(wsconn *wsx.WSConn) {
	for {
		msg, err := wsconn.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading message from %v: %v\n", wsconn.Addr(), err)
			break
		}
		msgStr := msg.Payload.String()

		fmt.Printf("Received message from %v: %s\n", wsconn.Addr(), msgStr)

		if err := wsconn.SendMessage(msg.Opcode, []byte(msgStr)); err != nil {
			fmt.Printf("Error sending echo to %v: %v\n", wsconn.Addr(), err)
			break
		}

		fmt.Printf("Echoed message back to %v\n", wsconn.Addr())
	}
}
