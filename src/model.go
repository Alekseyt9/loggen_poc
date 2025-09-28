package main

// -------------------------------
// Data Models
// -------------------------------

type NodeID int32

// Role: 0=none, 1=trunk, 2=feeder, 3=start, 4=end
type Node struct {
	Type int16
	Role byte
}

type Graph struct {
    Nodes []Node
    Out   [][]NodeID // u -> v
    In    [][]NodeID // v <- u
    Pairs [][2]NodeID
    Paths []PairPath
    Allow [][]bool // Allow[fromType][toType]
    SpikeEdges int // number of edges created by fake branches (spikes)
}

// PairPath holds a start/end pair and the corresponding path between them
type PairPath struct {
    A    NodeID
    B    NodeID
    Path []NodeID
}
