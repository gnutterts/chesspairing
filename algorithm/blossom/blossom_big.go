package blossom

// Edmonds' Blossom algorithm for maximum weight matching with big.Int weights.
//
// This is a port of MaxWeightMatching (blossom.go) adapted for multi-precision
// integer edge weights, needed because bbpPairings' edge weight bit layout
// requires >64 bits for realistic tournaments.
//
// Time complexity: O(n^3) where n = number of vertices.

import (
	"fmt"
	"math/big"
)

// BigEdge represents an undirected weighted edge with multi-precision weight.
type BigEdge struct {
	I, J   int      // vertex indices (0-based)
	Weight *big.Int // edge weight (higher = preferred)
}

// MaxWeightMatchingBig computes a maximum weight matching on a general
// undirected graph using Edmonds' Blossom algorithm with big.Int weights.
//
// edges: list of undirected weighted edges. Vertex indices are 0-based.
// maxCardinality: if true, maximize edges first, then weight.
//
// Returns a slice m where m[i] is the vertex matched to i, or -1 if
// unmatched.
func MaxWeightMatchingBig(edges []BigEdge, maxCardinality bool) []int {
	if len(edges) == 0 {
		return nil
	}

	nedge := len(edges)

	// Count vertices.
	nvertex := 0
	for _, e := range edges {
		if e.I+1 > nvertex {
			nvertex = e.I + 1
		}
		if e.J+1 > nvertex {
			nvertex = e.J + 1
		}
	}

	// Find maximum edge weight.
	maxweight := new(big.Int)
	for i, e := range edges {
		if e.Weight == nil {
			panic(fmt.Sprintf("blossom: BigEdge[%d].Weight is nil", i))
		}
		if e.Weight.Cmp(maxweight) > 0 {
			maxweight.Set(e.Weight)
		}
	}

	// endpoint[p]: vertex at endpoint p.
	endpoint := make([]int, 2*nedge)
	for k, e := range edges {
		endpoint[2*k] = e.I
		endpoint[2*k+1] = e.J
	}

	// neighbend[v]: list of remote endpoint indices for edges at vertex v.
	neighbend := make([][]int, nvertex)
	for k, e := range edges {
		neighbend[e.I] = append(neighbend[e.I], 2*k+1)
		neighbend[e.J] = append(neighbend[e.J], 2*k)
	}

	// mate[v]: remote endpoint index of v's matched edge, or -1.
	mate := make([]int, nvertex)
	for i := range mate {
		mate[i] = -1
	}

	label := make([]int, 2*nvertex)

	labelend := make([]int, 2*nvertex)
	for i := range labelend {
		labelend[i] = -1
	}

	inblossom := make([]int, nvertex)
	for i := range inblossom {
		inblossom[i] = i
	}

	blossomparent := make([]int, 2*nvertex)
	for i := range blossomparent {
		blossomparent[i] = -1
	}

	blossomchilds := make([][]int, 2*nvertex)
	blossomendps := make([][]int, 2*nvertex)

	blossombase := make([]int, 2*nvertex)
	for i := 0; i < nvertex; i++ {
		blossombase[i] = i
	}
	for i := nvertex; i < 2*nvertex; i++ {
		blossombase[i] = -1
	}

	bestedge := make([]int, 2*nvertex)
	for i := range bestedge {
		bestedge[i] = -1
	}

	blossombestedges := make([][]int, 2*nvertex)

	unusedblossoms := make([]int, 0, nvertex)
	for i := nvertex; i < 2*nvertex; i++ {
		unusedblossoms = append(unusedblossoms, i)
	}

	// Dual variables: big.Int.
	dualvar := make([]*big.Int, 2*nvertex)
	for i := 0; i < nvertex; i++ {
		dualvar[i] = new(big.Int).Set(maxweight)
	}
	for i := nvertex; i < 2*nvertex; i++ {
		dualvar[i] = new(big.Int)
	}

	allowedge := make([]bool, nedge)
	queue := make([]int, 0, nvertex)

	// Temporary big.Int values for slack computation to reduce allocations.
	tmpSlack := new(big.Int)

	// slack returns 2 * slack of edge k.
	slack := func(k int) *big.Int {
		// result = dualvar[i] + dualvar[j] - 2*weight
		r := new(big.Int).Add(dualvar[edges[k].I], dualvar[edges[k].J])
		tw := tmpSlack.Mul(edges[k].Weight, big.NewInt(2))
		r.Sub(r, tw)
		return r
	}

	var blossomLeaves func(b int) []int
	blossomLeaves = func(b int) []int {
		if b < nvertex {
			return []int{b}
		}
		var leaves []int
		for _, c := range blossomchilds[b] {
			leaves = append(leaves, blossomLeaves(c)...)
		}
		return leaves
	}

	var assignLabel func(w, t, p int)
	assignLabel = func(w, t, p int) {
		b := inblossom[w]
		label[w] = t
		label[b] = t
		labelend[w] = p
		labelend[b] = p
		bestedge[w] = -1
		bestedge[b] = -1
		switch t {
		case 1:
			queue = append(queue, blossomLeaves(b)...)
		case 2:
			base := blossombase[b]
			assignLabel(endpoint[mate[base]], 1, mate[base]^1)
		}
	}

	scanBlossom := func(v, w int) int {
		var path []int
		base := -1
		for v != -1 || w != -1 {
			b := inblossom[v]
			if label[b]&4 != 0 {
				base = blossombase[b]
				break
			}
			path = append(path, b)
			label[b] = 5
			if labelend[b] == -1 {
				v = -1
			} else {
				v = endpoint[labelend[b]]
				b = inblossom[v]
				v = endpoint[labelend[b]]
			}
			if w != -1 {
				v, w = w, v
			}
		}
		for _, b := range path {
			label[b] = 1
		}
		return base
	}

	addBlossom := func(base, k int) {
		v := edges[k].I
		w := edges[k].J
		bb := inblossom[base]
		bv := inblossom[v]
		bw := inblossom[w]

		b := unusedblossoms[len(unusedblossoms)-1]
		unusedblossoms = unusedblossoms[:len(unusedblossoms)-1]

		blossombase[b] = base
		blossomparent[b] = -1
		blossomparent[bb] = b

		path := make([]int, 0)
		endps := make([]int, 0)

		for bv != bb {
			blossomparent[bv] = b
			path = append(path, bv)
			endps = append(endps, labelend[bv])
			v = endpoint[labelend[bv]]
			bv = inblossom[v]
		}

		path = append(path, bb)
		reverseInts(path)
		reverseInts(endps)
		endps = append(endps, 2*k)

		for bw != bb {
			blossomparent[bw] = b
			path = append(path, bw)
			endps = append(endps, labelend[bw]^1)
			w = endpoint[labelend[bw]]
			bw = inblossom[w]
		}

		blossomchilds[b] = path
		blossomendps[b] = endps
		label[b] = 1
		labelend[b] = labelend[bb]
		dualvar[b] = new(big.Int)

		for _, lv := range blossomLeaves(b) {
			if label[inblossom[lv]] == 2 {
				queue = append(queue, lv)
			}
			inblossom[lv] = b
		}

		bestedgeto := make([]int, 2*nvertex)
		for i := range bestedgeto {
			bestedgeto[i] = -1
		}
		for _, bv := range path {
			var nblists [][]int
			if blossombestedges[bv] == nil {
				for _, lv := range blossomLeaves(bv) {
					edgeList := make([]int, len(neighbend[lv]))
					for idx, p := range neighbend[lv] {
						edgeList[idx] = p >> 1
					}
					nblists = append(nblists, edgeList)
				}
			} else {
				nblists = [][]int{blossombestedges[bv]}
			}
			for _, nblist := range nblists {
				for _, kk := range nblist {
					ii := edges[kk].I
					jj := edges[kk].J
					if inblossom[jj] == b {
						ii, jj = jj, ii
					}
					_ = ii
					bj := inblossom[jj]
					if bj != b && label[bj] == 1 &&
						(bestedgeto[bj] == -1 || slack(kk).Cmp(slack(bestedgeto[bj])) < 0) {
						bestedgeto[bj] = kk
					}
				}
			}
			blossombestedges[bv] = nil
			bestedge[bv] = -1
		}
		bestList := make([]int, 0)
		for _, kk := range bestedgeto {
			if kk != -1 {
				bestList = append(bestList, kk)
			}
		}
		blossombestedges[b] = bestList
		bestedge[b] = -1
		for _, kk := range bestList {
			if bestedge[b] == -1 || slack(kk).Cmp(slack(bestedge[b])) < 0 {
				bestedge[b] = kk
			}
		}
	}

	var expandBlossom func(b int, endstage bool)
	expandBlossom = func(b int, endstage bool) {
		for _, s := range blossomchilds[b] {
			blossomparent[s] = -1
			switch {
			case s < nvertex:
				inblossom[s] = s
			case endstage && dualvar[s].Sign() == 0:
				expandBlossom(s, endstage)
			default:
				for _, lv := range blossomLeaves(s) {
					inblossom[lv] = s
				}
			}
		}

		if !endstage && label[b] == 2 {
			entrychild := inblossom[endpoint[labelend[b]^1]]
			j := indexOf(blossomchilds[b], entrychild)

			var jstep, endptrick int
			if j&1 != 0 {
				j -= len(blossomchilds[b])
				jstep = 1
				endptrick = 0
			} else {
				jstep = -1
				endptrick = 1
			}

			p := labelend[b]
			for j != 0 {
				label[endpoint[p^1]] = 0
				label[endpoint[blossomendps[b][modLen(j-endptrick, len(blossomendps[b]))]^endptrick^1]] = 0
				assignLabel(endpoint[p^1], 2, p)
				allowedge[blossomendps[b][modLen(j-endptrick, len(blossomendps[b]))]>>1] = true
				j += jstep
				p = blossomendps[b][modLen(j-endptrick, len(blossomendps[b]))] ^ endptrick
				allowedge[p>>1] = true
				j += jstep
			}

			bv := blossomchilds[b][modLen(j, len(blossomchilds[b]))]
			label[endpoint[p^1]] = 2
			label[bv] = 2
			labelend[endpoint[p^1]] = p
			labelend[bv] = p
			bestedge[bv] = -1

			j += jstep
			for blossomchilds[b][modLen(j, len(blossomchilds[b]))] != entrychild {
				bv = blossomchilds[b][modLen(j, len(blossomchilds[b]))]
				if label[bv] == 1 {
					j += jstep
					continue
				}
				reachableV := -1
				for _, lv := range blossomLeaves(bv) {
					if label[lv] != 0 {
						reachableV = lv
						break
					}
				}
				if reachableV != -1 {
					label[reachableV] = 0
					label[endpoint[mate[blossombase[bv]]]] = 0
					assignLabel(reachableV, 2, labelend[reachableV])
				}
				j += jstep
			}
		}

		label[b] = -1
		labelend[b] = -1
		blossomchilds[b] = nil
		blossomendps[b] = nil
		blossombase[b] = -1
		blossombestedges[b] = nil
		bestedge[b] = -1
		unusedblossoms = append(unusedblossoms, b)
	}

	var augmentBlossom func(b, v int)
	augmentBlossom = func(b, v int) {
		t := v
		for blossomparent[t] != b {
			t = blossomparent[t]
		}
		if t >= nvertex {
			augmentBlossom(t, v)
		}

		i := indexOf(blossomchilds[b], t)
		j := i

		var jstep, endptrick int
		if i&1 != 0 {
			j -= len(blossomchilds[b])
			jstep = 1
			endptrick = 0
		} else {
			jstep = -1
			endptrick = 1
		}

		for j != 0 {
			j += jstep
			tt := blossomchilds[b][modLen(j, len(blossomchilds[b]))]
			p := blossomendps[b][modLen(j-endptrick, len(blossomendps[b]))] ^ endptrick
			if tt >= nvertex {
				augmentBlossom(tt, endpoint[p])
			}
			j += jstep
			tt = blossomchilds[b][modLen(j, len(blossomchilds[b]))]
			if tt >= nvertex {
				augmentBlossom(tt, endpoint[p^1])
			}
			mate[endpoint[p]] = p ^ 1
			mate[endpoint[p^1]] = p
		}

		blossomchilds[b] = rotateSlice(blossomchilds[b], i)
		blossomendps[b] = rotateSlice(blossomendps[b], i)
		blossombase[b] = blossombase[blossomchilds[b][0]]
	}

	augmentMatching := func(k int) {
		v := edges[k].I
		w := edges[k].J

		for _, sp := range [2]struct {
			s, p int
		}{
			{v, 2*k + 1},
			{w, 2 * k},
		} {
			s, p := sp.s, sp.p
			for {
				bs := inblossom[s]
				if bs >= nvertex {
					augmentBlossom(bs, s)
				}
				mate[s] = p
				if labelend[bs] == -1 {
					break
				}
				t := endpoint[labelend[bs]]
				bt := inblossom[t]
				s = endpoint[labelend[bt]]
				j := endpoint[labelend[bt]^1]
				if bt >= nvertex {
					augmentBlossom(bt, j)
				}
				mate[j] = labelend[bt]
				p = labelend[bt] ^ 1
			}
		}
	}

	// --- Main loop: iterate stages ---

	for t := 0; t < nvertex; t++ {
		for i := range label {
			label[i] = 0
		}
		for i := range bestedge {
			bestedge[i] = -1
		}
		for i := nvertex; i < 2*nvertex; i++ {
			blossombestedges[i] = nil
		}
		for i := range allowedge {
			allowedge[i] = false
		}
		queue = queue[:0]

		for v := 0; v < nvertex; v++ {
			if mate[v] == -1 && label[inblossom[v]] == 0 {
				assignLabel(v, 1, -1)
			}
		}

		augmented := false

	substageLoop:
		for {
			for len(queue) > 0 && !augmented {
				v := queue[len(queue)-1]
				queue = queue[:len(queue)-1]

				for _, p := range neighbend[v] {
					k := p >> 1
					w := endpoint[p]

					if inblossom[v] == inblossom[w] {
						continue
					}

					var kslack *big.Int
					if !allowedge[k] {
						kslack = slack(k)
						if kslack.Sign() <= 0 {
							allowedge[k] = true
						}
					}

					if allowedge[k] { //nolint:gocritic // nested if-else with break cannot be a switch
						switch label[inblossom[w]] { //nolint:gocritic // inner switch on label value
						case 0:
							assignLabel(w, 2, p^1)
						case 1:
							base := scanBlossom(v, w)
							if base >= 0 {
								addBlossom(base, k)
							} else {
								augmentMatching(k)
								augmented = true
							}
						default:
							if label[w] == 0 {
								label[w] = 2
								labelend[w] = p ^ 1
							}
						}
						if augmented {
							break
						}
					} else if label[inblossom[w]] == 1 {
						bi := inblossom[v]
						if bestedge[bi] == -1 || kslack.Cmp(slack(bestedge[bi])) < 0 {
							bestedge[bi] = k
						}
					} else if label[w] == 0 {
						if bestedge[w] == -1 || kslack.Cmp(slack(bestedge[w])) < 0 {
							bestedge[w] = k
						}
					}
				}
			}

			if augmented {
				break
			}

			// Compute deltas.
			deltatype := -1
			delta := new(big.Int)
			deltaedge := -1
			deltablossom := -1

			if !maxCardinality {
				deltatype = 1
				delta.Set(dualvar[0])
				for v := 0; v < nvertex; v++ {
					if dualvar[v].Cmp(delta) < 0 {
						delta.Set(dualvar[v])
					}
				}
			}

			for v := 0; v < nvertex; v++ {
				if label[inblossom[v]] == 0 && bestedge[v] != -1 {
					d := slack(bestedge[v])
					if deltatype == -1 || d.Cmp(delta) < 0 {
						delta.Set(d)
						deltatype = 2
						deltaedge = bestedge[v]
					}
				}
			}

			for b := 0; b < 2*nvertex; b++ {
				if blossomparent[b] == -1 && label[b] == 1 && bestedge[b] != -1 {
					d := new(big.Int).Rsh(slack(bestedge[b]), 1) // d = slack / 2
					if deltatype == -1 || d.Cmp(delta) < 0 {
						delta.Set(d)
						deltatype = 3
						deltaedge = bestedge[b]
					}
				}
			}

			for b := nvertex; b < 2*nvertex; b++ {
				if blossombase[b] >= 0 && blossomparent[b] == -1 && label[b] == 2 {
					if deltatype == -1 || dualvar[b].Cmp(delta) < 0 {
						delta.Set(dualvar[b])
						deltatype = 4
						deltablossom = b
					}
				}
			}

			if deltatype == -1 {
				deltatype = 1
				delta.SetInt64(0)
				for v := 0; v < nvertex; v++ {
					if dualvar[v].Cmp(delta) < 0 || delta.Sign() == 0 {
						delta.Set(dualvar[v])
					}
				}
				if delta.Sign() < 0 {
					delta.SetInt64(0)
				}
			}

			// Update dual variables.
			for v := 0; v < nvertex; v++ {
				switch label[inblossom[v]] {
				case 1:
					dualvar[v].Sub(dualvar[v], delta)
				case 2:
					dualvar[v].Add(dualvar[v], delta)
				}
			}
			for b := nvertex; b < 2*nvertex; b++ {
				if blossombase[b] >= 0 && blossomparent[b] == -1 {
					switch label[b] {
					case 1:
						dualvar[b].Add(dualvar[b], delta)
					case 2:
						dualvar[b].Sub(dualvar[b], delta)
					}
				}
			}

			switch deltatype {
			case 1:
				break substageLoop
			case 2:
				allowedge[deltaedge] = true
				ii := edges[deltaedge].I
				if label[inblossom[ii]] == 0 {
					ii = edges[deltaedge].J
				}
				queue = append(queue, ii)
			case 3:
				allowedge[deltaedge] = true
				ii := edges[deltaedge].I
				if label[inblossom[ii]] != 1 {
					ii = edges[deltaedge].J
				}
				queue = append(queue, ii)
			case 4:
				expandBlossom(deltablossom, false)
			}
		}

		if !augmented {
			break
		}

		for b := nvertex; b < 2*nvertex; b++ {
			if blossomparent[b] == -1 && blossombase[b] >= 0 &&
				label[b] == 1 && dualvar[b].Sign() == 0 {
				expandBlossom(b, true)
			}
		}
	}

	result := make([]int, nvertex)
	for v := 0; v < nvertex; v++ {
		if mate[v] >= 0 {
			result[v] = endpoint[mate[v]]
		} else {
			result[v] = -1
		}
	}
	return result
}
