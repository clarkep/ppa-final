# Parallel Graph Drawing
This project explores parallelizing graph drawing algorithms. In particular, it implements parallel force-directed and Sugiyama drawing algorithms.

## Generating Random Graphs
Run `python graph_generator.py vertex_count edge_count`. This will output an `input.txt` where each line consists of an edge between two vertices.

The edges generated will be purely random.

## Running the Layout Generator
1. Run `go get github.com/clarkep/ppa-final`
2. Run `go build .`
3. Run `./ppa-final input.txt` after generating random points. This will output a layout of the graph provided in input.txt after computing a layout algorithm.

## Future Work
TODO
