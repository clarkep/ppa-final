package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
    "sort"
)

// Graph type using adjacency list
type Graph [][]int

// Converts graph from map to slice
func convertGraph(graph map[int][]int) Graph {
	keys := make([]int, len(graph))
	i := 0
	for k, _ := range graph {
		keys[i] = k
		i += 1
	}
    // Put the keys in sorted order to make debugging easier: if the node names in the file
    // start with 0 and don't have gaps, they will be the same as the node indices in the graph.
    // Can put this behind a debug mode flag if we want.
    sort.Ints(keys)
    keysToIndices := make(map[int]int)
    for i, k := range keys {
        keysToIndices[k] = i
    }
	out := make([][]int, len(graph))
	for i, k := range keys {
		for _, neighbor_key := range graph[k] {
			out[i] = append(out[i], keysToIndices[neighbor_key])
		}
	}
	return out
}

// Builds graph from file input
func buildGraphFromFile(filename string, directed bool) (Graph, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    graph := make(map[int][]int)

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Fields(line)
        if len(parts) != 2 {
            return nil, fmt.Errorf("invalid line format: %s", line)
        }

        u, err := strconv.Atoi(parts[0])
        if err != nil {
            return nil, fmt.Errorf("invalid node %s: %v", parts[0], err)
        }

        v, err := strconv.Atoi(parts[1])
        if err != nil {
            return nil, fmt.Errorf("invalid node %s: %v", parts[1], err)
        }

        // Add edges both ways for undirected graph
        graph[u] = append(graph[u], v)
        if (!directed) {
            graph[v] = append(graph[v], u)
        } else {
            // Need to initialize nodes even if they have no outgoing edges
            _, ok := graph[v]
            if !ok {
                graph[v] = make([]int, 0)
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return convertGraph(graph), nil
}

// Prints adjacency list
func (g Graph) Print() {
    for i, u := range g {
        for j, _ := range u {
            fmt.Printf("%d -> %d\n", i, j)
        }
    }
}