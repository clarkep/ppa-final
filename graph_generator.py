import random
import sys

def generate_edges(N, num_edges):
    edges = []
    for _ in range(num_edges):
        u = random.randint(1, N)
        v = random.randint(1, N)
        while u == v:  # Avoid self-loops
            v = random.randint(1, N)
        edges.append((u, v))
    return edges

if len(sys.argv) != 3:
    print("Usage: python graph_generator.py <N> <num_edges>")
    sys.exit(1)

# Parameters
N = int(sys.argv[1])  # Number of vertices
num_edges = int(sys.argv[2])  # Number of edges

# Generate edges
edges = generate_edges(N, num_edges)

# Write edges to file
with open("input.txt", "w") as f:
    for u, v in edges:
        f.write(f"{u} {v}\n")

print("Edges written to input.txt successfully.")
