package main

import (
	"fmt"
	"os"
	"time"
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

func endPhase(phaseName string, phaseStart *time.Time) {
	phaseEnd := time.Now()
	fmt.Printf("%s: %d ns\n", phaseName, phaseEnd.Sub(*phaseStart).Nanoseconds())
	*phaseStart = phaseEnd
}

func main() {
	phaseStart := time.Now()
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
	endPhase("Parse command line", &phaseStart)

	graph, err := buildGraphFromFile(filename)
	if err != nil {
		errexit(fmt.Sprintf("Error building graph: %v\n", err))
	}
	endPhase("Build graph", &phaseStart)

	positions := ForceDirectedLayout(graph, 10000, 800., 600.)
	/*
	for node, pos := range positions {
		fmt.Printf("Node %d: (%.2f, %.2f)\n", node, pos.X, pos.Y)
	}
	*/
	endPhase("Compute layout", &phaseStart)

	outGraph := augmentGraph(graph, positions)

	if drawGui {
		RenderGUI(outGraph)
	} else {
		RenderPNG(outGraph)
		endPhase("Create PNG", &phaseStart)
	}
}
