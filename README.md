# Parallel Graph Drawing Algorithms
This repository consists of code for exploring parallel graph drawing algorithms and their benchmarks.

## Generating Random Graph
In order to generate random points, do the following - 
Run `python graph_generator.py vertex_count edge_count` where `vertex_count` is the number of vertices and `edge_count` is the edge count for the graph. This will output an `input.txt` where each line consists of an edge between two vertices.

The edges generated will be purely random.

## Run Graph Layout Generator
Perform the following steps - 
1. Run `go get github.com/clarkep/ppa-final`
2. Run `go build .`
3. Run `./ppa-final input.txt` after generating random points. This will output a layout of the graph provided in input.txt after computing a layout algorithm.

## Future Work
TODO