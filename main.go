package main

import (
	"fmt"
	"os"
)

func augmentGraph(graph Graph, positions []Point) PosGraph {
	out := make([]PosNode, len(graph))
	for i, u := range graph {
		out[i].X = float32(positions[i].X)
		out[i].Y = float32(positions[i].Y)
		out[i].Edges = u
	}
	return out
}

func errexit(message string) {
	fmt.Println(message)
	os.Exit(1)
}

func main() {
	drawGui := true
	var filename string
	filenameSet := false
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i][0] == '-' {
			if os.Args[i][1:] == "png" {
				drawGui = false
			}
		} else {
			if filenameSet {
				errexit("Only one filename argument allowed.")
			} else {
				filename = os.Args[i]
				filenameSet = true
			}
		}
	}
	if !filenameSet {
		errexit("Usage: ppa-final [-png] <input-file>")
	}

	graph, err := buildGraphFromFile(filename)
	if err != nil {
		errexit(fmt.Sprintf("Error building graph: %v\n", err))
	}

	positions := ForceDirectedLayout(graph, 100, 800., 600.)
	for node, pos := range positions {
		fmt.Printf("Node %d: (%.2f, %.2f)\n", node, pos.X, pos.Y)
	}
	outGraph := augmentGraph(graph, positions)

	if drawGui {
		RenderGUI(outGraph)
	} else {
		RenderPNG(outGraph)
	}
}
