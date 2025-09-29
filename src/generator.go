package main

import (
	picker "LogGen/src/picker"
	"math"
	"slices"
)

func Generate(cfg Config) *Graph {
	//return GenerateWithPicker(cfg, picker.NewRandomPicker(cfg.Seed))
	return GenerateWithPicker(cfg, picker.NewDeterministicPicker(cfg.NumTypes))
}

func GenerateWithPicker(cfg Config, pick picker.Picker) *Graph {
	g := newGenerator(cfg, pick)
	ends := g.buildTrunks()
	g.attachTrees(ends)
	paths := g.buildPaths()

	return g.graph(paths)
}

// generator encapsulates mutable state and helpers
type generator struct {
	cfg       Config
	picker    picker.Picker
	allow     [][]bool
	typeProb  []float64
	trunkProb []float64
	nodes     []Node
	out       [][]NodeID
	in        [][]NodeID

	parentHead []NodeID   // child -> parent (к head)
	parentTail []NodeID   // child -> parent (к tail)
	isSpike    []bool     // фейковые узлы
	trunks     [][]NodeID // узлы вдоль каждого ствола от головы до хвоста
	starts     [][]NodeID // стартовые точки
	ends       [][]NodeID // конечные точки
	spikeEdges int        // число фейковых узлов (для статистики)
}

func newGenerator(cfg Config, picker picker.Picker) *generator {
	g := &generator{cfg: cfg, picker: picker}
	g.allow = prepareAllow(cfg)
	g.typeProb, g.trunkProb = prepareTypeProbs(cfg)

	g.nodes = make([]Node, 0)
	g.out = make([][]NodeID, 0)
	g.in = make([][]NodeID, 0)
	g.parentHead = make([]NodeID, 0)
	g.parentTail = make([]NodeID, 0)
	g.isSpike = make([]bool, 0)
	g.trunks = make([][]NodeID, 0, cfg.NumTrunks)
	g.starts = make([][]NodeID, cfg.NumTrunks)
	g.ends = make([][]NodeID, cfg.NumTrunks)

	return g
}

// построить магистрали
func (g *generator) buildTrunks() [][2]NodeID {
	ends := make([][2]NodeID, g.cfg.NumTrunks)

	for m := 0; m < g.cfg.NumTrunks; m++ {
		trunkType := int(g.picker.WeightedGlobal(g.trunkProb))
		u := g.addNode(trunkType, 1)
		head := u
		seq := make([]NodeID, 0, 8)
		seq = append(seq, u)

		length := g.picker.GeomLen(g.cfg.AvgTrunkLen)

		for i := 1; i < length; i++ {
			v := g.addNode(trunkType, 1)
			g.addEdge(u, v)
			u = v
			seq = append(seq, v)
		}

		ends[m] = [2]NodeID{head, u}
		g.trunks = append(g.trunks, seq)
	}

	return ends
}

func (g *generator) attachTrees(trunkEnds [][2]NodeID) {
	for m := 0; m < g.cfg.NumTrunks; m++ {
		head := trunkEnds[m][0]
		tail := trunkEnds[m][1]

		if g.cfg.StartOnTrunkProb > 0 && g.picker.Bernoulli(g.cfg.StartOnTrunkProb) {
			g.setRole(head, 3)
		}

		if g.cfg.EndOnTrunkProb > 0 && g.picker.Bernoulli(g.cfg.EndOnTrunkProb) {
			g.setRole(tail, 4)
		}

		g.buildTree(m, head, true)
		g.buildTree(m, tail, false)
	}
}

func (g *generator) buildTree(trunkIdx int, root NodeID, makeStarts bool) {
	depth := g.picker.GeomLen(g.cfg.AvgBranchLen)
	if depth < 1 {
		depth = 1
	}

	g.buildTreeRec(trunkIdx, root, g.getType(root), depth, makeStarts)
	return
}

// построение дерева
func (g *generator) buildTreeRec(trunkIdx int, parent NodeID, fromType int, depthLeft int, makeStarts bool) {
	if depthLeft <= 0 {
		if makeStarts {
			g.setRole(parent, 3)
			g.starts[trunkIdx] = append(g.starts[trunkIdx], parent)
		} else {
			g.setRole(parent, 4)
			g.ends[trunkIdx] = append(g.ends[trunkIdx], parent)
		}
		return
	}

	add := g.picker.Poisson(math.Max(g.cfg.BranchChildrenMean-1, 0.0))
	children := 1 + add

	for c := 0; c < children; c++ {
		nxtType, ok := g.picker.WeightedAllowed(fromType, g.typeProb, g.allow)
		if !ok {
			if makeStarts {
				g.setRole(parent, 3)
				g.starts[trunkIdx] = append(g.starts[trunkIdx], parent)
			} else {
				g.setRole(parent, 4)
				g.ends[trunkIdx] = append(g.ends[trunkIdx], parent)
			}
			break
		}

		child := g.addNode(nxtType, 2)
		if makeStarts {
			// от стартовой точки child -> parent к голове
			if g.allow[nxtType][fromType] {
				g.addEdge(child, parent)
				g.parentHead[child] = parent
			}
		} else {
			// до конечной точки edges parent -> child от хвоста
			if g.allow[fromType][nxtType] {
				g.addEdge(parent, child)
				g.parentTail[child] = parent
			}
		}

		if g.cfg.FakeBranchFac > 0 {
			g.addFakeBranches(child, nxtType)
		}

		g.buildTreeRec(trunkIdx, child, nxtType, depthLeft-1, makeStarts)
	}
}

func (g *generator) addFakeBranches(base NodeID, baseType int) {
	tries := 0

	for g.picker.Bernoulli(g.cfg.FakeBranchFac) && tries < 3 {
		tries++
		length := g.picker.GeomLen(g.cfg.FakeBranchLenMean)
		prev := base
		pt := baseType

		for i := 0; i < length; i++ {
			nxt, ok := g.picker.WeightedAllowed(pt, g.typeProb, g.allow)
			if !ok {
				break
			}
			n := g.addNode(nxt, 2)
			g.isSpike[n] = true

			if g.picker.CoinFlip() {
				if g.allow[nxt][int(pt)] {
					g.addEdge(n, prev)
					g.spikeEdges++
				}
			} else {
				if g.allow[int(pt)][nxt] {
					g.addEdge(prev, n)
					g.spikeEdges++
				}
			}

			prev = n
			pt = nxt
		}

		if g.picker.Bernoulli(g.cfg.DeadEndProb) {
			leafT, ok := g.picker.WeightedAllowed(pt, g.typeProb, g.allow)
			if !ok {
				continue
			}
			leaf := g.addNode(leafT, 2)
			g.isSpike[leaf] = true

			if g.allow[int(pt)][leafT] {
				g.addEdge(prev, leaf)
				g.spikeEdges++
			}
		}
	}
}

func (g *generator) buildPaths() (paths []PairPath) {
	paths = make([]PairPath, 0)

	for m := range g.trunks {
		starts := g.starts[m]
		ends := g.ends[m]
		length := min(len(starts), len(ends))

		for i := 0; i < length; i++ {
			j := (i*7 + 3) % length
			start := starts[i]
			end := ends[j]
			path := g.buildPathForPair(m, start, end)
			paths = append(paths, PairPath{Start: start, End: end, Path: path})
		}
	}

	return paths
}

func (g *generator) buildPathForPair(trunkIdx int, start NodeID, end NodeID) []NodeID {
	trunk := g.trunks[trunkIdx]
	head := trunk[0]
	tail := trunk[len(trunk)-1]

	// сегмент A: start -> ... -> head
	segA := make([]NodeID, 0, 8)
	cur := start
	for {
		segA = append(segA, cur)
		if cur == head {
			break
		}
		cur = g.parentHead[cur]
		if cur == -1 {
			return nil
		}
	}

	// сегмент B: trunk head..tail
	segB := trunk

	// сегмент C: tail .. end (идет end->tail через parentTail, потом назад)
	tmp := make([]NodeID, 0, 8)
	cur = end
	for {
		tmp = append(tmp, cur)
		if cur == tail {
			break
		}
		cur = g.parentTail[cur]
		if cur == -1 {
			return nil
		}
	}

	// переворачиваем tail->...->end
	slices.Reverse(tmp)
	segC := tmp

	// соединяем: A + B[1:] + C[1:]
	res := make([]NodeID, 0, len(segA)+len(segB)+len(segC)-2)
	res = append(res, segA...)
	if len(segB) > 1 {
		res = append(res, segB[1:]...)
	}
	if len(segC) > 1 {
		res = append(res, segC[1:]...)
	}

	return res
}

// вспомогательные методы
func (g *generator) addNode(t int, role byte) NodeID {
	id := NodeID(len(g.nodes))
	g.nodes = append(g.nodes, Node{Type: int16(t), Role: role})
	g.out = append(g.out, nil)
	g.in = append(g.in, nil)
	g.parentHead = append(g.parentHead, -1)
	g.parentTail = append(g.parentTail, -1)
	g.isSpike = append(g.isSpike, false)

	return id
}

func (g *generator) addEdge(u, v NodeID) {
	g.out[u] = append(g.out[u], v)
	g.in[v] = append(g.in[v], u)
}

func (g *generator) setRole(id NodeID, role byte) { g.nodes[id].Role = role }

func (g *generator) getType(id NodeID) int { return int(g.nodes[id].Type) }

func (g *generator) graph(paths []PairPath) *Graph {
	return &Graph{Nodes: g.nodes, Out: g.out, In: g.in, Allow: g.allow, Paths: paths, SpikeEdges: g.spikeEdges}
}
