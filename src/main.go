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
	outTotal := 0
	inTotal := 0
	outMax := 0
	inMax := 0
	for _, row := range g.Out {
		l := len(row)
		outTotal += l
		if l > outMax {
			outMax = l
		}
	}
	for _, row := range g.In {
		l := len(row)
		inTotal += l
		if l > inMax {
			inMax = l
		}
	}

	fmt.Printf("Out size: rows=%d, total=%d, maxRow=%d\n", outRows, outTotal, outMax)
	fmt.Printf("In size:  rows=%d, total=%d, maxRow=%d\n", inRows, inTotal, inMax)
	fmt.Printf("generated graph with %d nodes, %d edges, %d pairs\n", len(g.Nodes), edges, len(g.Pairs))

	// Memory usage breakdown (approx bytes)
	total, bNodes, bOut, bIn, bPairs, bPathsHdr, bPathsElems, bAllow := GraphMemBreakdown(g)
	toMB := func(b int64) float64 { return float64(b) / (1024 * 1024) }
	fmt.Printf("mem Nodes:      %d bytes (%.2f MB)\n", bNodes, toMB(bNodes))
	fmt.Printf("mem Out edges:  %d bytes (%.2f MB)\n", bOut, toMB(bOut))
	fmt.Printf("mem In edges:   %d bytes (%.2f MB)\n", bIn, toMB(bIn))
	fmt.Printf("mem Pairs:      %d bytes (%.2f MB)\n", bPairs, toMB(bPairs))
	fmt.Printf("mem Paths hdrs: %d bytes (%.2f MB)\n", bPathsHdr, toMB(bPathsHdr))
	fmt.Printf("mem Paths elems:%d bytes (%.2f MB)\n", bPathsElems, toMB(bPathsElems))
	fmt.Printf("mem Allow:      %d bytes (%.2f MB)\n", bAllow, toMB(bAllow))
	fmt.Printf("mem TOTAL:      %d bytes (%.2f MB)\n", total, toMB(total))
}
