package main

import (
	"math"
	"math/rand"
	"sync"

	"github.com/schollz/progressbar/v3"
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

func assignRandomPositions(nodes Graph, width, height float64) []Point {
	n := len(nodes)
	if n == 0 {
		return make([]Point, 0)
	}

	// Initialize positions randomly
	positions := make([]Point, n)
	for i, _ := range nodes {
		positions[i] = Point{
			X: rand.Float64() * width,
			Y: rand.Float64() * height,
		}
	}

	return positions
}

func forceDirectedLayout(nodes Graph, iterations int, width, height float64) []Point {
	// rand.Seed(time.Now().UnixNano())
	n := len(nodes)
	positions := assignRandomPositions(nodes, width, height)

	k := math.Sqrt((width * height) / float64(n))
	t := width / 10.0
	coolingRate := t / float64(iterations)
	epsilon := 1e-6

	bar := progressbar.Default(int64(iterations))
	for iter := 0; iter < iterations; iter++ {
		bar.Add(1)
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
			for _, v := range u {
				if i < v { // Process each edge once
					delta := positions[v].Sub(positions[i])
					distance := delta.Norm()
					if distance < epsilon {
						distance = epsilon
					}
					force := delta.Scale(distance / k)
					displacements[i] = displacements[i].Add(force)
					displacements[v] = displacements[v].Sub(force)
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

func forceDirectedLayoutParallel(nodes Graph, iterations int, width, height float64, CHUNK_SIZE int) []Point {
	// rand.Seed(time.Now().UnixNano())
	n := len(nodes)
	positions := assignRandomPositions(nodes, width, height)

	k := math.Sqrt((width * height) / float64(n))
	t := width / 10.0
	coolingRate := t / float64(iterations)
	epsilon := 1e-6

	bar := progressbar.Default(int64(iterations))
	for iter := 0; iter < iterations; iter++ {
		bar.Add(1)

		displacements := make([]Point, n)

		// Calculate repulsive forces using goroutines
		for i := 0; i < n; i++ {
			var wg sync.WaitGroup

			// Calculate the number of goroutines needed - break into chunks of CHUNK_SIZE
			goRoutineCount := (n - i - 1 + CHUNK_SIZE - 1) / CHUNK_SIZE

			// Create a slice to collect forces calculated by goroutines
			forcesForI := make([]Point, goRoutineCount)

			for j := 0; j < goRoutineCount; j++ {
				wg.Add(1)

				// Capture loop variables to avoid closure issues
				jCopy := j
				iCopy := i

				go func() {
					defer wg.Done()

					// calculate start and end index for the goroutine from iCopy + 1
					startIndex := iCopy + 1 + jCopy*CHUNK_SIZE
					endIndex := min(startIndex+CHUNK_SIZE, n)

					// Store variable for summing total force to add to i'th node
					totalForce := Point{0, 0}

					for idx := startIndex; idx < endIndex; idx++ {

						// Calculate force between nodes i and idx
						delta := positions[iCopy].Sub(positions[idx])
						distance := delta.Norm()
						if distance < epsilon {
							distance = epsilon
						}
						force := delta.Scale((k * k) / (distance * distance))

						// Directly update displacement for node idx (subtract force)
						displacements[idx] = displacements[idx].Sub(force)

						// Increment force to totalForce
						totalForce = totalForce.Add(force)
					}

					// Store force for j'th goRoutine in the collection array
					forcesForI[jCopy] = totalForce
				}()
			}

			wg.Wait()

			// Sum up all forces and add to displacement for node i
			for _, force := range forcesForI {
				displacements[i] = displacements[i].Add(force)
			}
		}

		// Calculate attractive forces using adjacency list with goroutines
		for i, u := range nodes {
			var wg sync.WaitGroup

			// Calculate the number of goroutines needed - break into chunks of CHUNK_SIZE
			adjacentCount := len(u)
			goRoutineCount := (adjacentCount + CHUNK_SIZE - 1) / CHUNK_SIZE

			// Create a slice to collect forces calculated by goroutines
			forcesForI := make([]Point, goRoutineCount)

			for j := 0; j < goRoutineCount; j++ {

				wg.Add(1)

				jCopy := j
				iCopy := i

				go func() {
					defer wg.Done()

					// calculate start and end index for the goroutine from 0 to adjacentCount
					startIndex := jCopy * CHUNK_SIZE
					endIndex := min(startIndex+CHUNK_SIZE, adjacentCount)

					// Store variable for summing total force to add to i'th node
					totalForce := Point{0, 0}

					for idx := startIndex; idx < endIndex; idx++ {
						v := u[idx]

						// Make sure every edge is calculated once
						if iCopy < v {
							// Calculate force between nodes i and v
							delta := positions[v].Sub(positions[iCopy])
							distance := delta.Norm()
							if distance < epsilon {
								distance = epsilon
							}

							force := delta.Scale(distance / k)

							// Directly update displacement for node v (subtract force)
							displacements[v] = displacements[v].Sub(force)

							// Increment force to totalForce
							totalForce = totalForce.Add(force)
						}
					}

					// Store force for j'th goRoutine in the collection array
					forcesForI[jCopy] = totalForce
				}()
			}

			wg.Wait()

			// Sum up all forces and add to displacement for node i
			for _, force := range forcesForI {
				displacements[i] = displacements[i].Add(force)
			}
		}

		// Update positions with temperature cooling
		var wg sync.WaitGroup

		// Calculate the number of goroutines needed - break into chunks of CHUNK_SIZE
		goRoutineCount := (n + CHUNK_SIZE - 1) / CHUNK_SIZE

		for j := 0; j < goRoutineCount; j++ {
			wg.Add(1)
			jCopy := j

			go func() {
				defer wg.Done()

				// Calculate start and end index for the goroutine from 0 to n
				startIndex := jCopy * CHUNK_SIZE
				endIndex := min(startIndex+CHUNK_SIZE, n)

				for i := startIndex; i < endIndex; i++ {
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
			}()

		}
		wg.Wait()

		t -= coolingRate
	}

	return positions
}

func computeRepulsiveForceBarnesHut(p *Point, node *Quadtree, k, theta, epsilon float64) Point {
	if node == nil || (node.Count == 1 && node.Points[0] == p) {
		return Point{0, 0}
	}

	// Compute distance to grid center
	dx := p.X - node.MidPoint[0]
	dy := p.Y - node.MidPoint[1]
	distance := math.Sqrt(dx*dx + dy*dy)

	// Grid width
	s := node.TopRightCorner[0] - node.BottomLeftCorner[0]

	if s/distance < theta || node.Count == 1 {
		if distance < epsilon {
			distance = epsilon
		}
		forceMag := (k * k * float64(node.Count)) / (distance * distance)
		return Point{X: dx, Y: dy}.Scale(forceMag / distance)
	}

	// Otherwise recurse into children
	totalForce := Point{0, 0}
	children := []*Quadtree{node.BottomLeft, node.BottomRight, node.TopLeft, node.TopRight}
	for _, child := range children {
		f := computeRepulsiveForceBarnesHut(p, child, k, theta, epsilon)
		totalForce = totalForce.Add(f)
	}
	return totalForce
}

func forceDirectedQuadtree(nodes Graph, iterations int, width, height float64, CHUNK_SIZE int) []Point {
	n := len(nodes)
	positions := assignRandomPositions(nodes, width, height)

	k := math.Sqrt((width * height) / float64(n))
	t := width / 10.0
	coolingRate := t / float64(iterations)
	epsilon := 1e-6
	theta := 0.5

	points := make([]*Point, n)
	for i, node := range positions {
		points[i] = &Point{X: node.X, Y: node.Y}
	}

	bar := progressbar.Default(int64(iterations))
	for iter := 0; iter < iterations; iter++ {
		bar.Add(1)

		root := constructQuadtreeLayer(points, [2]float64{0, 0}, [2]float64{width, height}, nil, 0)

		displacements := make([]Point, n)

		// Wait group for goroutines
		var wg sync.WaitGroup

		// Calculate the number of goroutines needed - break into chunks of CHUNK_SIZE
		goRoutineCount := (n + CHUNK_SIZE - 1) / CHUNK_SIZE

		// Repulsive forces (Barnes-Hut)
		for i := 0; i < goRoutineCount; i++ {
			wg.Add(1)
			startIndex := i * CHUNK_SIZE
			endIndex := min(startIndex+CHUNK_SIZE, n)
			go func() {
				defer wg.Done()
				for j := startIndex; j < endIndex; j++ {
					displacements[j] = computeRepulsiveForceBarnesHut(points[j], root, k, theta, epsilon)
				}
			}()
		}

		wg.Wait()

		// Calculate attractive forces using adjacency list with goroutines
		for i, u := range nodes {
			var wg sync.WaitGroup

			// Calculate the number of goroutines needed - break into chunks of CHUNK_SIZE
			adjacentCount := len(u)
			goRoutineCount := (adjacentCount + CHUNK_SIZE - 1) / CHUNK_SIZE

			// Create a slice to collect forces calculated by goroutines
			forcesForI := make([]Point, goRoutineCount)

			for j := 0; j < goRoutineCount; j++ {

				wg.Add(1)

				jCopy := j
				iCopy := i

				go func() {
					defer wg.Done()

					// calculate start and end index for the goroutine from 0 to adjacentCount
					startIndex := jCopy * CHUNK_SIZE
					endIndex := min(startIndex+CHUNK_SIZE, adjacentCount)

					// Store variable for summing total force to add to i'th node
					totalForce := Point{0, 0}

					for idx := startIndex; idx < endIndex; idx++ {
						v := u[idx]

						// Make sure every edge is calculated once
						if iCopy < v {
							// Calculate force between nodes i and v
							delta := positions[v].Sub(positions[iCopy])
							distance := delta.Norm()
							if distance < epsilon {
								distance = epsilon
							}

							force := delta.Scale(distance / k)

							// Directly update displacement for node v (subtract force)
							displacements[v] = displacements[v].Sub(force)

							// Increment force to totalForce
							totalForce = totalForce.Add(force)
						}
					}

					// Store force for j'th goRoutine in the collection array
					forcesForI[jCopy] = totalForce
				}()
			}

			wg.Wait()

			// Sum up all forces and add to displacement for node i
			for _, force := range forcesForI {
				displacements[i] = displacements[i].Add(force)
			}
		}

		// Update positions with temperature cooling
		for j := 0; j < goRoutineCount; j++ {
			wg.Add(1)
			jCopy := j

			go func() {
				defer wg.Done()

				// Calculate start and end index for the goroutine from 0 to n
				startIndex := jCopy * CHUNK_SIZE
				endIndex := min(startIndex+CHUNK_SIZE, n)

				for i := startIndex; i < endIndex; i++ {
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
			}()

		}
		wg.Wait()

		t -= coolingRate
	}

	return positions
}
