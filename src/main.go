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
        if l > outMax { outMax = l }
    }
    for _, row := range g.In {
        l := len(row)
        inTotal += l
        if l > inMax { inMax = l }
    }

    fmt.Printf("Out size: rows=%d, total=%d, maxRow=%d\n", outRows, outTotal, outMax)
    fmt.Printf("In size:  rows=%d, total=%d, maxRow=%d\n", inRows, inTotal, inMax)
    fmt.Printf("generated graph with %d nodes, %d edges, %d pairs\n", len(g.Nodes), edges, len(g.Pairs))
}
