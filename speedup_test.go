/* Run these tests with:
	go test -v .
*/
package main

import (
	"testing"
	"fmt"
	"time"
)

func TestSugiyamaSpeedup1 (t *testing.T) {
	files := [...]string{
		"examples/dag8.txt",
		"examples/dag40.txt",
		"examples/dag100.txt",
		"examples/dag1000.txt",
		"examples/dag10k.txt",
		"examples/dag100k.txt",
	}

	procs := [...]int { 1, 2, 4, 8, 12 }

	fmt.Printf("filename              1         2         4         8          12\n")
	for _, fn := range files {
		fmt.Printf("%-21s", fn)
		graph, err := buildGraphFromFile(fn, true)
		for _, p := range procs {
			if err != nil {
				errexit(fmt.Sprintf("Error building graph: %v\n", err))
			}
			const trials = 10
			var time_sum int64 = 0
			for range trials {
				start := time.Now()
				_ = assignLevelsPar(graph, p)
				end := time.Now()
				time_sum += end.Sub(start).Nanoseconds()
			}
			avg_time := time_sum / trials

			fmt.Printf("%10s", scaledTime(avg_time))
		}
		fmt.Printf("\n")
	}
}