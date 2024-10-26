package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	tilesWide = 80 * 10
	tilesHigh = 45 * 10
)

var interval float64 = 0.125

var updateInterval = time.Duration(interval * float64(time.Second))
var tiles [tilesWide][tilesHigh]Tile // 80x45 grid of tiles

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

// Sends tile updates to a connected client
func sendTileUpdates(conn *websocket.Conn) {
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// Convert the tiles to a JSON object with type="tiles"
			simplifiedTiles := make([][]int, tilesWide)
			for i := 0; i < tilesWide; i++ {
				simplifiedTiles[i] = make([]int, tilesHigh)
				for j := 0; j < tilesHigh; j++ {
					simplifiedTiles[i][j] = tiles[i][j].Type
				}
			}

			// Send the JSON to the client
			tilesJson, err := json.Marshal(map[string]interface{}{"type": "tiles", "tiles": simplifiedTiles})
			if err != nil {
				fmt.Println("JSON marshal error:", err)
				return
			}

			// Send the JSON to the client
			err = conn.WriteMessage(websocket.TextMessage, tilesJson)
			if err != nil {
				fmt.Println("Write error:", err)
				return
			}

			// fmt.Println("Sent tiles JSON to client")
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Client connected:", conn.RemoteAddr())

	// Start a goroutine to send tile updates to the client
	go sendTileUpdates(conn)

	for {
		// Read message from client
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}

		// Handle client message (if necessary)
		fmt.Printf("Received from client: %s\n", message)
		// parse a JSON message from the client
		var msg map[string]interface{}
		err = json.Unmarshal(message, &msg)
		if err != nil {
			fmt.Println("JSON unmarshal error:", err)
			continue
		}

		if msg["type"] == "login" {
			username, ok := msg["username"].(string)
			if !ok {
				fmt.Println("Invalid username format")
				continue
			}
			fmt.Println("Received login message from client:", username)
		}

		if msg["type"] == "updateTile" {
			x, ok := msg["x"].(float64)
			if !ok {
				fmt.Println("Invalid x format")
				continue
			}
			y, ok := msg["y"].(float64)
			if !ok {
				fmt.Println("Invalid y format")
			}
			tile, ok := msg["value"].(float64)
			if !ok {
				fmt.Println("Invalid value format")
			}
			if x < 0 || x >= tilesWide || y < 0 || y >= tilesHigh {
				fmt.Println("Invalid tile coordinates")
				continue
			}
			tiles[int(x)][int(y)].Type = int(tile)
		}

		if msg["type"] == "resetTiles" {
			resetSimulation()
		}
	}

	fmt.Println("Client disconnected:", conn.RemoteAddr())
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	go resetSimulation()
	// go runSimulation()

	fmt.Println("WebSocket server starting on :8152")
	err := http.ListenAndServe(":8152", nil)
	if err != nil {
		fmt.Println("ListenAndServe error:", err)
	}
}
