package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gorilla/websocket"
)

var (
	connectionStatus string = "Disconnected"
)

var (
	cameraX float32 = 0
	cameraY float32 = 0
)

// create tilesWide and tilesHigh constants
const (
	tilesWide              = 80 * 3
	tilesHigh              = 45 * 3
	windowWidth            = 1200
	windowHeight           = 675
	tileSizeX      float32 = 15
	tileSizeY      float32 = 15
	tilesOnScreenX         = float32(windowWidth) / tileSizeX
	tilesOnScreenY         = float32(windowHeight) / tileSizeY
)

func sendUpdateTile(wsConn *websocket.Conn, x, y, value int) error {
	if wsConn == nil {
		return fmt.Errorf("WebSocket connection is nil")
	}
	// Create the message as a map
	msg := map[string]interface{}{
		"type":  "updateTile",
		"x":     x,
		"y":     y,
		"value": value,
	}
	// Serialize the message to JSON
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Send the JSON message over the WebSocket
	err = wsConn.WriteMessage(websocket.TextMessage, msgJSON)
	if err != nil {
		return err
	}

	return nil
}

func sendResetTiles(wsConn *websocket.Conn) error {
	if wsConn == nil {
		return fmt.Errorf("WebSocket connection is nil")
	}
	// Create the message as a map
	msg := map[string]interface{}{
		"type": "resetTiles",
	}
	// Serialize the message to JSON
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Send the JSON message over the WebSocket
	err = wsConn.WriteMessage(websocket.TextMessage, msgJSON)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	rl.InitWindow(windowWidth, windowHeight, "Raylib WebSocket Client")
	defer rl.CloseWindow()
	// create an 80x45 array of tiles
	var tiles [tilesWide][tilesHigh]int

	rl.SetTargetFPS(120)

	// WebSocket connection setup
	var wsConn *websocket.Conn
	var err error
	var wsUrl string = "ws://localhost:8152/ws"
	var loggedIn bool = false

	// Start a goroutine to handle the WebSocket connection
	go func() {
		for {
			// Attempt to connect to the WebSocket server
			wsConn, _, err = websocket.DefaultDialer.Dial(wsUrl, nil)
			if err != nil {
				log.Println("Connection failed:", err)
				connectionStatus = "Disconnected"
				time.Sleep(5 * time.Second) // Wait before retrying
				continue
			}
			connectionStatus = "Connected"
			log.Println("Connected to WebSocket server")
			if !loggedIn {
				// Send a login message
				loginMessage := []byte(`{"type":"login","username":"raylib"}`)
				err = wsConn.WriteMessage(websocket.TextMessage, loginMessage)
				if err != nil {
					log.Println("Write error:", err)
					connectionStatus = "Disconnected"
					break
				}
				loggedIn = true

			}

			// Read messages in a loop
			for {
				_, message, err := wsConn.ReadMessage()
				if err != nil {
					log.Println("Read error:", err)
					connectionStatus = "Disconnected"
					break
				}
				// log.Printf("Received: %s\n", message)
				// try to parse the message as a tile update, and update the tiles array
				// the message is a JSON object with a "type" field whych may be "tiles", and if so it has a "tiles" field which is the actual tile data
				var msg map[string]interface{}
				err = json.Unmarshal(message, &msg)
				if err != nil {
					log.Println("JSON unmarshal error:", err)
					continue
				}
				if msg["type"] == "tiles" { // update the tiles array
					// the "tiles" field is a 2d array of int of size 80x45
					tileData, ok := msg["tiles"].([]interface{})
					if !ok {
						log.Println("Invalid tiles format")
						continue
					}
					// set tiles = tileData
					for x := 0; x < tilesWide; x++ {
						for y := 0; y < tilesHigh; y++ {
							tiles[x][y] = int(tileData[x].([]interface{})[y].(float64))
						}
					}

				} else {
					log.Println("Received message of unknown type:", msg)
				}
			}

			// Close the connection and retry
			wsConn.Close()
			time.Sleep(5 * time.Second)
		}
	}()

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		// rl.ClearBackground(rl.RayWhite)

		// Display the connection status
		rl.DrawText("WebSocket Client", 10, 10, 20, rl.DarkGray)

		// Calculate the range of tiles to draw
		tileXStart := int(math.Floor(float64(cameraX)))
		tileYStart := int(math.Floor(float64(cameraY)))
		tileXEnd := int(math.Ceil(float64(cameraX + tilesOnScreenX)))
		tileYEnd := int(math.Ceil(float64(cameraY + tilesOnScreenY)))

		// Clamp the tile indices to valid ranges
		if tileXStart < 0 {
			tileXStart = 0
		}
		if tileYStart < 0 {
			tileYStart = 0
		}
		if tileXEnd > tilesWide {
			tileXEnd = tilesWide
		}
		if tileYEnd > tilesHigh {
			tileYEnd = tilesHigh
		}

		// Draw the tiles
		for x := tileXStart; x < tileXEnd; x++ {
			for y := tileYStart; y < tileYEnd; y++ {
				tileValue := tiles[x][y]
				var tileColor rl.Color
				// Determine the color based on tileValue
				switch tileValue {
				case 0:
					tileColor = rl.Black
				case 1:
					tileColor = rl.RayWhite
				case 2:
					tileColor = rl.Green
				case 3:
					tileColor = rl.Brown
				case 4:
					tileColor = rl.Blue
				}

				screenX := (float32(x) - cameraX) * tileSizeX
				screenY := (float32(y) - cameraY) * tileSizeY

				rl.DrawRectangle(int32(screenX), int32(screenY), int32(tileSizeX), int32(tileSizeY), tileColor)
			}
		}

		// Handle mouse input to update tiles
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && connectionStatus == "Connected" {
			mouseX := rl.GetMouseX()
			mouseY := rl.GetMouseY()
			tileX := int(cameraX + float32(mouseX)/tileSizeX)
			tileY := int(cameraY + float32(mouseY)/tileSizeY)

			// Ensure the tile coordinates are within bounds
			if tileX >= 0 && tileX < tilesWide && tileY >= 0 && tileY < tilesHigh {
				currentValue := tiles[tileX][tileY]
				var newValue int

				// Determine the new value based on the current tile color
				if currentValue == 0 {
					newValue = 1 // If black, change to white
				} else {
					newValue = 0 // If white, change to black
				}

				tiles[tileX][tileY] = newValue

				// Send the updateTile message to the server
				err := sendUpdateTile(wsConn, tileX, tileY, newValue)
				if err != nil {
					log.Println("Error sending updateTile message:", err)
				}
			}
		}

		moveSpeed := 5.0 / tileSizeX // Adjust as needed

		if rl.IsKeyDown(rl.KeyLeft) {
			cameraX -= moveSpeed
		}
		if rl.IsKeyDown(rl.KeyRight) {
			cameraX += moveSpeed
		}
		if rl.IsKeyDown(rl.KeyUp) {
			cameraY -= moveSpeed
		}
		if rl.IsKeyDown(rl.KeyDown) {
			cameraY += moveSpeed
		}

		// Clamp camera position
		if cameraX < 0 {
			cameraX = 0
		}
		if cameraY < 0 {
			cameraY = 0
		}
		maxCameraX := float32(tilesWide) - tilesOnScreenX
		if cameraX > maxCameraX {
			cameraX = maxCameraX
		}
		maxCameraY := float32(tilesHigh) - tilesOnScreenY
		if cameraY > maxCameraY {
			cameraY = maxCameraY
		}

		if rl.IsKeyPressed(rl.KeyR) && connectionStatus == "Connected" {
			err := sendResetTiles(wsConn)
			if err != nil {
				log.Println("Error sending resetTiles message:", err)
			}
		}

		statusText := "Status: " + connectionStatus
		statusColor := rl.Red
		if connectionStatus == "Connected" {
			statusColor = rl.Green
		}
		rl.DrawText(statusText, 10, 40, 20, statusColor)

		rl.EndDrawing()
	}

	// Clean up the WebSocket connection on exit
	if wsConn != nil {
		wsConn.Close()
	}
}
