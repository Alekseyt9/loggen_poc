package main

import "fmt"

func main() {
	cfg := defaultConfig()
	g := Generate(cfg)
	edges := 0

	for _, outs := range g.Out {
		edges += len(outs)
	}

	// Output sizes of Out and In matrices
	outRows := len(g.Out)
	inRows := len(g.In)
	outMax := 0
	inMax := 0
	for _, row := range g.Out {
		l := len(row)
		if l > outMax {
			outMax = l
		}
	}
	for _, row := range g.In {
		l := len(row)
		if l > inMax {
			inMax = l
		}
	}

	fmt.Printf("Out size: rows=%d, maxRow=%d\n", outRows, outMax)
	fmt.Printf("In size:  rows=%d, maxRow=%d\n", inRows, inMax)
	fmt.Printf("generated graph with %d nodes, %d edges, %d pairs\n", len(g.Nodes), edges, len(g.Pairs))
	fmt.Printf("spike edges: %d\n", g.SpikeEdges)

	total := GraphMemBreakdown(g)
	toMB := func(b int64) float64 { return float64(b) / (1024 * 1024) }
	fmt.Printf("mem TOTAL:      %d bytes (%.2f MB)\n", total, toMB(total))
}
