package main

import (
    "bufio"
    "fmt"
    "os"
    "sort"
    "strconv"
    "strings"
    "math"
    "math/rand"
)

type Point struct {
    X, Y float64
}

func (p Point) Add(q Point) Point {
    return Point{X: p.X + q.X, Y: p.Y + q.Y}
}

func (p Point) Sub(q Point) Point {
    return Point{X: p.X - q.X, Y: p.Y - q.Y}
}

func (p Point) Scale(s float64) Point {
    return Point{X: p.X * s, Y: p.Y * s}
}

func (p Point) Norm() float64 {
    return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

func clamp(val, min, max float64) float64 {
    if val < min {
        return min
    }
    if val > max {
        return max
    }
    return val
}

func ComputeForceDirectedLayout(g *Graph, iterations int, width, height float64) map[int]Point {
    // rand.Seed(time.Now().UnixNano())
    nodes := make([]int, 0, len(g.adjList))
    for node := range g.adjList {
        nodes = append(nodes, node)
    }
    
    n := len(nodes)
    if n == 0 {
        return make(map[int]Point)
    }
    
    // Initialize positions randomly
    positions := make(map[int]Point)
    for _, node := range nodes {
        positions[node] = Point{
            X: rand.Float64() * width,
            Y: rand.Float64() * height,
        }
    }
    
    k := math.Sqrt((width * height) / float64(n))
    t := width / 10.0
    coolingRate := t / float64(iterations)
    epsilon := 1e-6

    for iter := 0; iter < iterations; iter++ {
        displacements := make(map[int]Point)
        
        // Calculate repulsive forces
        for i := 0; i < len(nodes); i++ {
            u := nodes[i]
            for j := i + 1; j < len(nodes); j++ {
                v := nodes[j]
                delta := positions[u].Sub(positions[v])
                distance := delta.Norm()
                if distance < epsilon {
                    distance = epsilon
                }
                force := delta.Scale((k * k) / (distance * distance))
                displacements[u] = displacements[u].Add(force)
                displacements[v] = displacements[v].Sub(force)
            }
        }
        
        // Calculate attractive forces using adjacency list
        for _, u := range nodes {
            for _, v := range g.adjList[u] {
                if u < v { // Process each edge once
                    delta := positions[v].Sub(positions[u])
                    distance := delta.Norm()
                    if distance < epsilon {
                        distance = epsilon
                    }
                    force := delta.Scale(distance / k)
                    displacements[u] = displacements[u].Add(force)
                    displacements[v] = displacements[v].Sub(force)
                }
            }
        }
        
        // Update positions with temperature cooling
        for _, node := range nodes {
            disp := displacements[node]
            dispNorm := disp.Norm()
            if dispNorm > 0 {
                disp = disp.Scale(math.Min(dispNorm, t) / dispNorm)
                newPos := positions[node].Add(disp)
                newPos.X = clamp(newPos.X, 0, width)
                newPos.Y = clamp(newPos.Y, 0, height)
                positions[node] = newPos
            }
        }
        
        t -= coolingRate
    }
    
    return positions
}

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