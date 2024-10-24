// simulation.go
package main

import (
	"math/rand"
	"time"
)

type Tile struct {
	Type     int
	Nutrient float64
}

func runSimulation() {
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			simulateNutrientDecay()
			simulateWaterNutrition()
			simulateInorganicNutrientDecay()
			simulateNutrientGrowth()
		}
	}
}

func simulateWaterNutrition() {
	// create a map containing the set of tiles in (waterNearby or waterNearby2) and (nutrientTiles or nutrientsNearby)
	// then iterate over this list and add nutrients according to whether the tile is in waterNearby or waterNearby2

	wateredNutrientValue := 0.045
	wateredNutrientNearbyValue := 0.03

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
			if randFloat <= 0.55 {
				randFloat = rand.Float64()
				// Add to the nutrient value
				if tiles[i][j].Type == 0 {
					tiles[i][j].Nutrient += 0.08 * (randFloat + 0.4)
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

func simulateNutrientDecay() {
	for i := 0; i < tilesWide; i++ {
		for j := 0; j < tilesHigh; j++ {
			rand := rand.Float64()
			// Check if the tile is a nutrient tile and randomly decay it
			if (tiles[i][j].Type == 2 || tiles[i][j].Type == 0) && (int(rand*100)%2) == 0 {
				// Decrease the nutrient value
				// changing the rand offset between 0.45 and 0.55 seems like a good way of varying environment conditions for now
				tiles[i][j].Nutrient -= (0.055 * (rand + 0.51))
				if tiles[i][j].Nutrient < nutrientGreenCutOff {
					tiles[i][j].Type = 0 // Tile becomes ground
					if tiles[i][j].Nutrient < 0.0 {
						tiles[i][j].Nutrient = 0.0
					}
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
