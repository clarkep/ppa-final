package main

import (
	"sort"
	"sync"
	"time"
	//	   "fmt"
	//		"math/rand"
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

type BucketQueue struct {
	Nbuckets int
	// convenience for indexing with a shifted index
	off int
	// bucket[node index] = bucket the node is in, or -1 for none
	bucket []int
	// head[bucket index] = node at the head of the bucket, or -1 for none
	head []int
	// next[node index] = next node, or -1 for none
	next []int
	// prev[node index] = previous node, or -1 for none
	prev []int
}

func newBucketQueue(n, m, off int) BucketQueue {
	q := BucketQueue{
		m,
		off,
		make([]int, n),
		make([]int, m),
		make([]int, n),
		make([]int, n),
	}
	for i := range n {
		q.bucket[i] = -1
		q.next[i] = -1
		q.prev[i] = -1
	}
	for i := range m {
		q.head[i] = -1
	}
	return q
}

// Push node n to bucket i
func (q *BucketQueue) push(n, b int) {
	i := b + q.off
	q.bucket[n] = i
	q.prev[n] = -1
	q.next[n] = q.head[i]
	if q.head[i] != -1 {
		q.prev[q.head[i]] = n
	}
	q.head[i] = n
}

// Remove node n from bucket i
func (q *BucketQueue) pop(n int) {
	i := q.bucket[n]
	if i >= 0 {
		q.bucket[n] = -1
		if q.prev[n] == -1 {
			q.head[i] = q.next[n]
		} else {
			nextPrev := q.prev[n]
			q.next[nextPrev] = q.next[n]
		}
		if q.next[n] != -1 {
			q.prev[q.next[n]] = q.prev[n]
		}
		q.next[n], q.prev[n] = -1, -1
	}
}

func removeCycles(outgoing Graph) (Graph, [][2]int) {
	// Greedy Cycle Removal
	n := len(outgoing)
	incoming := get_incoming_edges(outgoing)

	indeg, outdeg := make([]int, n), make([]int, n)
	delta := make([]int, n) // out‑deg – in‑deg
	alive := make([]bool, n)

	Nbuckets := 2*(n-1) + 1 // [-(n-1), (n-1)]
	off := n - 1
	q := newBucketQueue(n, Nbuckets, off)

	sinks, sources := []int{}, []int{}
	for v := range n {
		alive[v] = true
		indeg[v] = len(incoming[v])
		outdeg[v] = len(outgoing[v])
		delta[v] = outdeg[v] - indeg[v]
		q.push(v, delta[v])
		if outdeg[v] == 0 {
			sinks = append(sinks, v)
		}
		if indeg[v] == 0 {
			sources = append(sources, v)
		}
	}

	// Track current maximum delta bucket
	curMax := Nbuckets - 1
	s1, s2 := []int{}, []int{} // final ordering pieces
	remaining := n

	// Update a node with a new delta
	update := func(u, deltaIncr int) {
		old := delta[u]
		delta[u] += deltaIncr
		if delta[u] != old {
			q.pop(u)
			q.push(u, delta[u])
		}
	}

	// Remove a node from the graph and update buckets, O(1+deg[v]).
	remove := func(v int) {
		alive[v] = false
		remaining--
		q.pop(v)
		for _, u := range incoming[v] {
			if alive[u] {
				outdeg[u]--
				if outdeg[u] == 0 {
					q.pop(u)
					sinks = append(sinks, u)
				}
				update(u, -1)
			}
		}
		for _, u := range outgoing[v] {
			if alive[u] {
				indeg[u]--
				if indeg[u] == 0 {
					q.pop(u)
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
			// choose vertex with current maximum delta
			for curMax >= 0 && q.head[curMax] == -1 {
				curMax--
			}
			v := q.head[curMax]
			remove(v)
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
	for i, v := range order {
		pos[v] = i
	}
	res := make([][]int, n)
	fas := make([][2]int, n)
	for u := range n {
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

func assignLevels(graph Graph) ([][]int, [][2]int) {
	// the "longest path algorithm"
	out := make([][]int, 1)
	n := len(graph)
	levels := make([][2]int, n)
	U := make([]bool, n)
	Z := make([]bool, n)
	currentLayer := 0
	for !allTrue(U) {
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
					levels[i] = [2]int{currentLayer, len(out[currentLayer])}
					out[currentLayer] = append(out[currentLayer], i)
					U[i] = true
					// goto used as a "continue", but for the outer loop
					continue
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
	return out[:len(out)-1], levels
}

func assignLevelsPar(graph Graph, nWorkers int) ([][]int, [][2]int) {
	// the "longest path algorithm", with parallelized search
	// over nodes
	out := make([][]int, 1)
	n := len(graph)
	levels := make([][2]int, n)
	chunkN := (n + nWorkers - 1) / nWorkers // xxx
	U := make([]bool, n)
	Z := make([]bool, n)
	currentLayer := 0
	for !allTrue(U) {
		var wg sync.WaitGroup
		result := make(chan int, n)
		for wi := range nWorkers {
			start := wi * chunkN
			end := min(start+chunkN, n)
			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				for i := start; i < end; i++ {
					// select all vertices from V \ U with all outgoing edges in Z
					if !U[i] {
						selected := true
						for _, v := range graph[i] {
							if !Z[v] {
								selected = false
								break
							}
						}
						if selected {
							result <- i
						}
					}
				}
			}(start, end)
		}
		wg.Wait()
		close(result)

		for i := range result {
			levels[i] = [2]int{currentLayer, len(out[currentLayer])}
			out[currentLayer] = append(out[currentLayer], i)
			U[i] = true
		}
		currentLayer++
		out = append(out, make([]int, 0))
		// Z = Z union U
		for i := range n {
			if U[i] {
				Z[i] = true
			}
		}
	}
	return out[:len(out)-1], levels
}

type sortPair struct {
	P float64
	U int
}

func orderLevels(graph Graph, levels [][]int, levelmap [][2]int) [][]int {
	incoming := get_incoming_edges(graph)
	nLevels := len(levels)
	nSources := make([]int, nLevels)
	nSources[nLevels-1] = len(levels[nLevels-1])
	// Barycentric averaging
	for i := nLevels - 2; i >= 0; i-- {
		lvl := levels[i]
		order := make([]sortPair, len(levels[i]))
		for j, v := range lvl {
			pos := 0.0
			if len(incoming[v]) == 0 {
				nSources[i] += 1
				pos = -1
			} else {
				for _, u := range incoming[v] {
					lvl, lvl_i := levelmap[u][0], levelmap[u][1]
					n := len(levels[lvl])
					pos += float64(n-lvl_i) / float64(n+1)
				}
			}
			order[j] = sortPair{pos, lvl[j]}
		}
		sort.Slice(order, func(i, j int) bool {
			return order[i].P < order[j].P
		})
		for j, o := range order {
			levels[i][j] = o.U
			levelmap[o.U][1] = j
		}
	}
	// re-insert sources at their ideal locations. Slow, O(|nodes in level||edges in level|^2),
    // because it explicitly counts the number of crossings that will result
	for i := nLevels - 1; i >= 1; i-- {
		if len(levels[i]) == 1 {
			continue
		}
		levelSources := make([]int, nSources[i])
		copy(levelSources, levels[i][:nSources[i]])
		startN := nSources[i]
		if startN == len(levels[i]) {
			startN--
		}
		// spot -> number of crossings if the current source is placed there
		crossings := make([]int, len(levels[i]))
		for _, src := range levelSources {
			for k, other := range levels[i][startN:] {
				k = k + startN
				for _, otherEdge := range graph[other] {
					if levelmap[otherEdge][0] != i-1 {
						continue
					}
					otherEdgeI := levelmap[otherEdge][1]
					for _, srcEdge := range graph[src] {
						if levelmap[srcEdge][0] != i-1 {
							continue
						}
						srcEdgeI := levelmap[srcEdge][1]
						if srcEdgeI > otherEdgeI {
							for h := startN; h <= k; h++ {
								crossings[h] += 1
							}
						} else if srcEdgeI < otherEdgeI {
							for h := k + 1; h < len(levels[i]); h++ {
								crossings[h] += 1
							}
						}
					}
				}
			}
			minCrossings := 100000000
			bestSpot := -1
			for k := startN; k < len(crossings); k++ {
				if crossings[k] < minCrossings {
					minCrossings = crossings[k]
					bestSpot = k
				}
			}
			copy(levels[i], levels[i][1:bestSpot+1])
			levels[i][bestSpot] = src
			startN--
		}
	}
	return levels
}

func orderLevelsPar(graph Graph, levels [][]int, levelmap [][2]int) [][]int {
    incoming := get_incoming_edges(graph)
    nLevels := len(levels)
    nSources := make([]int, nLevels)
    nSources[nLevels-1] = len(levels[nLevels-1])
    // Barycentric averaging
    for i := nLevels - 2; i >= 0; i-- {
        lvl := levels[i]
        order := make([]sortPair, len(levels[i]))
        for j, v := range lvl {
            pos := 0.0
            if len(incoming[v]) == 0 {
                nSources[i] += 1
                pos = -1
            } else {
                for _, u := range incoming[v] {
                    lvl, lvl_i := levelmap[u][0], levelmap[u][1]
                    n := len(levels[lvl])
                    pos += float64(n-lvl_i) / float64(n+1)
                }
            }
            order[j] = sortPair{pos, lvl[j]}
        }
        sort.Slice(order, func(i, j int) bool {
            return order[i].P < order[j].P
        })
        for j, o := range order {
            levels[i][j] = o.U
            levelmap[o.U][1] = j
        }
    }

    levels2 := make([][]int, len(levels))
    var wg sync.WaitGroup
    // In parallel, re-insert sources at their ideal locations. Slow, O(|nodes in level||edges in level|^2) work.
    for i := nLevels - 1; i >= 0; i-- {
		if (len(levels[i]) <= 1 || i == 0) {
			levels2[i] = levels[i]
			continue
		}
        level := make([]int, len(levels[i]))
		copy(level, levels[i])
        levelSources := make([]int, nSources[i])
        copy(levelSources, levels[i][:nSources[i]])
        startN := nSources[i]
        if startN == len(levels[i]) {
            startN--
        }
        wg.Add(1)
        go func() {
            defer wg.Done()
            // spot -> number of crossings if the current source is placed there
            crossings := make([]int, len(level))
			if len(levels[i]) == 1 {
				goto done
			}
            for _, src := range levelSources {
                for k, other := range level[startN:] {
                    k = k + startN
                    for _, otherEdge := range graph[other] {
                        if levelmap[otherEdge][0] != i-1 {
                            continue
                        }
                        otherEdgeI := levelmap[otherEdge][1]
                        for _, srcEdge := range graph[src] {
                            if levelmap[srcEdge][0] != i-1 {
                                continue
                            }
                            srcEdgeI := levelmap[srcEdge][1]
                            if srcEdgeI > otherEdgeI {
                                for h := startN; h <= k; h++ {
                                    crossings[h] += 1
                                }
                            } else if srcEdgeI < otherEdgeI {
                                for h := k + 1; h < len(level); h++ {
                                    crossings[h] += 1
                                }
                            }
                        }
                    }
                }
                minCrossings := 100000000
                bestSpot := -1
                for k := startN; k < len(crossings); k++ {
                    if crossings[k] < minCrossings {
                        minCrossings = crossings[k]
                        bestSpot = k
                    }
                }
                copy(level, level[1:bestSpot+1])
                level[bestSpot] = src
                startN--
            }
			done:
            levels2[i] = level
        }()
    }
    wg.Wait()
    return levels2
}

func assignCoordinates(graph Graph, orders [][]int) []Point {
	out := make([]Point, len(graph))
	for x, lvl := range orders {
		n := len(lvl)
		for i, u := range lvl {
			// just assign coordinates based on (level, order in level)
			out[u] = Point{X: float64(len(orders) - x), Y: (100 * float64(n-i)) / float64(n+1)}
		}
	}
	return out
}

const subphases = true

func SugiyamaLayout(graph Graph, iterations int) []Point {
	startTime := time.Now()

	graph2, _ := removeCycles(graph)

	if subphases {
		endPhase("\tRemove cycles", &startTime)
	}

	levels, levelmap := assignLevelsPar(graph2, iterations)

	if subphases {
		endPhase("\tAssign levels", &startTime)
	}

	orders := orderLevelsPar(graph2, levels, levelmap)

	if subphases {
		endPhase("\tOrder levels", &startTime)
	}

	positions := assignCoordinates(graph2, orders)

	if subphases {
		endPhase("\tAssign coordinates", &startTime)
	}

	return positions
}
