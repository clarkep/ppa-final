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

func forceDirectedStd (graph Graph) []Point {
	return ForceDirectedLayout(graph, 10000, 800., 600.)
}

func forceDirectedParallelStd (graph Graph) []Point {
	return ForceDirectedLayoutParallel(graph, 10000, 800., 600., 1000)
}

func SugiyamaMain () {
	fmt.Printf("Not implemented yet.\n")
}

func main() {
	phaseStart := time.Now()
	drawGui := true
	var filename string
	filenameSet := false
	layoutFunc := forceDirectedStd
	directed := false
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i][0] == '-' {
			flag := os.Args[1][1:]
			switch flag {
			case "png" :
				drawGui = false
			case "l1":
				layoutFunc = forceDirectedStd
				directed = false
			case "l2":
				layoutFunc = forceDirectedParallelStd
				directed = false
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

	graph, err := buildGraphFromFile(filename, directed)
	if err != nil {
		errexit(fmt.Sprintf("Error building graph: %v\n", err))
	}
	endPhase("Build graph", &phaseStart)

	positions := layoutFunc(graph)
	endPhase("Compute layout", &phaseStart)

	outGraph := augmentGraph(graph, positions)

	if drawGui {
		RenderGUI(outGraph)
	} else {
		RenderPNG(outGraph)
		endPhase("Create PNG", &phaseStart)
	}
}
