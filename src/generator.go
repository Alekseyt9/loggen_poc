package main

import (
	picker "LogGen/src/picker"
	"math"
)

func Generate(cfg Config) *Graph {
	//return GenerateWithPicker(cfg, picker.NewRandomPicker(cfg.Seed))
	return GenerateWithPicker(cfg, picker.NewDeterministicPicker(cfg.NumTypes))
}

func GenerateWithPicker(cfg Config, pick picker.Picker) *Graph {
	g := newGenerator(cfg, pick)
	ends := g.buildTrunks()
	g.attachTrees(ends)
	pairs, paths := g.buildPairs()

	return g.graph(pairs, paths)
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

	parentHead []NodeID   // for head-side feeders: child -> parent (towards head)
	parentTail []NodeID   // for tail-side feeders: child -> parent (towards tail)
	isSpike    []bool     // nodes created as fake spikes
	trunks     [][]NodeID // nodes along each trunk from head to tail
    starts     [][]NodeID // per-trunk start leaves (head side)
    ends       [][]NodeID // per-trunk end leaves (tail side)
    spikeEdges int        // count of edges added in fake branches
}

func newGenerator(cfg Config, picker picker.Picker) *generator {
	g := &generator{cfg: cfg, picker: picker}
	g.allow = prepareAllow(cfg)
	g.typeProb, g.trunkProb = prepareTypeProbs(cfg)

	g.nodes = make([]Node, 0, 0)
	g.out = make([][]NodeID, 0, 0)
	g.in = make([][]NodeID, 0, 0)
	g.parentHead = make([]NodeID, 0, 0)
	g.parentTail = make([]NodeID, 0, 0)
	g.isSpike = make([]bool, 0, 0)
	g.trunks = make([][]NodeID, 0, cfg.NumTrunks)
	g.starts = make([][]NodeID, cfg.NumTrunks)
	g.ends = make([][]NodeID, cfg.NumTrunks)

	return g
}

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

func (g *generator) graph(pairs [][2]NodeID, paths []PairPath) *Graph {
    return &Graph{Nodes: g.nodes, Out: g.out, In: g.in, Pairs: pairs, Allow: g.allow, Paths: paths, SpikeEdges: g.spikeEdges}
}

func (g *generator) buildTrunks() [][2]NodeID {
	ends := make([][2]NodeID, g.cfg.NumTrunks)

	for m := 0; m < g.cfg.NumTrunks; m++ {
		startType := g.picker.WeightedGlobal(g.trunkProb)
		u := g.addNode(startType, 1)
		head := u
		seq := make([]NodeID, 0, 8)
		seq = append(seq, u)

		L := g.picker.GeomLen(g.cfg.AvgTrunkLen)
		if g.cfg.TrunkLenMin > 0 && L < g.cfg.TrunkLenMin {
			L = g.cfg.TrunkLenMin
		}

		prevType := int(startType)
		for i := 1; i < L; i++ {
			nextType := g.picker.WeightedAllowed(prevType, g.trunkProb, g.allow)
			v := g.addNode(nextType, 1)
			if g.allow[prevType][nextType] {
				g.addEdge(u, v)
			}

			u = v
			prevType = nextType
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

		g.buildFeederTree(m, head, true)
		g.buildFeederTree(m, tail, false)
	}
}

func (g *generator) buildFeederTree(trunkIdx int, root NodeID, makeStarts bool) {
	depth := g.picker.GeomLen(g.cfg.AvgBranchLen)
	if depth < 1 {
		depth = 1
	}

	frontier := []NodeID{root}

	for level := 1; level <= depth; level++ {
		nextFrontier := make([]NodeID, 0, len(frontier))
		for _, parent := range frontier {
			fromType := g.getType(parent)
			add := g.picker.Poisson(math.Max(g.cfg.BranchChildrenMean-1, 0.0))
			children := 1 + add

			for c := 0; c < children; c++ {
				nxtType := g.picker.WeightedAllowed(fromType, g.typeProb, g.allow)
				child := g.addNode(nxtType, 2)
				if makeStarts {
					// head-side: edges child -> parent (towards head)
					if g.allow[nxtType][fromType] {
						g.addEdge(child, parent)
						g.parentHead[child] = parent
					}
				} else {
					// tail-side: edges parent -> child (away from tail)
					if g.allow[fromType][nxtType] {
						g.addEdge(parent, child)
						g.parentTail[child] = parent
					}
				}
				if g.cfg.FakeBranchFac > 0 {
					g.addFakeBranches(child, nxtType)
				}
				nextFrontier = append(nextFrontier, child)
			}
		}

		if level == depth {
			for _, leaf := range nextFrontier {
				if makeStarts {
					g.setRole(leaf, 3)
					g.starts[trunkIdx] = append(g.starts[trunkIdx], leaf)
				} else {
					g.setRole(leaf, 4)
					g.ends[trunkIdx] = append(g.ends[trunkIdx], leaf)
				}
			}
		}

		frontier = nextFrontier
	}
}

func (g *generator) addFakeBranches(base NodeID, baseType int) {
	tries := 0

	for g.picker.Bernoulli(g.cfg.FakeBranchFac) && tries < 3 {
		tries++
		L := g.picker.GeomLen(g.cfg.FakeBranchLenMean)
		prev := base
		pt := baseType

        for i := 0; i < L; i++ {
            nxt := g.picker.WeightedAllowed(pt, g.typeProb, g.allow)
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
            leafT := g.picker.WeightedAllowed(pt, g.typeProb, g.allow)
            leaf := g.addNode(leafT, 2)
            g.isSpike[leaf] = true

            if g.allow[int(pt)][leafT] {
                g.addEdge(prev, leaf)
                g.spikeEdges++
            }
        }
    }
}

func (g *generator) buildPairs() (pairs [][2]NodeID, paths []PairPath) {
	if !g.cfg.MakePairs {
		return nil, nil
	}

	totalCap := 0
	for m := range g.trunks {
		a := len(g.starts[m])
		b := len(g.ends[m])
		if a < b {
			totalCap += a
		} else {
			totalCap += b
		}
	}

	if totalCap == 0 {
		return nil, nil
	}

	pairs = make([][2]NodeID, 0, totalCap)
	paths = make([]PairPath, 0, totalCap)

	for m := range g.trunks {
		s := g.starts[m]
		e := g.ends[m]
		n := len(s)

		if len(e) < n {
			n = len(e)
		}

		if n == 0 {
			continue
		}

		for i := 0; i < n; i++ {
			j := (i*7 + 3) % n
			a := s[i]
			b := e[j]
			p := g.buildPathForPair(m, a, b)
			pairs = append(pairs, [2]NodeID{a, b})
			paths = append(paths, PairPath{A: a, B: b, Path: p})
		}
	}

	return pairs, paths
}

func (g *generator) buildPathForPair(trunkIdx int, start NodeID, end NodeID) []NodeID {
	trunk := g.trunks[trunkIdx]
	head := trunk[0]
	tail := trunk[len(trunk)-1]

	// segment A: start -> ... -> head
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

	// segment B: trunk head..tail
	segB := trunk
	// segment C: tail .. end (walk end->tail via parentTail, then reverse)
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

	// reverse tmp to get tail->...->end
	for i, j := 0, len(tmp)-1; i < j; i, j = i+1, j-1 {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}
	segC := tmp

	// concat: A + B[1:] + C[1:]
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

// Заполняем матрицу возможных переходов
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
