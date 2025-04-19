package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

// Graph type using adjacency list
type Graph [][]int

// Converts graph from map to slice
func convertGraph(graph map[int][]int) Graph {
	keysToIndices := make(map[int]int, len(graph))
	i := 0
	for k, _ := range graph {
		keysToIndices[k] = i
		i += 1
	}
	out := make([][]int, len(graph))
	for k, v := range graph {
		ki := keysToIndices[k]
		for _, neighbor_key := range v {
			out[ki] = append(out[ki], keysToIndices[neighbor_key])
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