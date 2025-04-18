package main

import (
	"math"
	"math/rand"
	"sync"
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
	for i, _ := range nodes {
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
		// print iteration count after every 1000 iterations
		if iter%1000 == 0 {
			println("Iteration: ", iter)
		}

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

func ForceDirectedLayoutParallel(nodes Graph, iterations int, width, height float64, CHUNK_SIZE int) []Point {
	// rand.Seed(time.Now().UnixNano())
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

	k := math.Sqrt((width * height) / float64(n))
	t := width / 10.0
	coolingRate := t / float64(iterations)
	epsilon := 1e-6

	for iter := 0; iter < iterations; iter++ {
		// print iteration count after every 1000 iterations
		if iter%1000 == 0 {
			println("Iteration: ", iter)
		}

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
