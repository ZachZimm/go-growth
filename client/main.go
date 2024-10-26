package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gorilla/websocket"
)

// type Config interface{}
type Config struct {
	WindowWidth    int     `json:"windowWidth"`
	WindowHeight   int     `json:"windowHeight"`
	TileSizeX      float32 `json:"tileSizeX"`
	TileSizeY      float32 `json:"tileSizeY"`
	TilesOnScreenX float32 `json:"tilesOnScreenX"`
	TilesOnScreenY float32 `json:"tilesOnScreenY"`
	WsUrl          string  `json:"wsUrl"`
}

func NewConfig() Config {
	configPath := "config.json"
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
		fmt.Println("Error opening config file.")
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
		fmt.Println("Error decoding config file.")
	}

	config.TilesOnScreenX = float32(config.WindowWidth) / config.TileSizeX
	config.TilesOnScreenY = float32(config.WindowHeight) / config.TileSizeY
	return config
}

var (
	connectionStatus string = "Disconnected"
	configuration    Config
)

var (
	cameraX float32 = 0
	cameraY float32 = 0
)

// create tilesWide and tilesHigh constants
const (
	tilesWide = 80 * 10 // This should come from the server on connection
	tilesHigh = 45 * 10 // This too
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
	// Load configuration
	configuration = NewConfig()
	rl.InitWindow(int32(configuration.WindowWidth), int32(configuration.WindowHeight), "Raylib WebSocket Client")
	defer rl.CloseWindow()
	// create an 80x45 array of tiles
	var tiles [tilesWide][tilesHigh]int

	rl.SetTargetFPS(120)

	// WebSocket connection setup
	var wsConn *websocket.Conn
	var err error
	var loggedIn bool = false

	// Start a goroutine to handle the WebSocket connection
	go func() {
		for {
			// Attempt to connect to the WebSocket server
			wsConn, _, err = websocket.DefaultDialer.Dial(configuration.WsUrl, nil)
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

	var groundColor = rl.NewColor(176, 143, 28, 255)
	var nutrientColor = rl.NewColor(16, 144, 16, 255)
	var oilColor = rl.NewColor(64, 64, 64, 255)
	var concreteColor = rl.NewColor(128, 128, 128, 255)
	var lowlandsColor = rl.NewColor(177, 144, 103, 255)
	var mountainColor = rl.NewColor(140, 141, 139, 255)
	var highMountainColor = rl.NewColor(202, 215, 215, 255)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		// Display the connection status
		rl.DrawText("WebSocket Client", 10, 10, 20, rl.DarkGray)

		// Calculate the range of tiles to draw
		tileXStart := int(math.Floor(float64(cameraX)))
		tileYStart := int(math.Floor(float64(cameraY)))
		tileXEnd := int(math.Ceil(float64(cameraX + configuration.TilesOnScreenX)))
		tileYEnd := int(math.Ceil(float64(cameraY + configuration.TilesOnScreenY)))

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
					// tileColor = rl.Black
					tileColor = groundColor
				case 1:
					tileColor = highMountainColor
				case 2:
					// tileColor = rl.Green
					tileColor = nutrientColor
				case 3:
					tileColor = mountainColor
				case 4:
					tileColor = rl.Blue
				case 5:
					tileColor = oilColor
				case 6:
					tileColor = concreteColor
				case 7:
					tileColor = lowlandsColor
				}

				screenX := (float32(x) - cameraX) * configuration.TileSizeX
				screenY := (float32(y) - cameraY) * configuration.TileSizeY

				rl.DrawRectangle(int32(screenX), int32(screenY), int32(configuration.TileSizeX), int32(configuration.TileSizeY), tileColor)
			}
		}

		// Handle mouse input to update tiles
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && connectionStatus == "Connected" {
			mouseX := rl.GetMouseX()
			mouseY := rl.GetMouseY()
			tileX := int(cameraX + float32(mouseX)/configuration.TileSizeX)
			tileY := int(cameraY + float32(mouseY)/configuration.TileSizeY)

			// Ensure the tile coordinates are within bounds
			if tileX >= 0 && tileX < tilesWide && tileY >= 0 && tileY < tilesHigh {
				currentValue := tiles[tileX][tileY]
				var newValue int

				// Determine the new value based on the current tile color
				if currentValue == 1 {
					newValue = 0 // If white, change to black
				} else {
					newValue = 1 // Otherwise, change to white
				}

				// tiles[tileX][tileY] = newValue

				// Send the updateTile message to the server
				err := sendUpdateTile(wsConn, tileX, tileY, newValue)
				if err != nil {
					log.Println("Error sending updateTile message:", err)
				}
			}
		}

		moveSpeed := 500.0 / configuration.TileSizeX // Adjust as needed
		var speedMultiplier float32 = 1.0
		if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
			speedMultiplier = 2.0
		}
		if rl.IsKeyDown(rl.KeyLeft) || rl.IsKeyDown(rl.KeyA) {
			cameraX -= moveSpeed * speedMultiplier * rl.GetFrameTime()
		}
		if rl.IsKeyDown(rl.KeyRight) || rl.IsKeyDown(rl.KeyD) {
			cameraX += moveSpeed * speedMultiplier * rl.GetFrameTime()
		}
		if rl.IsKeyDown(rl.KeyUp) || rl.IsKeyDown(rl.KeyW) {
			cameraY -= moveSpeed * speedMultiplier * rl.GetFrameTime()
		}
		if rl.IsKeyDown(rl.KeyDown) || rl.IsKeyDown(rl.KeyS) {
			cameraY += moveSpeed * speedMultiplier * rl.GetFrameTime()
		}

		if rl.IsKeyPressed(rl.KeyPageUp) || rl.IsKeyPressed(rl.KeyEqual) || rl.IsKeyPressed(rl.KeyKpAdd) {
			if configuration.TileSizeX < 128 && configuration.TileSizeY < 128 {
				configuration.TileSizeX += 1
				configuration.TileSizeY += 1

				configuration.TilesOnScreenX = float32(rl.GetScreenWidth()) / configuration.TileSizeX
				configuration.TilesOnScreenY = float32(rl.GetScreenHeight()) / configuration.TileSizeY
			}
		}

		if rl.IsKeyPressed(rl.KeyPageDown) || rl.IsKeyPressed(rl.KeyMinus) || rl.IsKeyPressed(rl.KeyKpSubtract) {
			if configuration.TileSizeX > 1 && configuration.TileSizeY > 1 {
				configuration.TileSizeX -= 1
				configuration.TileSizeY -= 1

				configuration.TilesOnScreenX = float32(rl.GetScreenWidth()) / configuration.TileSizeX
				configuration.TilesOnScreenY = float32(rl.GetScreenHeight()) / configuration.TileSizeY
			}
		}

		// Clamp camera position
		if cameraX < 0 {
			cameraX = 0
		}
		if cameraY < 0 {
			cameraY = 0
		}
		maxCameraX := float32(tilesWide) - configuration.TilesOnScreenX
		if cameraX > maxCameraX {
			cameraX = maxCameraX
		}
		maxCameraY := float32(tilesHigh) - configuration.TilesOnScreenY
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
