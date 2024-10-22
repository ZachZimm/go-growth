package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	tilesWide = 80 * 3
	tilesHigh = 45 * 3

	nutrientRate  = 0.0015
	inorganicRate = 0.0035
)

var interval float64 = 0.125

var updateInterval = time.Duration(interval * float64(time.Second))
var tiles [tilesWide][tilesHigh]Tile // 80x45 grid of tiles

// Keep track of nutrient index tuples in a slice of [2]int
// As well as a list of tiles with nearby nutrients

// Much of this should be moved to the simulation.go file
var nutrientsNearby = make(map[[2]int]struct{})
var nutrientTiles = make(map[[2]int]struct{})

func resetNutrientsMaps() {
	nutrientsNearby = make(map[[2]int]struct{})
	nutrientTiles = make(map[[2]int]struct{})
}

// Set all tiles to 0
func initTiles() {
	for i := 0; i < tilesWide; i++ {
		for j := 0; j < tilesHigh; j++ {
			tiles[i][j].Type = 0
		}
	}
}

func addNutrients() {
	// Nutrients are represented by a value of 2
	for i := 0; i < tilesWide; i++ {
		for j := 0; j < tilesHigh; j++ {
			randFloat := rand.Float64()
			if tiles[i][j].Type == 0 && randFloat <= nutrientRate {
				tiles[i][j].Type = 2
				tiles[i][j].Nutrient = 1
				// Add the nutrient tile to the nutrientTiles map
				nutrientTiles[[2]int{i, j}] = struct{}{}

				// loop through the 8 surrounding tiles and add them to the nutrientsNearby list
				for x := -1; x <= 1; x++ {
					for y := -1; y <= 1; y++ {
						if i+x >= 0 && i+x < tilesWide && j+y >= 0 && j+y < tilesHigh {
							// Add the nearby nutrient tile to the nutrientsNearby map
							nutrientsNearby[[2]int{i + x, j + y}] = struct{}{}
						}
					}
				}
			}
		}
	}
}

func addInorganics() {
	// Rocks are represented by a value of 3
	for i := 0; i < tilesWide; i++ {
		for j := 0; j < tilesHigh; j++ {
			randFloat := rand.Float64()
			if tiles[i][j].Type == 0 && randFloat <= inorganicRate {
				tiles[i][j].Type = 3
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

// Function to send tile updates to a connected client
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
			initTiles()
			resetNutrientsMaps()
			addNutrients()
			addInorganics()
		}
	}

	fmt.Println("Client disconnected:", conn.RemoteAddr())
}

func main() {
	initTiles()
	resetNutrientsMaps() // This is redundant, but it's here for clarity
	addNutrients()
	addInorganics()
	go runSimulation()
	http.HandleFunc("/ws", wsHandler)

	fmt.Println("WebSocket server starting on :8152")
	err := http.ListenAndServe(":8152", nil)
	if err != nil {
		fmt.Println("ListenAndServe error:", err)
	}
}
