package main

import (
	"math/rand"
)

func SugiyamaLayout (graph Graph) []Point {
	// random for now
	positions := make([]Point, len(graph))
	for i, _ := range graph {
		positions[i] = Point{
			X: rand.Float64() * 500,
			Y: rand.Float64() * 500,
		}
	}
	return positions
}