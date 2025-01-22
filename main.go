package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader för att hantera WebSocket-anslutningar
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Tillåt alla anslutningar (bra för utveckling)
	},
}

// Karta för att lagra anslutna klienter
var clients = make(map[*websocket.Conn]bool)

// Kanal för att hantera inkommande meddelanden
var broadcast = make(chan string)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// uppgradera http till websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	//Lägg till klienten i listan över anslutna klienter
	clients[conn] = true
	fmt.Println("New client connected!")

	for {
		//Läs meddelande från klienten
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			delete(clients, conn) // Ta bort klienten vid fel
			break
		}

		// Skicka meddelandet till broadcast-kanalen
		broadcast <- string(msg)
	}
}

func handleMessages() {
	for {
		// Vänta på meddelande från broadcast-kanalen
		msg := <-broadcast

		// Skicka meddelandet till alla anslutna klienter
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
	// webSocket endpoint
	http.HandleFunc("/ws", handleConnections)

	// Starta en goroutine för att hantera meddelanden
	go handleMessages()

	// Starta serverm
	fmt.Println("Websocket server is running on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
