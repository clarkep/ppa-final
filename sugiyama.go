package main

import (
	"slices"
	"sort"
//	"math/rand"
)

// Augment a graph with incoming edges. O(n + m)
func get_incoming_edges(graph Graph) [][]int {
	res := make([][]int, len(graph))
	for i, u := range graph {
		for _, v := range u {
			res[v] = append(res[v], i)
		}
	}
	return res
}

// Push value v to bucket i
// func (b: *BucketQueue) pushBucket (i, v int)
// Remove value v from bucket i
// func (b: *BucketQueue) popBucket (i, v int)

func removeCycles(outgoing Graph) (Graph, [][2]int) {
	// Greedy Cycle Removal
    n := len(outgoing)
	incoming := get_incoming_edges(outgoing)

    indeg, outdeg := make([]int, n), make([]int, n)
    delta := make([]int, n) // out‑deg – in‑deg
    alive := make([]bool, n)

    Nbuckets := 2*(n-1) + 1       // [-(n-1), (n-1)]
    off := n - 1                  
    head := make([]int, Nbuckets) // head[delta + off] = first node in bucket delta
    next := make([]int, n)        // forward links inside buckets
    prev := make([]int, n)        // backward links
    for i := range head { head[i] = -1 }

    pushBucket := func(v int) {
	    i := delta[v] + off
        prev[v] = -1
        next[v] = head[i]
        if head[i] != -1 { prev[head[i]] = v }
        head[i] = v
    }
    popFromBucket := func(v int) {
        i := delta[v] + off
        if prev[v] == -1 {
            head[i] = next[v]
        } else {
            nextPrev := prev[v]
            next[nextPrev] = next[v]
        }
        if next[v] != -1 { prev[next[v]] = prev[v] }
        next[v], prev[v] = -1, -1
    }

    sinks, sources := []int{}, []int{}
	for v := range n { 
        alive[v] = true
        indeg[v] = len(incoming[v])
        outdeg[v] = len(outgoing[v])
        delta[v] = outdeg[v] - indeg[v]
		pushBucket(v) 
        if outdeg[v] == 0 { sinks = append(sinks, v) }
        if indeg[v] == 0 { sources = append(sources, v) }
	}

    // Track current maximum delta bucket
    curMax := Nbuckets - 1
    advanceMax := func() {
        for curMax >= 0 && head[curMax] == -1 { curMax-- }
    }
    advanceMax()

    s1, s2 := []int{}, []int{} // final ordering pieces
    remaining := n

    // Update a node with a new delta
    update := func(u, deltaIncr int) {
        old := delta[u]
        delta[u] += deltaIncr
	    if delta[u] != old { 
            popFromBucket(u); 
            pushBucket(u);
        }
    }

    // Remove a node from the graph and update buckets, O(1+deg[v]).
    remove := func(v int) {
        alive[v] = false
        remaining--
        popFromBucket(v)
        for _, u := range incoming[v] {
            if alive[u] { 
                outdeg[u]--
                if outdeg[u] == 0 {
                    popFromBucket(u)
                    sinks = append(sinks, u)
                }
                update(u, -1)
            }
        }
        for _, u := range outgoing[v] {
            if alive[u] { 
                indeg[u]--
                if indeg[u] == 0 {
                    popFromBucket(u)
                    sources = append(sources, u)
                }
                update(u, +1)
            }
        }
    }

    // Main loop
    for remaining > 0 {
        switch {
        case len(sinks) > 0:
            v := sinks[len(sinks)-1]
            sinks = sinks[:len(sinks)-1]
            if alive[v] { 
            	remove(v)
                s2 = append(s2, v)
            }
        case len(sources) > 0:
            v := sources[len(sources)-1]
            sources = sources[:len(sources)-1]
            if alive[v] {
            	remove(v)
                s1 = append(s1, v)
            }
        default:
            // choose vertex with current maximum Δ
            advanceMax()
            v := head[curMax]
            s1 = append(s1, v)
        }
    }

    // Create ordering
    order := s1
    for i := len(s2) - 1; i >= 0; i-- { 
    	order = append(order, s2[i])
    }

    // Extract feedback‑arc set and modified graph
    pos := make([]int, n)
    for i, v := range order { pos[v] = i }
    res := make([][]int, n)
    fas := make([][2]int, n)
    for u := 0; u < n; u++ {
        for _, v := range outgoing[u] {
            if pos[u] > pos[v] { 
            	res[v] = append(res[v], u)
            	fas = append(fas, [2]int{u, v})
            } else {
            	res[u] = append(res[u], v)
            }
        }
    }
    return res, fas
}

func allTrue(vertexSet []bool) bool {
	for i := range vertexSet {
		if !vertexSet[i] {
			return false
		}
	}
	return true
}

func assignLevels(graph Graph) [][]int {
	// the "longest path algorithm"
	out := make([][]int, 1)
	n := len(graph)
	U := make([]bool, n)
	Z := make([]bool, n)
	currentLayer := 0
	for !allTrue(U) {
	outerLoop:
		for i := range n {
			// select a vertex from V \ U with all outgoing edges in Z
			if !U[i] {
				selected := true
				for _, v := range graph[i] {
					if !Z[v] {
						selected = false
						break
					}
				}
				if selected {
					out[currentLayer] = append(out[currentLayer], i)
					U[i] = true
					// goto used as a "continue", but for the outer loop
					goto outerLoop
				}
			}
		}
		// none have been selected, so go up a layer
		currentLayer++
		out = append(out, make([]int, 0))
		// Z = Z union U
		for i := range n {
			if U[i] {
				Z[i] = true
			}
		}
	}
	return out
}

func orderLevels(graph Graph, levels [][]int) [][]int {
	// Barycentric averaging
	for i := 1; i < len(levels); i++ {
		lvl := levels[i]
		order := make([][2]int, len(levels[i]))
		for j, v := range lvl {
			pos := 0
			for k, u := range levels[i - 1] {
				if slices.Contains(graph[u], v) {
					pos += k
				}
			}
			order[j][0] = pos
			order[j][1] = lvl[j]
		}
		sort.Slice(order, func (i, j int) bool {
			return order[i][0] < order[j][0]
		})
		for j := range order {
			levels[i][j] = order[j][1]
		}
	}
	return levels
}

func assignCoordinates(graph Graph, orders [][]int) []Point {
	out := make([]Point, len(graph))
	for x, lvl := range orders {
		n := len(lvl)
		for i, u := range lvl {
			// just assign coordinates based on (level, order in level)
			out[u] = Point{X: float64(len(orders) - x), Y: (100*float64(i+1))/float64(n+1)}
		}
	}
	return out
}

func SugiyamaLayout(graph Graph, iterations int) []Point {
	graph2, _ := removeCycles(graph)

	levels := assignLevels(graph2)

	orders := orderLevels(graph2, levels)

	positions := assignCoordinates(graph2, orders)

	return positions
}