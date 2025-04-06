package main

import (
    "bufio"
    "fmt"
    "os"
    "sort"
    "strconv"
    "strings"
)

// Graph structure using adjacency list
type Graph struct {
    adjList map[int][]int
}

// Constructor for new graph
func NewGraph() *Graph {
    return &Graph{
        adjList: make(map[int][]int),
    }
}

// Builds graph from file input
func buildGraphFromFile(filename string) (*Graph, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    graph := NewGraph()

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
        graph.adjList[u] = append(graph.adjList[u], v)
        graph.adjList[v] = append(graph.adjList[v], u)
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return graph, nil
}

// Prints adjacency list with sorted nodes
func (g *Graph) Print() {
    nodes := make([]int, 0, len(g.adjList))
    for node := range g.adjList {
        nodes = append(nodes, node)
    }
    sort.Ints(nodes)

    for _, node := range nodes {
        fmt.Printf("%d -> %v\n", node, g.adjList[node])
    }
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run graph.go <input-file>")
        return
    }

    filename := os.Args[1]
    graph, err := buildGraphFromFile(filename)
    if err != nil {
        fmt.Printf("Error building graph: %v\n", err)
        return
    }

    fmt.Println("Graph adjacency list:")
    graph.Print()
}
