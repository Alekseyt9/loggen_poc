package main

import "unsafe"

func GraphMemBreakdown(g *Graph) (total int64, bNodes, bOut, bIn, bPairs, bPathsHdr, bPathsElems, bAllow int64) {
	if g == nil {
		return 0, 0, 0, 0, 0, 0, 0, 0
	}
	szNode := int64(unsafe.Sizeof(Node{}))
	szNodeID := int64(unsafe.Sizeof(NodeID(0)))
	szPair := int64(unsafe.Sizeof([2]NodeID{}))
	szPairPathHdr := int64(unsafe.Sizeof(PairPath{}))
	szBool := int64(unsafe.Sizeof(true))

	bNodes = int64(len(g.Nodes)) * szNode
	for i := range g.Out {
		bOut += int64(len(g.Out[i])) * szNodeID
	}
	for i := range g.In {
		bIn += int64(len(g.In[i])) * szNodeID
	}

	bPairs = int64(len(g.Pairs)) * szPair
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
