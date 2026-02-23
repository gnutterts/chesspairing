// Package blossom implements Edmonds' maximum weight matching algorithm
// for general graphs. It supports both int64 weights (MaxWeightMatching)
// and arbitrary-precision *big.Int weights (MaxWeightMatchingBig).
//
// This is a Go port of Joris van Rantwijk's Python reference implementation.
package blossom

// Edmonds' Blossom algorithm for maximum weight matching in general graphs.
//
// Ported from Joris van Rantwijk's Python reference implementation
// (mwmatching.py, public domain). The algorithm handles odd cycles
// (blossoms) by contracting them to super-vertices.
//
// Time complexity: O(n^3) where n = number of vertices.

// BlossomEdge represents an undirected weighted edge in the matching graph.
type BlossomEdge struct {
	I, J   int   // vertex indices (0-based)
	Weight int64 // edge weight (higher = preferred)
}

// MaxWeightMatching computes a maximum weight matching on a general
// undirected graph using Edmonds' Blossom algorithm.
//
// edges: list of undirected weighted edges. Vertex indices are 0-based
// and inferred from the edges (max index + 1 = number of vertices).
//
// maxCardinality: if true, finds matching with maximum number of edges
// first, then maximizes weight among those. If false, maximizes weight
// only (may leave matchable vertices unmatched).
//
// Returns a slice m where m[i] is the vertex matched to i, or -1 if
// unmatched. The slice length equals the number of vertices.
func MaxWeightMatching(edges []BlossomEdge, maxCardinality bool) []int {
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
	maxweight := int64(0)
	for _, e := range edges {
		if e.Weight > maxweight {
			maxweight = e.Weight
		}
	}

	// endpoint[p]: vertex at endpoint p. Edge k has endpoints 2*k and 2*k+1.
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

	// label[b]: 0=unlabeled, 1=S, 2=T.
	label := make([]int, 2*nvertex)

	// labelend[b]: remote endpoint through which b got its label, or -1.
	labelend := make([]int, 2*nvertex)
	for i := range labelend {
		labelend[i] = -1
	}

	// inblossom[v]: top-level blossom containing vertex v.
	inblossom := make([]int, nvertex)
	for i := range inblossom {
		inblossom[i] = i
	}

	// blossomparent[b]: parent blossom, or -1.
	blossomparent := make([]int, 2*nvertex)
	for i := range blossomparent {
		blossomparent[i] = -1
	}

	// blossomchilds[b]: ordered child list for non-trivial blossom b.
	blossomchilds := make([][]int, 2*nvertex)

	// blossomendps[b]: endpoint list linking children of blossom b.
	blossomendps := make([][]int, 2*nvertex)

	// blossombase[b]: base vertex of blossom b.
	blossombase := make([]int, 2*nvertex)
	for i := 0; i < nvertex; i++ {
		blossombase[i] = i
	}
	for i := nvertex; i < 2*nvertex; i++ {
		blossombase[i] = -1
	}

	// bestedge[b]: least-slack edge from b to different S-blossom, or -1.
	bestedge := make([]int, 2*nvertex)
	for i := range bestedge {
		bestedge[i] = -1
	}

	// blossombestedges[b]: list of least-slack edges to neighbouring S-blossoms.
	blossombestedges := make([][]int, 2*nvertex)

	// Pool of unused blossom IDs.
	unusedblossoms := make([]int, 0, nvertex)
	for i := nvertex; i < 2*nvertex; i++ {
		unusedblossoms = append(unusedblossoms, i)
	}

	// Dual variables.
	dualvar := make([]int64, 2*nvertex)
	for i := 0; i < nvertex; i++ {
		dualvar[i] = maxweight
	}

	// allowedge[k]: true if edge k has zero slack.
	allowedge := make([]bool, nedge)

	// Queue of S-vertices to scan.
	queue := make([]int, 0, nvertex)

	// --- Helper functions ---

	// Return 2 * slack of edge k (works outside blossoms only).
	slack := func(k int) int64 {
		return dualvar[edges[k].I] + dualvar[edges[k].J] - 2*edges[k].Weight
	}

	// blossomLeaves yields all base vertices in blossom b.
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

	// assignLabel labels a top-level blossom and queues vertices.
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
			// S-blossom: add vertices to queue.
			queue = append(queue, blossomLeaves(b)...)
		case 2:
			// T-blossom: label its mate as S.
			base := blossombase[b]
			assignLabel(endpoint[mate[base]], 1, mate[base]^1)
		}
	}

	// scanBlossom traces back from v and w to find a common ancestor.
	// Returns base vertex of new blossom, or -1 (augmenting path).
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
			// Trace one step back.
			if labelend[b] == -1 {
				v = -1
			} else {
				v = endpoint[labelend[b]]
				b = inblossom[v]
				// b is a T-blossom; trace one more step back.
				v = endpoint[labelend[b]]
			}
			if w != -1 {
				v, w = w, v
			}
		}
		// Remove breadcrumbs.
		for _, b := range path {
			label[b] = 1
		}
		return base
	}

	// addBlossom creates a new blossom from edge k with given base.
	addBlossom := func(base, k int) {
		v := edges[k].I
		w := edges[k].J
		bb := inblossom[base]
		bv := inblossom[v]
		bw := inblossom[w]

		// Allocate a new blossom number.
		b := unusedblossoms[len(unusedblossoms)-1]
		unusedblossoms = unusedblossoms[:len(unusedblossoms)-1]

		blossombase[b] = base
		blossomparent[b] = -1
		blossomparent[bb] = b

		// Build child and endpoint lists.
		// Python: blossomchilds[b] = path = []
		//         blossomendps[b] = endps = []
		path := make([]int, 0)
		endps := make([]int, 0)

		// Trace back from v to base.
		for bv != bb {
			blossomparent[bv] = b
			path = append(path, bv)
			endps = append(endps, labelend[bv])
			v = endpoint[labelend[bv]]
			bv = inblossom[v]
		}

		// Add base blossom, reverse path so it goes base → ... → v.
		path = append(path, bb)
		reverseInts(path)
		reverseInts(endps)

		// Add edge endpoint connecting v-side to w-side.
		endps = append(endps, 2*k)

		// Trace back from w to base.
		for bw != bb {
			blossomparent[bw] = b
			path = append(path, bw)
			endps = append(endps, labelend[bw]^1)
			w = endpoint[labelend[bw]]
			bw = inblossom[w]
		}

		blossomchilds[b] = path
		blossomendps[b] = endps

		// Set label to S.
		label[b] = 1
		labelend[b] = labelend[bb]

		// Set dual variable to zero.
		dualvar[b] = 0

		// Relabel vertices: T-vertices inside become S-vertices.
		for _, lv := range blossomLeaves(b) {
			if label[inblossom[lv]] == 2 {
				queue = append(queue, lv)
			}
			inblossom[lv] = b
		}

		// Compute blossombestedges[b].
		bestedgeto := make([]int, 2*nvertex)
		for i := range bestedgeto {
			bestedgeto[i] = -1
		}
		for _, bv := range path {
			var nblists [][]int
			if blossombestedges[bv] == nil {
				// Get edge info from vertices.
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
					jj := edges[kk].J
					if inblossom[jj] == b {
						jj = edges[kk].I
					}
					bj := inblossom[jj]
					if bj != b && label[bj] == 1 &&
						(bestedgeto[bj] == -1 || slack(kk) < slack(bestedgeto[bj])) {
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
			if bestedge[b] == -1 || slack(kk) < slack(bestedge[b]) {
				bestedge[b] = kk
			}
		}
	}

	// expandBlossom expands blossom b.
	var expandBlossom func(b int, endstage bool)
	expandBlossom = func(b int, endstage bool) {
		// Convert sub-blossoms into top-level blossoms.
		for _, s := range blossomchilds[b] {
			blossomparent[s] = -1
			switch {
			case s < nvertex:
				inblossom[s] = s
			case endstage && dualvar[s] == 0:
				// Recursively expand zero-dual sub-blossom.
				expandBlossom(s, endstage)
			default:
				for _, lv := range blossomLeaves(s) {
					inblossom[lv] = s
				}
			}
		}

		// If expanding a T-blossom during a stage, relabel sub-blossoms.
		if !endstage && label[b] == 2 {
			// Find the entry child (sub-blossom through which b got its label).
			entrychild := inblossom[endpoint[labelend[b]^1]]

			// Find entrychild position.
			j := indexOf(blossomchilds[b], entrychild)

			// Decide direction.
			var jstep, endptrick int
			if j&1 != 0 {
				// Odd index: go forward and wrap.
				j -= len(blossomchilds[b])
				jstep = 1
				endptrick = 0
			} else {
				// Even index: go backward.
				jstep = -1
				endptrick = 1
			}

			// Move along the blossom until we get to the base.
			p := labelend[b]
			for j != 0 {
				// Relabel the T-sub-blossom.
				label[endpoint[p^1]] = 0
				label[endpoint[blossomendps[b][modLen(j-endptrick, len(blossomendps[b]))]^endptrick^1]] = 0
				assignLabel(endpoint[p^1], 2, p)
				// Step to the next S-sub-blossom.
				allowedge[blossomendps[b][modLen(j-endptrick, len(blossomendps[b]))]>>1] = true
				j += jstep
				p = blossomendps[b][modLen(j-endptrick, len(blossomendps[b]))] ^ endptrick
				// Step to the next T-sub-blossom.
				allowedge[p>>1] = true
				j += jstep
			}

			// Relabel the base T-sub-blossom (don't call assignLabel).
			bv := blossomchilds[b][modLen(j, len(blossomchilds[b]))]
			label[endpoint[p^1]] = 2
			label[bv] = 2
			labelend[endpoint[p^1]] = p
			labelend[bv] = p
			bestedge[bv] = -1

			// Continue along the blossom until we get back to entrychild.
			j += jstep
			for blossomchilds[b][modLen(j, len(blossomchilds[b]))] != entrychild {
				bv = blossomchilds[b][modLen(j, len(blossomchilds[b]))]
				if label[bv] == 1 {
					// Already labeled S through a neighbour; skip.
					j += jstep
					continue
				}
				// Check if any vertex in this sub-blossom is reachable.
				var reachableV = -1
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

		// Recycle the blossom number.
		label[b] = -1
		labelend[b] = -1
		blossomchilds[b] = nil
		blossomendps[b] = nil
		blossombase[b] = -1
		blossombestedges[b] = nil
		bestedge[b] = -1
		unusedblossoms = append(unusedblossoms, b)
	}

	// augmentBlossom swaps matched/unmatched edges in blossom b
	// so that vertex v becomes the new base.
	var augmentBlossom func(b, v int)
	augmentBlossom = func(b, v int) {
		// Bubble up to an immediate sub-blossom of b.
		t := v
		for blossomparent[t] != b {
			t = blossomparent[t]
		}
		// Recursively deal with the first sub-blossom.
		if t >= nvertex {
			augmentBlossom(t, v)
		}

		// Find t's position in the child list.
		i := indexOf(blossomchilds[b], t)
		j := i

		// Decide direction.
		var jstep, endptrick int
		if i&1 != 0 {
			j -= len(blossomchilds[b])
			jstep = 1
			endptrick = 0
		} else {
			jstep = -1
			endptrick = 1
		}

		// Move along the blossom until we get to the base.
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
			// Match the edge connecting those sub-blossoms.
			mate[endpoint[p]] = p ^ 1
			mate[endpoint[p^1]] = p
		}

		// Rotate child/endpoint lists so that t is first.
		blossomchilds[b] = rotateSlice(blossomchilds[b], i)
		blossomendps[b] = rotateSlice(blossomendps[b], i)
		blossombase[b] = blossombase[blossomchilds[b][0]]
	}

	// augmentMatching augments along the path through edge k.
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
				// Augment through the S-blossom from s to base.
				if bs >= nvertex {
					augmentBlossom(bs, s)
				}
				// Update mate[s].
				mate[s] = p
				// Trace one step back.
				if labelend[bs] == -1 {
					break
				}
				t := endpoint[labelend[bs]]
				bt := inblossom[t]
				// Trace one step back through T-blossom.
				s = endpoint[labelend[bt]]
				j := endpoint[labelend[bt]^1]
				// Augment through the T-blossom from j to base.
				if bt >= nvertex {
					augmentBlossom(bt, j)
				}
				// Update mate[j].
				mate[j] = labelend[bt]
				// Keep the opposite endpoint for next iteration.
				p = labelend[bt] ^ 1
			}
		}
	}

	// --- Main loop: iterate stages ---

	for t := 0; t < nvertex; t++ {
		// Reset labels.
		for i := range label {
			label[i] = 0
		}

		// Forget least-slack edges.
		for i := range bestedge {
			bestedge[i] = -1
		}
		for i := nvertex; i < 2*nvertex; i++ {
			blossombestedges[i] = nil
		}

		// Reset allowedge.
		for i := range allowedge {
			allowedge[i] = false
		}

		// Empty queue.
		queue = queue[:0]

		// Label single blossoms/vertices with S.
		for v := 0; v < nvertex; v++ {
			if mate[v] == -1 && label[inblossom[v]] == 0 {
				assignLabel(v, 1, -1)
			}
		}

		augmented := false

	substageLoop:
		for {
			// Substage: scan S-vertices.
			for len(queue) > 0 && !augmented {
				v := queue[len(queue)-1]
				queue = queue[:len(queue)-1]

				for _, p := range neighbend[v] {
					k := p >> 1
					w := endpoint[p]

					if inblossom[v] == inblossom[w] {
						continue
					}

					kslack := int64(0)
					if !allowedge[k] {
						kslack = slack(k)
						if kslack <= 0 {
							allowedge[k] = true
						}
					}

					if allowedge[k] { //nolint:gocritic // nested if-else with break cannot be a switch
						switch label[inblossom[w]] { //nolint:gocritic // inner switch on label value
						case 0:
							// w is free; label T.
							assignLabel(w, 2, p^1)
						case 1:
							// S meets S: blossom or augmenting path.
							base := scanBlossom(v, w)
							if base >= 0 {
								addBlossom(base, k)
							} else {
								augmentMatching(k)
								augmented = true
							}
						default:
							if label[w] == 0 {
								// w inside T-blossom but not yet reached.
								label[w] = 2
								labelend[w] = p ^ 1
							}
						}
						if augmented {
							break
						}
					} else if label[inblossom[w]] == 1 {
						// Keep track of least-slack edge to different S-blossom.
						bi := inblossom[v]
						if bestedge[bi] == -1 || kslack < slack(bestedge[bi]) {
							bestedge[bi] = k
						}
					} else if label[w] == 0 {
						// Track least-slack edge to free/unreached vertex.
						if bestedge[w] == -1 || kslack < slack(bestedge[w]) {
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
			delta := int64(0)
			deltaedge := -1
			deltablossom := -1

			// Delta type 1: minimum dual of any S-vertex.
			if !maxCardinality {
				deltatype = 1
				delta = dualvar[0]
				for v := 0; v < nvertex; v++ {
					if dualvar[v] < delta {
						delta = dualvar[v]
					}
				}
			}

			// Delta type 2: minimum slack on S-to-free edge.
			for v := 0; v < nvertex; v++ {
				if label[inblossom[v]] == 0 && bestedge[v] != -1 {
					d := slack(bestedge[v])
					if deltatype == -1 || d < delta {
						delta = d
						deltatype = 2
						deltaedge = bestedge[v]
					}
				}
			}

			// Delta type 3: half minimum slack on S-S edge.
			for b := 0; b < 2*nvertex; b++ {
				if blossomparent[b] == -1 && label[b] == 1 && bestedge[b] != -1 {
					d := slack(bestedge[b]) / 2
					if deltatype == -1 || d < delta {
						delta = d
						deltatype = 3
						deltaedge = bestedge[b]
					}
				}
			}

			// Delta type 4: minimum dual of T-blossom.
			for b := nvertex; b < 2*nvertex; b++ {
				if blossombase[b] >= 0 && blossomparent[b] == -1 && label[b] == 2 {
					if deltatype == -1 || dualvar[b] < delta {
						delta = dualvar[b]
						deltatype = 4
						deltablossom = b
					}
				}
			}

			if deltatype == -1 {
				// Max-cardinality: no further improvement possible.
				// Final delta update.
				deltatype = 1
				delta = int64(0)
				for v := 0; v < nvertex; v++ {
					if dualvar[v] < delta || delta == 0 {
						delta = dualvar[v]
					}
				}
				if delta < 0 {
					delta = 0
				}
			}

			// Update dual variables.
			for v := 0; v < nvertex; v++ {
				switch label[inblossom[v]] {
				case 1:
					dualvar[v] -= delta
				case 2:
					dualvar[v] += delta
				}
			}
			for b := nvertex; b < 2*nvertex; b++ {
				if blossombase[b] >= 0 && blossomparent[b] == -1 {
					switch label[b] {
					case 1:
						// top-level S-blossom: z = z + 2*delta
						// Note: dualvar for blossoms stores z directly (not 2*z),
						// unlike vertex dualvar which stores 2*u.
						dualvar[b] += delta
					case 2:
						// top-level T-blossom: z = z - 2*delta
						dualvar[b] -= delta
					}
				}
			}

			// Take action at the point where minimum delta occurred.
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

		// End of stage: expand S-blossoms with zero dual.
		for b := nvertex; b < 2*nvertex; b++ {
			if blossomparent[b] == -1 && blossombase[b] >= 0 &&
				label[b] == 1 && dualvar[b] == 0 {
				expandBlossom(b, true)
			}
		}
	}

	// Convert from endpoint-based mate to vertex-based.
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

// modLen computes i % n with always-positive result.
func modLen(i, n int) int {
	return ((i % n) + n) % n
}

// reverseInts reverses a slice of ints in place.
func reverseInts(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// indexOf returns the index of val in s, or -1 if not found.
func indexOf(s []int, val int) int {
	for i, v := range s {
		if v == val {
			return i
		}
	}
	return -1
}

// rotateSlice returns a new slice with elements rotated so index i becomes 0.
func rotateSlice(s []int, i int) []int {
	n := len(s)
	if n == 0 || i == 0 {
		return s
	}
	result := make([]int, n)
	for idx := range result {
		result[idx] = s[(i+idx)%n]
	}
	return result
}
