package main

type NodeID int32

// Role: 0=none, 1=trunk, 2=branch, 3=start, 4=end
type Node struct {
	Type int16
	Role byte
}

type Graph struct {
	Nodes      []Node
	Out        [][]NodeID // граф u -> v  out[u] - список исходящих вершин из u в v
	In         [][]NodeID // граф v <- u  in[v] - список входящих вершин в v из u
	Paths      []PairPath // пути которые будем искать
	Allow      [][]bool   // Allow[fromType][toType]
	SpikeEdges int        // число фейковых плечей
}

// Результирующий путь
type PairPath struct {
	Start NodeID
	End   NodeID
	Path  []NodeID
}
