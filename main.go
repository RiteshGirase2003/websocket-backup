package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"math/rand"
	"time"
	"github.com/gorilla/websocket"
)



var initiator = []string{}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func generateRandomString() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 4)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func getUniqueString() string {
	for {
		newStr := generateRandomString()
		exists := false
		for _, s := range initiator {
			if s == newStr {
				exists = true
				break
			}
		}
		if !exists {
			return newStr
		}
	}
}




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


func GenSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := getUniqueString()
	response := map[string]string{"session_id": sessionID}

	genLink(sessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func genLink(code string ) {
	api := fmt.Sprintf("/ws/%s", code)
	http.HandleFunc(api, handleConnections)
}


func main() {
	http.HandleFunc("/genSession", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		GenSession(w, r)
	})

	port := ":9218"
	fmt.Println("Server running on http://localhost" + port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
