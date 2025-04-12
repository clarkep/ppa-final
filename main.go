package main

import (
	"fmt"
	"os"
)

func convertGraph (graph *Graph, positions map[int]Point) PosGraph { 
	keysToIndices := make(map[int]int, len(graph.adjList))
	i := 0
	for k, _ := range graph.adjList {
		keysToIndices[k] = i;
		i += 1
	}
	nodes := make([]PosNode, len(graph.adjList))
	for k, v := range graph.adjList {
		ki := keysToIndices[k]
		nodes[ki].X = float32(positions[k].X)
		nodes[ki].Y = float32(positions[k].Y)
		for _, neighbor_key := range v {
			nodes[ki].Edges = append(nodes[ki].Edges, keysToIndices[neighbor_key])
		}
	}
	return nodes
}

func errexit (message string) {
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
				drawGui = false;
			}
		} else {
			if filenameSet {
				errexit("Only one filename argument allowed.")
			} else {
				filename = os.Args[i]
				filenameSet = true;
			}
		}
	}
	if !filenameSet {
        errexit("Usage: go run graph.go <input-file>")
    }

    graph, err := buildGraphFromFile(filename)
    if err != nil {
        errexit(fmt.Sprintf("Error building graph: %v\n", err))
    }

	positions := ForceDirectedLayout(graph, 100, 800., 600.)
    for node, pos := range positions {
        fmt.Printf("Node %d: (%.2f, %.2f)\n", node, pos.X, pos.Y)
    }
	outGraph := convertGraph(graph, positions)

	if drawGui {
		RenderGUI(outGraph)
	} else {
		RenderPNG(outGraph)
	}
}
