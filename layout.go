package main

import (
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

func ForceDirectedLayout(nodes Graph, iterations int, width, height float64) []Point {
    // rand.Seed(time.Now().UnixNano())
    n := len(nodes)
    if n == 0 {
        return make([]Point, 0)
    }
    
    // Initialize positions randomly
    positions := make([]Point, n)
    for i,_ := range nodes {
        positions[i] = Point{
            X: rand.Float64() * width,
            Y: rand.Float64() * height,
        }
    }
    
    k := math.Sqrt((width * height) / float64(n))
    t := width / 10.0
    coolingRate := t / float64(iterations)
    epsilon := 1e-6

    for iter := 0; iter < iterations; iter++ {
        displacements := make([]Point, n)
        
        // Calculate repulsive forces
        for i := 0; i < len(nodes); i++ {
            for j := i + 1; j < len(nodes); j++ {
                delta := positions[i].Sub(positions[j])
                distance := delta.Norm()
                if distance < epsilon {
                    distance = epsilon
                }
                force := delta.Scale((k * k) / (distance * distance))
                displacements[i] = displacements[i].Add(force)
                displacements[j] = displacements[j].Sub(force)
            }
        }
        
        // Calculate attractive forces using adjacency list
        for i, u := range nodes {
            for j, _ := range u {
                if i < j { // Process each edge once
                    delta := positions[j].Sub(positions[i])
                    distance := delta.Norm()
                    if distance < epsilon {
                        distance = epsilon
                    }
                    force := delta.Scale(distance / k)
                    displacements[i] = displacements[i].Add(force)
                    displacements[j] = displacements[j].Sub(force)
                }
            }
        }
        
        // Update positions with temperature cooling
        for i, _ := range nodes {
            disp := displacements[i]
            dispNorm := disp.Norm()
            if dispNorm > 0 {
                disp = disp.Scale(math.Min(dispNorm, t) / dispNorm)
                newPos := positions[i].Add(disp)
                newPos.X = clamp(newPos.X, 0, width)
                newPos.Y = clamp(newPos.Y, 0, height)
                positions[i] = newPos
            }
        }
        
        t -= coolingRate
    }
    
    return positions
}
