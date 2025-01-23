package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// Message representerar ett meddelande från en klient
type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

// Upgrader för att hantera WebSocket-anslutningar
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Tillåt alla anslutningar (bra för utveckling)
	},
}

// Karta för att lagra anslutna klienter
var clients = make(map[*websocket.Conn]bool)

// Kanal för att hantera inkommande meddelanden
var broadcast = make(chan Message)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true
	fmt.Println("New client connected!")

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading message:", err)
			delete(clients, conn)
			break
		}

		// Debug: Skriv ut mottagna meddelanden
		fmt.Printf("Received message: %+v\n", msg)

		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		// Debug: Skriv ut meddelanden som skickas till klienter
		fmt.Printf("Broadcasting message: %+v\n", msg)

		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println("Error sending message:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	// WebSocket endpoint
	http.HandleFunc("/ws", handleConnections)

	// Starta en goroutine för att hantera meddelanden
	go handleMessages()

	// Starta servern
	fmt.Println("WebSocket server is running on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
