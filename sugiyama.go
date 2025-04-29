package main

import (
	"slices"
	"sort"
//	"math/rand"
)

func removeCycles(graph Graph) Graph {
	// Todo
	return graph
}

func allTrue(vertexSet []bool) bool {
	for i := range vertexSet {
		if !vertexSet[i] {
			return false
		}
	}
	return true
}

func assignLevels(graph Graph) [][]int {
	// the "longest path algorithm"
	out := make([][]int, 1)
	n := len(graph)
	U := make([]bool, n)
	Z := make([]bool, n)
	currentLayer := 0
	for !allTrue(U) {
	outerLoop:
		for i := range n {
			// select a vertex from V \ U with all outgoing edges in Z
			if !U[i] {
				selected := true
				for _, v := range graph[i] {
					if !Z[v] {
						selected = false
						break
					}
				}
				if selected {
					out[currentLayer] = append(out[currentLayer], i)
					U[i] = true
					// goto used as a "continue", but for the outer loop
					goto outerLoop
				}
			}
		}
		// none have been selected, so go up a layer
		currentLayer++
		out = append(out, make([]int, 0))
		// Z = Z union U
		for i := range n {
			if U[i] {
				Z[i] = true
			}
		}
	}
	return out
}

func orderLevels(graph Graph, levels [][]int) [][]int {
	for i := 1; i < len(levels); i++ {
		lvl := levels[i]
		order := make([][2]int, len(levels[i]))
		for j, v := range lvl {
			pos := 0
			for k, u := range levels[i - 1] {
				if slices.Contains(graph[u], v) {
					pos += k
				}
			}
			order[j][0] = pos
			order[j][1] = lvl[j]
		}
		sort.Slice(order, func (i, j int) bool {
			return order[i][0] < order[j][0]
		})
		for j := range order {
			levels[i][j] = order[j][1]
		}
	}
	return levels
}

func assignCoordinates(graph Graph, orders [][]int) []Point {
	out := make([]Point, len(graph))
	for x, lvl := range orders {
		n := len(lvl)
		for i, u := range lvl {
			// just assign coordinates based on (level, order in level)
			out[u] = Point{X: float64(len(orders) - x), Y: (100*float64(i+1))/float64(n+1)}
		}
	}
	return out
}

func SugiyamaLayout(graph Graph, iterations int) []Point {
	graph2 := removeCycles(graph)

	levels := assignLevels(graph2)

	orders := orderLevels(graph2, levels)

	positions := assignCoordinates(graph2, orders)

	return positions
}
