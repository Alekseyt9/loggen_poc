package main

import (
	"fmt"
	"unsafe"
)

// Заполняем матрицу возможных переходов (склеек)
func prepareAllow(cfg Config) [][]bool {
	if cfg.Allow != nil {
		return cfg.Allow
	}

	allow := make([][]bool, cfg.NumTypes)
	for i := 0; i < cfg.NumTypes; i++ {
		row := make([]bool, cfg.NumTypes)
		for j := 0; j < cfg.NumTypes; j++ {
			if i == j || j == i+1 || j == i-1 {
				row[j] = true
			}
		}
		allow[i] = row
	}

	return allow
}

// Инициализация из конфига
func prepareTypeProbs(cfg Config) (typeProb []float64, trunkProb []float64) {
	if len(cfg.TypeProb) == cfg.NumTypes {
		typeProb = cfg.TypeProb
	} else {
		typeProb = make([]float64, cfg.NumTypes)
		for i := range typeProb {
			typeProb[i] = 1
		}
	}

	if len(cfg.TrunkTypeProb) == cfg.NumTypes {
		trunkProb = cfg.TrunkTypeProb
	} else {
		trunkProb = typeProb
	}

	return
}

func graphMemBreakdown(g *Graph) (total int64) {
	var bNodes, bOut, bIn, bPairs, bPathsHdr, bPathsElems, bAllow int64

	if g == nil {
		return 0
	}

	szNode := int64(unsafe.Sizeof(Node{}))
	szNodeID := int64(unsafe.Sizeof(NodeID(0)))
	szPairPathHdr := int64(unsafe.Sizeof(PairPath{}))
	szBool := int64(unsafe.Sizeof(true))

	bNodes = int64(len(g.Nodes)) * szNode
	for i := range g.Out {
		bOut += int64(len(g.Out[i])) * szNodeID
	}
	for i := range g.In {
		bIn += int64(len(g.In[i])) * szNodeID
	}

	bPathsHdr = int64(len(g.Paths)) * szPairPathHdr
	for i := range g.Paths {
		bPathsElems += int64(len(g.Paths[i].Path)) * szNodeID
	}

	for i := range g.Allow {
		bAllow += int64(len(g.Allow[i])) * szBool
	}

	total = bNodes + bOut + bIn + bPairs + bPathsHdr + bPathsElems + bAllow

	return
}

func printInfo(g *Graph) {
	edges := 0

	for _, outs := range g.Out {
		edges += len(outs)
	}

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
	fmt.Printf("generated graph with %d nodes, %d edges, %d pairs\n", len(g.Nodes), edges, len(g.Paths))
	fmt.Printf("spike edges: %d\n", g.SpikeEdges)

	total := graphMemBreakdown(g)
	toMB := func(b int64) float64 { return float64(b) / (1024 * 1024) }
	fmt.Printf("mem TOTAL:      %d bytes (%.2f MB)\n", total, toMB(total))
}
