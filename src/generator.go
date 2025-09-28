package main

import (
	picker "LogGen/src/picker"
	"math"
)

// Public entry points
func Generate(cfg Config) *Graph { return GenerateWithPicker(cfg, picker.NewRandomPicker(cfg.Seed)) }

func GenerateWithPicker(cfg Config, pick picker.Picker) *Graph {
    g := newGenerator(cfg, pick)
    ends := g.buildTrunks()
    g.attachTrees(ends)
    pairs := g.buildPairs()
    paths := g.buildPaths(pairs)

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
}

func newGenerator(cfg Config, picker picker.Picker) *generator {
	g := &generator{cfg: cfg, picker: picker}
	g.allow = prepareAllow(cfg)
	g.typeProb, g.trunkProb = prepareTypeProbs(cfg)

	/*
		est := estimateNodeCapacityPick(cfg, pick)

		if est < 1 {
			est = 1
		}

		g.nodes = make([]Node, 0, est)
		g.out = make([][]NodeID, 0, est)
		g.in = make([][]NodeID, 0, est)
	*/

	g.nodes = make([]Node, 0, 0)
	g.out = make([][]NodeID, 0, 0)
	g.in = make([][]NodeID, 0, 0)

	return g
}

func (g *generator) addNode(t int, role byte) NodeID {
	id := NodeID(len(g.nodes))
	g.nodes = append(g.nodes, Node{Type: int16(t), Role: role})
	g.out = append(g.out, nil)
	g.in = append(g.in, nil)

	return id
}

func (g *generator) addEdge(u, v NodeID) {
	g.out[u] = append(g.out[u], v)
	g.in[v] = append(g.in[v], u)
}

func (g *generator) setRole(id NodeID, role byte) { g.nodes[id].Role = role }

func (g *generator) getType(id NodeID) int { return int(g.nodes[id].Type) }

func (g *generator) graph(pairs [][2]NodeID, paths []PairPath) *Graph {
    return &Graph{Nodes: g.nodes, Out: g.out, In: g.in, Pairs: pairs, Allow: g.allow, Paths: paths}
}

func (g *generator) buildTrunks() [][2]NodeID {
	ends := make([][2]NodeID, g.cfg.NumTrunks)

	for m := 0; m < g.cfg.NumTrunks; m++ {
		startType := g.picker.WeightedGlobal(g.trunkProb)
		u := g.addNode(startType, 1)
		head := u

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
		}
		ends[m] = [2]NodeID{head, u}
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

		g.buildFeederTree(head, true)
		g.buildFeederTree(tail, false)
	}
}

func (g *generator) buildFeederTree(root NodeID, makeStarts bool) {
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
				if g.allow[nxtType][fromType] {
					g.addEdge(child, parent)
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
				} else {
					g.setRole(leaf, 4)
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
		L := g.picker.GeomLen(g.cfg.SpikeLenMean)
		prev := base
		pt := baseType

		for i := 0; i < L; i++ {
			nxt := g.picker.WeightedAllowed(pt, g.typeProb, g.allow)
			n := g.addNode(nxt, 2)

			if g.picker.CoinFlip() {
				if g.allow[nxt][int(pt)] {
					g.addEdge(n, prev)
				}
			} else {
				if g.allow[int(pt)][nxt] {
					g.addEdge(prev, n)
				}
			}

			prev = n
			pt = nxt
		}

		if g.picker.Bernoulli(g.cfg.DeadEndProb) {
			leafT := g.picker.WeightedAllowed(pt, g.typeProb, g.allow)
			leaf := g.addNode(leafT, 2)

			if g.allow[int(pt)][leafT] {
				g.addEdge(prev, leaf)
			}
		}
	}
}

func (g *generator) buildPairs() [][2]NodeID {
	if !g.cfg.MakePairs {
		return nil
	}

	starts := make([]NodeID, 0)
	ends := make([]NodeID, 0)

	for i := range g.nodes {
		switch g.nodes[i].Role {
		case 3:
			starts = append(starts, NodeID(i))
		case 4:
			ends = append(ends, NodeID(i))
		}
	}

	n := len(starts)
	if len(ends) < n {

		n = len(ends)
	}

	if n == 0 {
		return nil
	}

	pairs := make([][2]NodeID, 0, n)
	for i := 0; i < n; i++ {
		j := (i*7 + 3) % n
		pairs = append(pairs, [2]NodeID{starts[i], ends[j]})
	}

    return pairs
}

// buildPaths computes shortest (by edges) paths for each pair using BFS
func (g *generator) buildPaths(pairs [][2]NodeID) []PairPath {
    if len(pairs) == 0 {
        return nil
    }
    res := make([]PairPath, 0, len(pairs))
    for _, pr := range pairs {
        p := shortestPath(g.out, pr[0], pr[1])
        res = append(res, PairPath{A: pr[0], B: pr[1], Path: p})
    }
    return res
}

// shortestPath returns node sequence from start to goal via directed edges
func shortestPath(out [][]NodeID, start, goal NodeID) []NodeID {
    if start == goal {
        return []NodeID{start}
    }
    n := len(out)
    if n == 0 {
        return nil
    }
    visited := make([]bool, n)
    prev := make([]NodeID, n)
    for i := range prev {
        prev[i] = -1
    }
    q := make([]NodeID, 0, 16)
    visited[start] = true
    q = append(q, start)
    found := false

    for len(q) > 0 && !found {
        u := q[0]
        q = q[1:]
        for _, v := range out[u] {
            if !visited[v] {
                visited[v] = true
                prev[v] = u
                if v == goal {
                    found = true
                    break
                }
                q = append(q, v)
            }
        }
    }

    if !found {
        return nil
    }

    // reconstruct from goal to start
    path := make([]NodeID, 0, 16)
    cur := goal
    path = append(path, cur)
    for cur != start {
        cur = prev[cur]
        if cur == -1 { // safety
            return nil
        }
        path = append(path, cur)
    }
    // reverse
    for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
        path[i], path[j] = path[j], path[i]
    }
    return path
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

func estimateNodeCapacityPick(cfg Config, pick picker.Picker) int {
	m := cfg.BranchChildrenMean
	D := cfg.AvgBranchLen
	factor := 10.0
	if m > 1 {
		factor = math.Min(2000, math.Pow(m, D))
	}

	estTrunkNodes := cfg.NumTrunks * int(pick.GeomLen(cfg.AvgTrunkLen))
	estFeederNodes := int(float64(cfg.NumTrunks) * 2 * factor)
	estConfNodes := int(float64(estFeederNodes) * cfg.FakeBranchFac)
	estNodes := estTrunkNodes + estFeederNodes + estConfNodes + cfg.NumTrunks*2

	if estNodes < 1 {
		estNodes = 1
	}

	return estNodes
}
