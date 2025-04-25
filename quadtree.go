package main

import (
	"math"
	"sync"

	"github.com/google/uuid"
)

const MAX_DEPTH = 6

type Quadtree struct {
	BottomLeft, BottomRight, TopLeft, TopRight *Quadtree
	Points                                     []*Point
	Parent                                     *Quadtree

	// Number of points in the grid
	Count int

	// ID of the grid
	ID string

	// Grid Dimension
	MidPoint         [2]float64
	TopRightCorner   [2]float64
	BottomLeftCorner [2]float64
}

func newGrid(bottomLeft, topRight [2]float64, id string, points []*Point, parent *Quadtree) *Quadtree {
	x1, y1 := bottomLeft[0], bottomLeft[1]
	x2, y2 := topRight[0], topRight[1]
	return &Quadtree{
		Points:           points,
		Parent:           parent,
		Count:            len(points),
		ID:               id,
		TopRightCorner:   topRight,
		BottomLeftCorner: bottomLeft,
		MidPoint:         [2]float64{(x1 + x2) / 2, (y1 + y2) / 2},
		BottomLeft:       nil,
		BottomRight:      nil,
		TopLeft:          nil,
		TopRight:         nil,
	}
}

func constructQuadtreeLayer(points []*Point, bottomLeft, topRight [2]float64, parent *Quadtree, depth int) *Quadtree {
	id := uuid.New().String()
	quadtree := newGrid(bottomLeft, topRight, id, points, parent)

	// Assign grid id as smallest grid of points
	// for _, point := range points {
	// 	pointsToGrid[pointsToPosition[point]] = quadtree
	// }

	x1, y1 := bottomLeft[0], bottomLeft[1]
	x2, y2 := topRight[0], topRight[1]

	// Calculate the midpoint of the bounding box
	midX := (x1 + x2) / 2
	midY := (y1 + y2) / 2

	if quadtree.Count > 1 {
		// Split the points into four quadrants
		var bottomLeftPoints, bottomRightPoints, topLeftPoints, topRightPoints []*Point
		for _, point := range points {
			if point.X <= midX && point.Y <= midY {
				bottomLeftPoints = append(bottomLeftPoints, point)
			} else if point.X > midX && point.Y <= midY {
				bottomRightPoints = append(bottomRightPoints, point)
			} else if point.X <= midX && point.Y > midY {
				topLeftPoints = append(topLeftPoints, point)
			} else {
				topRightPoints = append(topRightPoints, point)
			}
		}

		var wg sync.WaitGroup
		useGoRoutines := false
		if depth < MAX_DEPTH {
			useGoRoutines = true
		}

		if len(bottomLeftPoints) > 0 {
			if useGoRoutines {
				wg.Add(1)
				go func() {
					defer wg.Done()
					quadtree.BottomLeft = constructQuadtreeLayer(bottomLeftPoints, bottomLeft, [2]float64{midX, midY}, quadtree, depth+1)
				}()
			} else {
				quadtree.BottomLeft = constructQuadtreeLayer(bottomLeftPoints, bottomLeft, [2]float64{midX, midY}, quadtree, depth+1)
			}
		}

		if len(bottomRightPoints) > 0 {
			if useGoRoutines {
				wg.Add(1)
				go func() {
					defer wg.Done()
					quadtree.BottomRight = constructQuadtreeLayer(bottomRightPoints, [2]float64{midX, y1}, [2]float64{x2, midY}, quadtree, depth+1)
				}()
			} else {
				quadtree.BottomRight = constructQuadtreeLayer(bottomRightPoints, [2]float64{midX, y1}, [2]float64{x2, midY}, quadtree, depth+1)
			}
		}

		if len(topLeftPoints) > 0 {
			if useGoRoutines {
				wg.Add(1)
				go func() {
					defer wg.Done()
					quadtree.TopLeft = constructQuadtreeLayer(topLeftPoints, [2]float64{x1, midY}, [2]float64{midX, y2}, quadtree, depth+1)
				}()
			} else {
				quadtree.TopLeft = constructQuadtreeLayer(topLeftPoints, [2]float64{x1, midY}, [2]float64{midX, y2}, quadtree, depth+1)
			}
		}

		if len(topRightPoints) > 0 {
			if useGoRoutines {
				wg.Add(1)
				go func() {
					defer wg.Done()
					quadtree.TopRight = constructQuadtreeLayer(topRightPoints, [2]float64{midX, midY}, topRight, quadtree, depth+1)
				}()
			} else {
				quadtree.TopRight = constructQuadtreeLayer(topRightPoints, [2]float64{midX, midY}, topRight, quadtree, depth+1)
			}
		}

		// Wait for all goroutines to finish
		if useGoRoutines {
			wg.Wait()
		}
	}

	return quadtree
}

func (p *Point) getCommonAncestor(q *Point, pointToGrid map[*Point]*Quadtree) *Quadtree {
	len1 := 0
	for p_p := pointToGrid[p]; p_p != nil; p_p = p_p.Parent {
		len1++
	}

	len2 := 0
	for p_q := pointToGrid[q]; p_q != nil; p_q = p_q.Parent {
		len2++
	}

	larger := p
	smaller := q
	if len2 > len1 {
		larger = q
		smaller = p
	}
	diff := int(math.Abs(float64(len1 - len2)))

	r := pointToGrid[larger]
	for c := 0; c < diff; c++ {
		r = r.Parent
	}

	p_1 := r
	p_2 := pointToGrid[smaller]
	for p_1 != nil && p_2 != nil {
		if p_1.ID == p_2.ID {
			return p_1
		}
		p_1 = p_1.Parent
		p_2 = p_2.Parent
	}

	return nil
}
