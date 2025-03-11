package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
	"encoding/json"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Connected clients
var clients = make(map[*websocket.Conn]bool)

// Broadcast channel
var broadcast = make(chan Message)

// Message structure
type Message struct {
	Content  string `json:"content"`
	IsSender bool   `json:"isSender"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Register client
	clients[conn] = true

	// Listen for messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			delete(clients, conn)
			break
		}

		// Send message to all clients
		for client := range clients {
			isSender := (client == conn) // True if sender
			message := Message{Content: string(msg), IsSender: isSender}
			messageJSON, _ := json.Marshal(message)
			client.WriteMessage(websocket.TextMessage, messageJSON)
		}
	}
}

func genLink() string {
	
}


func main() {
	http.HandleFunc("/ws", handleConnections)



	// Start server
	port := ":9218"
	fmt.Println("Server running on http://localhost" + port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
