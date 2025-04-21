package main

import (
	"fmt"
	"github.com/spf13/cobra"
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

func forceDirectedStd(graph Graph, iterations int) []Point {
	return ForceDirectedLayout(graph, iterations, 800., 600.)
}

func forceDirectedParallelStd(graph Graph, iterations int) []Point {
	return ForceDirectedLayoutParallel(graph, iterations, 800., 600., 1000)
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
			validAlgos := map[string]bool{"seq": true, "parallel": true, "sugiyama": true}
			if !validAlgos[algoType] {
				cobra.CheckErr(fmt.Errorf("invalid algorithm type '%s'. Valid options: seq, parallel, sugiyama", algoType))
			}

			// Map algorithm type to layout function
			switch algoType {
			case "seq":
				layoutFunc = forceDirectedStd
				directed = false
			case "parallel":
				layoutFunc = forceDirectedParallelStd
				directed = false
			case "sugiyama":
				layoutFunc = SugiyamaLayout
				directed = true
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
		"Algorithm type (seq|parallel|sugiyama) (required)")
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
