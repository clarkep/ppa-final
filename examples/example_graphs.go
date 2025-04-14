package other

import (
	"fmt"
	"math/rand"
)

func main () {
	r := rand.New(rand.NewSource(99))
	n := 100
	for i := range n {
		for j := i; j < n; j++ {
			if i != j && r.Intn(2) == 1 { 
				fmt.Printf("%d %d\n", i, j)
			}
		}
	}
}