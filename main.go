package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
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

func scaledTime(ns int64) string {
	if ns > 1000000000 {
		return fmt.Sprintf("%.3g s", float64(ns) / 1000000000.0)
	} else if ns > 1000000 {
		return fmt.Sprintf("%.3g ms", float64(ns) / 1000000.0)
	} else if ns > 1000 {
		return fmt.Sprintf("%.3g us", float64(ns) / 1000.0)
	} else {
		return fmt.Sprintf("%.3g ns", float64(ns))
	}
}

func endPhase(phaseName string, phaseStart *time.Time) {
	phaseEnd := time.Now()
	fmt.Printf("%s: %s\n", phaseName, scaledTime(phaseEnd.Sub(*phaseStart).Nanoseconds()))
	*phaseStart = phaseEnd
}

func forceDirectedStd(graph Graph, iterations int) []Point {
	return forceDirectedLayout(graph, iterations, 800., 600.)
}

func forceDirectedParallelStd(graph Graph, iterations int) []Point {
	return forceDirectedLayoutParallel(graph, iterations, 800., 600., 1000)
}

func forceDirectedQuadtreeStd(graph Graph, iterations int) []Point {
	return forceDirectedQuadtree(graph, iterations, 800., 600., 1000)
}

func SugiyamaMain() {
	fmt.Printf("Not implemented yet.\n")
}

func main() {
	phaseStart := time.Now()
	layoutFunc := forceDirectedStd
	directed := false

	var (
		png        bool
		iterations int
		algoType   string
		filename   string
	)

	rootCmd := &cobra.Command{
		Use:   "ppa-final",
		Short: "Graph layout visualization tool",
		Run: func(cmd *cobra.Command, args []string) {
			// Validate algorithm type
			validAlgos := map[string]bool{"seq": true, "parallel": true, "sugiyama": true, "quadtree": true}
			if !validAlgos[algoType] {
				cobra.CheckErr(fmt.Errorf("invalid algorithm type '%s'. Valid options: seq, parallel, sugiyama", algoType))
			}

			// Map algorithm type to layout function
			switch algoType {
			case "seq":
				layoutFunc = forceDirectedStd
			case "parallel":
				layoutFunc = forceDirectedParallelStd
			case "sugiyama":
				layoutFunc = SugiyamaLayout
				directed = true
			case "quadtree":
				layoutFunc = forceDirectedQuadtreeStd
			}
		},
	}

	// Boolean flag (default: false)
	rootCmd.Flags().BoolVarP(&png, "png", "p", false, "Enable PNG output")

	// Required integer flag
	rootCmd.Flags().IntVarP(&iterations, "iter", "i", 100, "Number of iterations (required)")
	rootCmd.MarkFlagRequired("iterations")

	// Enumerated string flag
	rootCmd.Flags().StringVarP(&algoType, "algo", "a", "",
		"Algorithm type (seq|parallel|sugiyama|quadtree) (required)")
	rootCmd.MarkFlagRequired("algo")

	// Enumerated string flag
	rootCmd.Flags().StringVarP(&filename, "file", "f", "",
		"Filename (required)")
	rootCmd.MarkFlagRequired("file")

	cobra.CheckErr(rootCmd.Execute())

	graph, err := buildGraphFromFile(filename, directed)
	if err != nil {
		errexit(fmt.Sprintf("Error building graph: %v\n", err))
	}
	endPhase("Build graph", &phaseStart)

	positions := layoutFunc(graph, iterations)
	endPhase("Compute layout", &phaseStart)

	outGraph := augmentGraph(graph, positions)

	if !png {
		RenderGUI(outGraph, directed)
	} else {
		RenderPNG(outGraph, directed)
		endPhase("Create PNG", &phaseStart)
	}
}
