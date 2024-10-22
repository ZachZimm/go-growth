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
			simulateNutrientGrowth()
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
			if randFloat <= 0.225 {
				// Add to the nutrient value
				if tiles[i][j].Type == 0 {
					tiles[i][j].Nutrient += 0.08
				} else if tiles[i][j].Type == 2 {
					tiles[i][j].Nutrient += 0.15
				}

				if tiles[i][j].Nutrient >= 0.5 {
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
				tiles[i][j].Nutrient -= (0.02 * (rand + 0.5))
				if tiles[i][j].Nutrient < 0.5 {
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
