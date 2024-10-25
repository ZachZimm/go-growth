// simulation.go
package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type Tile struct {
	Type     int
	Nutrient float64
}

const (
	ticksPerCycle = 480.0 // 1 minute if 8 ticks per second

	nutrientRate        = 0.0015
	nutrientGreenCutOff = 0.18

	inorganicRate = 0.004
	waterRate     = 0.001
	oilspoutRate  = 0.001
)

// Much of this should be moved to the simulation.go file
var nutrientsNearby = make(map[[2]int]struct{})
var nutrientTiles = make(map[[2]int]struct{})
var waterTiles = make(map[[2]int]struct{})
var waterNearby = make(map[[2]int]struct{})
var waterNearby2 = make(map[[2]int]struct{})
var inorganicTiles = make(map[[2]int]struct{})
var inorganicNearby = make(map[[2]int]struct{})
var inorganicNearby2 = make(map[[2]int]struct{})
var oilspoutTiles = make(map[[2]int]struct{})
var oilspoutNearby = make(map[[2]int]struct{})

func resetNutrientsMaps() {
	nutrientsNearby = make(map[[2]int]struct{})
	nutrientTiles = make(map[[2]int]struct{})
	waterTiles = make(map[[2]int]struct{})
	waterNearby = make(map[[2]int]struct{})
	waterNearby2 = make(map[[2]int]struct{})
}

// Set all tiles to 0
func initTiles() {
	for i := 0; i < tilesWide; i++ {
		for j := 0; j < tilesHigh; j++ {
			tiles[i][j].Type = 0
			tiles[i][j].Nutrient = 0.09
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
							nutrientsNearby[[2]int{i + x, j + y}] = struct{}{}
						}
					}
				}
			}
		}
	}
}

func addStartingPlatform() {
	// The starting platform is a 5x5 square (missing outermost corners) of concrete tiles
	// concrete tiles are represented by a value of 6
	randX := rand.Intn(tilesWide - 7)
	randY := rand.Intn(tilesHigh - 7)
	randX = 10
	randY = 10
	for i := randX; i < randX+7; i++ {
		for j := randY; j < randY+7; j++ {
			// check if the tile is an outermost corner
			if !((i == randX || i == randX+6) && (j == randY || j == randY+6)) {
				tiles[i][j].Type = 6
			}
		}
	}
	fmt.Printf("Starting platform at (%d, %d)\n", randX, randY)
}

func addOilspouts() {
	// Oilspouts are represented by a value of 5
	for i := 0; i < tilesWide; i++ {
		for j := 0; j < tilesHigh; j++ {
			randFloat := rand.Float64()
			if tiles[i][j].Type == 0 && randFloat <= oilspoutRate {
				tiles[i][j].Type = 5
				oilspoutTiles[[2]int{i, j}] = struct{}{}

				// loop through the 8 surrounding tiles and add them to the oilspoutNearby list
				for x := -1; x <= 1; x++ {
					for y := -1; y <= 1; y++ {
						if i+x >= 0 && i+x < tilesWide && j+y >= 0 && j+y < tilesHigh {
							oilspoutNearby[[2]int{i + x, j + y}] = struct{}{}
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
				inorganicTiles[[2]int{i, j}] = struct{}{}
			}
		}
	}

	// loop through the 8 surrounding tiles and add them to the inorganicNearby list
	// additionally, add tiles 2 away (minus the corners) to the inorganicNearby2 list
	for coord := range inorganicTiles {
		i := coord[0]
		j := coord[1]
		for x := -1; x <= 1; x++ {
			for y := -1; y <= 1; y++ {
				if i+x >= 0 && i+x < tilesWide && j+y >= 0 && j+y < tilesHigh {
					// Add the nearby inorganic tile to the inorganicNearby map
					if !(x == 0 && y == 0) {
						inorganicNearby[[2]int{i + x, j + y}] = struct{}{}
					}
				}
			}
		}

		// populate the inorganicNearby2 map without the corners and without repeating the tiles in the inorganicNearby
		for x := -2; x <= 2; x++ {
			for y := -2; y <= 2; y++ {
				if i+x >= 0 && i+x < tilesWide && j+y >= 0 && j+y < tilesHigh {
					if !(x == 0 && y == 0) && !(x == 2 && y == 2) && !(x == -2 && y == -2) && !(x == 2 && y == -2) && !(x == -2 && y == 2) {
						inorganicNearby2[[2]int{i + x, j + y}] = struct{}{}
					}
				}
			}
		}
	}
}

func populateWaterNearbyMap(iI int, jJ int) {
	// loop through the 8 tiles surrounding the given index and add them to the waterNearby list
	// additionally, add tiles 2 away (minus the corners) to the waterNearby2 list

	// populate only the waterNearby map
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			if iI+x >= 0 && iI+x < tilesWide && jJ+y >= 0 && jJ+y < tilesHigh {
				// Add the nearby water tile to the waterNearby map
				if !(x == 0 && y == 0) {
					waterNearby[[2]int{iI + x, jJ + y}] = struct{}{}
				}
			}
		}
	}

	// populate the waterNearby2 map without the corners and without repeating the tiles in the waterNearby map or the center tile
	for x := -2; x <= 2; x++ {
		for y := -2; y <= 2; y++ {
			if iI+x >= 0 && iI+x < tilesWide && jJ+y >= 0 && jJ+y < tilesHigh {
				// Add the nearby water tile to the waterNearby2 map
				if !(x == 0 && y == 0) && !(x == 2 && y == 2) && !(x == 2 && y == -2) && !(x == -2 && y == 2) && !(x == -2 && y == -2) {
					waterNearby2[[2]int{iI + x, jJ + y}] = struct{}{}
				}
			}
		}
	}

}

func addWaterPockets() {
	// Water is represented by a value of 4
	// water will be added in pockets of size 1 to 9 contiguous tiles
	waterTiles := make(map[[2]int]struct{})
	for i := 0; i < tilesWide; i++ {
		for j := 0; j < tilesHigh; j++ {
			randFloat := rand.Float64()
			if randFloat <= waterRate {
				// Add a pocket of water to the tiles
				pocketSizeX := rand.Intn(3) + 1
				pocketSizeY := rand.Intn(3) + 1
				for x := 0; x < pocketSizeX; x++ {
					for y := 0; y < pocketSizeY; y++ {
						if i+x >= 0 && i+x < tilesWide && j+y >= 0 && j+y < tilesHigh {
							randFloat2 := rand.Float64()
							if randFloat2 <= 0.75 {
								tiles[i+x][j+y].Type = 4
								waterTiles[[2]int{i + x, j + y}] = struct{}{}
							}
						}
					}
				}
			}
		}
	}
	// populate the waterNearby and waterNearby2 maps
	for key := range waterTiles {
		populateWaterNearbyMap(key[0], key[1])
	}

}

func simulateWaterNutrition() {
	// create a map containing the set of tiles in (waterNearby or waterNearby2) and (nutrientTiles or nutrientsNearby)
	// then iterate over this list and add nutrients according to whether the tile is in waterNearby or waterNearby2

	wateredNutrientValue := 0.05
	wateredNutrientNearbyValue := 0.035

	for coord := range waterNearby {
		if _, ok := nutrientTiles[coord]; ok {
			tiles[coord[0]][coord[1]].Nutrient += wateredNutrientValue
		}
		if _, ok := nutrientsNearby[coord]; ok {
			tiles[coord[0]][coord[1]].Nutrient += wateredNutrientNearbyValue
		}

	}
	for coord := range waterNearby2 {
		if _, ok := nutrientTiles[coord]; ok {
			tiles[coord[0]][coord[1]].Nutrient += wateredNutrientValue / 3
		}
		if _, ok := nutrientsNearby[coord]; ok {
			tiles[coord[0]][coord[1]].Nutrient += wateredNutrientNearbyValue / 3
		}
	}
}

func simulateInorganicNutrientDecay() {
	nearbyInorganicDecayValue := 0.02
	nearbyInorganicDecayValue2 := 0.01
	nearbyOilspoutDecayValue := 0.03
	for coord := range inorganicNearby {
		i, j := coord[0], coord[1]
		if tiles[i][j].Nutrient > 0 {
			tiles[i][j].Nutrient -= nearbyInorganicDecayValue
		}
	}

	for coord := range inorganicNearby2 {
		i, j := coord[0], coord[1]
		if tiles[i][j].Nutrient > 0 {
			tiles[i][j].Nutrient -= nearbyInorganicDecayValue2
		}
	}

	for coord := range oilspoutNearby {
		i, j := coord[0], coord[1]
		if tiles[i][j].Nutrient > 0 {
			tiles[i][j].Nutrient -= nearbyOilspoutDecayValue
		}
	}
}

func simulateNutrientGrowth() {
	newNutrients := make(map[[2]int]struct{})
	newNutrientsNearby := make(map[[2]int]struct{})

	for coord := range nutrientsNearby {
		i, j := coord[0], coord[1]
		if tiles[i][j].Type == 0 {
			randFloat := rand.Float64()
			if randFloat <= 0.5 {
				randFloat = rand.Float64()
				// Add to the nutrient value
				if tiles[i][j].Type == 0 {
					tiles[i][j].Nutrient += 0.083 * (randFloat + 0.4)
				} else if tiles[i][j].Type == 2 {
					tiles[i][j].Nutrient += 0.15 * (randFloat + 0.4)
				}

				if tiles[i][j].Nutrient >= nutrientGreenCutOff {
					tiles[i][j].Type = 2
					newNutrients[coord] = struct{}{}

					// Add empty neighbors to newNutrientsNearby
					for x := -1; x <= 1; x++ {
						for y := -1; y <= 1; y++ {
							ni, nj := i+x, j+y
							if ni >= 0 && ni < tilesWide && nj >= 0 && nj < tilesHigh {
								neighborCoord := [2]int{ni, nj}
								if tiles[ni][nj].Type == 0 || tiles[ni][nj].Type == 2 {
									newNutrientsNearby[neighborCoord] = struct{}{}
								}
							}
						}
					}
				}
			} else {
				// The tile didn't become a nutrient, keep it in newNutrientsNearby
				newNutrientsNearby[coord] = struct{}{}
			}
		}
	}

	// Update nutrientTiles
	for coord := range newNutrients {
		nutrientTiles[coord] = struct{}{}
	}

	// Update nutrientsNearby
	nutrientsNearby = newNutrientsNearby
}

func simulateNutrientDecay(cycleMultiplier float64) {
	for i := 0; i < tilesWide; i++ {
		// print the first 4 decimals of cycleMultiplier
		fmt.Printf("%.4f\n", cycleMultiplier)
		for j := 0; j < tilesHigh; j++ {
			rand := rand.Float64()
			// Check if the tile is a nutrient tile and randomly decay it
			if (tiles[i][j].Type == 2 || tiles[i][j].Type == 0) && (int(rand*100)%2) == 0 {
				// Decrease the nutrient value
				tiles[i][j].Nutrient -= (0.075 * rand) * (cycleMultiplier + 0.5)
				if tiles[i][j].Nutrient < 0.0 {
					tiles[i][j].Nutrient = 0.0
				}
				if tiles[i][j].Nutrient < nutrientGreenCutOff && tiles[i][j].Type == 2 {
					tiles[i][j].Type = 0 // Tile becomes ground
					// Remove the nutrient tile from the nutrientTiles map
					// and check if it should be removed from the nutrientsNearby map

					delete(nutrientTiles, [2]int{i, j})
					shouldRemove := true
					for x := -1; x <= 1; x++ {
						for y := -1; y <= 1; y++ {
							ni, nj := i+x, j+y
							if ni >= 0 && ni < tilesWide && nj >= 0 && nj < tilesHigh {
								neighborCoord := [2]int{ni, nj}
								if tiles[ni][nj].Type == 2 {
									nutrientsNearby[neighborCoord] = struct{}{}
									shouldRemove = false
								}
							}
						}
					}

					if shouldRemove {
						delete(nutrientsNearby, [2]int{i, j})
					}
				}
			}
		}
	}
}

func runSimulation() {
	var iterationsOfCycle float64 = math.Floor(rand.Float64() * ticksPerCycle)
	var iterationAddAmount float64 = 1
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cycleMultiplier := iterationsOfCycle / ticksPerCycle
			simulateNutrientDecay(cycleMultiplier)
			simulateWaterNutrition()
			simulateInorganicNutrientDecay()
			simulateNutrientGrowth()
		}

		if iterationsOfCycle == ticksPerCycle || iterationsOfCycle == 0 {
			iterationAddAmount = -iterationAddAmount
		}
		iterationsOfCycle += iterationAddAmount
	}
}

func resetSimulation() {
	initTiles()
	resetNutrientsMaps()
	addNutrients()
	addWaterPockets()
	addInorganics()
	addOilspouts()
	addStartingPlatform()
}
