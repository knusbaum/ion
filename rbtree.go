package ion

import (
	"cmp"
	"fmt"
	"io"
)

// RBTree is a tree-based map of keys of type T to values of type U.
//
// RBTree is immutable, meaning operations performed on it return
// new trees without modifying the old. Because of the immutable nature
// of the structure, the new tree shares most of its memory with the
// original, meaning operations can be performed efficiently without
// needing to reconstruct an entirely new tree for every operation.
type RBTree[T cmp.Ordered, U any] struct {
	c color
	k T
	v U
	l *RBTree[T, U]
	r *RBTree[T, U]
}

// Insert returns a new tree, consisting of the original tree with the
// key/value pair `k`/`v` added to it.
func (r *RBTree[T, U]) Insert(k T, v U) *RBTree[T, U] {
	//fmt.Printf("INSERT %d\n", i)
	var chain [64]*RBTree[T, U]
	return r.tree_insert(k, v, chain[:0])
}

// Get looks up the element in the map associated with `k`.
// It also returns a boolean indicating whether the value was found.
func (r *RBTree[T, U]) Get(k T) (U, bool) {
	if r == nil {
		var r U
		return r, false
	}
	switch {
	case r.k == k:
		return r.v, true
	case r.k < k:
		return r.r.Get(k)
	case r.k > k:
		return r.l.Get(k)
	}
	panic("Not possible.")
}

// Dot writes out a graphviz dot formatted directed graph to
// the writer `w`. This can be used with graphviz to visualize
// the tree's internal structure.
func (r *RBTree[T, U]) Dot(w io.Writer) {
	fmt.Fprintf(w, "digraph tree {\n")
	r.rdot(w, make(map[T]struct{}))
	fmt.Fprintf(w, "}\n")
}

// Size returns the number of elements present in the tree.
func (r *RBTree[T, U]) Size() uint64 {
	if r == nil {
		return 0
	}
	return r.l.Size() + r.r.Size() + 1
}

func (r *RBTree[T, U]) rdot(w io.Writer, m map[T]struct{}) {
	if r == nil {
		return
	}
	if _, ok := m[r.k]; ok {
		return
	}
	m[r.k] = struct{}{}

	switch r.c {
	case red:
		fmt.Fprintf(w, "\t%v [color=\"red\"]\n", r.k)
	case black:
		fmt.Fprintf(w, "\t%v [color=\"black\"]\n", r.k)
	}
	if r.l != nil {
		fmt.Fprintf(w, "\t%v -> %v [color=\"green\"];\n",
			r.k, r.l.k)
	}
	if r.r != nil {
		fmt.Fprintf(w, "\t%v -> %v [color=\"yellow\"];\n",
			r.k, r.r.k)
	}
	r.l.rdot(w, m)
	r.r.rdot(w, m)
}

func (r *RBTree[T, U]) right_rotate() *RBTree[T, U] {
	if r.l == nil {
		// Nothing on the left to
		// rotate to the right.
		return r
	}

	//y := r
	//x := r.l

	y := &RBTree[T, U]{
		k: r.k,
		v: r.v,
		c: r.c,
		l: r.l,
		r: r.r,
	}
	x := &RBTree[T, U]{
		k: r.l.k,
		v: r.l.v,
		c: r.l.c,
		l: r.l.l,
		r: r.l.r,
	}

	a := r.l.l
	b := r.l.r
	c := r.r

	x.l = a
	x.r = y
	y.l = b
	y.r = c
	return x
}

func (r *RBTree[T, U]) left_rotate() *RBTree[T, U] {
	if r.r == nil {
		// Nothing on the right to
		// rotate to the left.
		return r
	}

	//x := r
	//y := r.r

	x := &RBTree[T, U]{
		k: r.k,
		v: r.v,
		c: r.c,
		l: r.l,
		r: r.r,
	}
	y := &RBTree[T, U]{
		k: r.r.k,
		v: r.r.v,
		c: r.r.c,
		l: r.r.l,
		r: r.r.r,
	}

	a := r.l
	b := r.r.l
	c := r.r.r

	y.l = x
	y.r = c
	x.l = a
	x.r = b
	return y
}

type color uint8

const (
	red color = iota
	black
)

func root[T cmp.Ordered, U any](n, p, g *RBTree[T, U], chain []*RBTree[T, U]) *RBTree[T, U] {
	if len(chain) > 0 {
		return chain[0]
	} else if g != nil {
		return g
	} else if p != nil {
		return p
	}
	return n
}

func npg[T cmp.Ordered, U any](chain []*RBTree[T, U]) (n, p, g *RBTree[T, U], c []*RBTree[T, U]) {
	c = chain
	if len(c) > 0 {
		n = c[len(c)-1]
		c = c[:len(c)-1]
	}
	if len(c) > 0 {
		p = c[len(c)-1]
		c = c[:len(c)-1]
	}
	if len(c) > 0 {
		g = c[len(c)-1]
		c = c[:len(c)-1]
	}
	return
}

func next[T cmp.Ordered, U any](c []*RBTree[T, U]) (*RBTree[T, U], []*RBTree[T, U]) {
	if len(c) > 0 {
		return c[len(c)-1], c[:len(c)-1]
	}
	return nil, c
}

func rebalance[T cmp.Ordered, U any](chain []*RBTree[T, U]) *RBTree[T, U] {
	n, p, g, chain := npg(chain)
	for {
		n.c = red
		if p == nil {
			//fmt.Printf("CASE i3\n")
			// Case i3
			// Tree is already balanced.
			return n
		}
		if p.c == black {
			//fmt.Printf("CASE i1\n")
			// Case i1
			// Tree is already balanced.
			return root(n, p, g, chain)
		}
		// else parent is red
		if g == nil {
			//fmt.Printf("CASE i4\n")
			// Case i4
			p.c = black
			return p
		}
		var u *RBTree[T, U]
		if g.l == p {
			// Parent is left branch
			u = g.r
			if u != nil {
				u = &RBTree[T, U]{
					k: u.k,
					v: u.v,
					c: u.c,
					l: u.l,
					r: u.r,
				}
				g.r = u
			}
		} else {
			u = g.l
			if u != nil {
				u = &RBTree[T, U]{
					k: u.k,
					v: u.v,
					c: u.c,
					l: u.l,
					r: u.r,
				}
				g.l = u
			}
		}

		if u != nil && u.c == red {
			//fmt.Printf("CASE i2\n")
			// Case i2
			p.c = black
			u.c = black
			g.c = red
			n = g
			p, chain = next(chain)
			g, chain = next(chain)
			continue
		}

		var inner bool
		// Uncle is black.
		// Need to determine i5 vs i6
		if g.l == p {
			if p.l == n {
				inner = false
			} else {
				inner = true
			}
		} else {
			if p.r == n {
				inner = false
			} else {
				inner = true
			}
		}

		if inner {
			//fmt.Printf("CASE i5\n")
			// Case i5
			// Case i5 does a rotation to turn it into case i6
			if g.l == p {
				// parent is on the left. Child on right. Rotate parent left
				p = p.left_rotate()
				g.l = p
				n = p.l
			} else {
				// parent is on the right. Child on left. Rotate parent right
				p = p.right_rotate()
				g.r = p
				n = p.r
			}
		}

		//fmt.Printf("CASE i6\n")
		// Case i6
		// need the great-grandparent
		gg, chain := next(chain)
		var left bool
		if gg != nil && gg.l == g {
			left = true
		}
		p.c = black
		g.c = red
		if g.l == p {
			// rotate right
			g = g.right_rotate()
		} else {
			// rotate left
			g = g.left_rotate()
		}

		if gg != nil {
			if left {
				gg.l = g
			} else {
				gg.r = g
			}
		}
		// Tree is balanced. Return the root.
		return root(p, g, gg, chain)
	}
}

func rechain[T cmp.Ordered, U any](chain []*RBTree[T, U]) []*RBTree[T, U] {
	nc := make([]*RBTree[T, U], len(chain))
	prev := &RBTree[T, U]{
		k: chain[0].k,
		v: chain[0].v,
		c: chain[0].c,
		l: chain[0].l,
		r: chain[0].r,
	}
	nc[0] = prev
	for i := 1; i < len(chain); i++ {
		t := chain[i]

		nt := &RBTree[T, U]{
			k: t.k,
			v: t.v,
			c: t.c,
			l: t.l,
			r: t.r,
		}
		if prev.l == t {
			prev.l = nt
		} else if prev.r == t {
			prev.r = nt
		} else {
			// fmt.Printf("CHAIN:\n")
			// chain[0].Dot(os.Stdout)
			// fmt.Printf("EXPECTED %d to be left or right of %d\n",
			// 	t.i, prev.i)
			panic("NOT FOUND")
		}
		nc[i] = nt
		prev = nt
	}
	return nc
}

func (r *RBTree[T, U]) tree_insert(k T, v U, chain []*RBTree[T, U]) *RBTree[T, U] {
	if r == nil {
		return &RBTree[T, U]{k: k, v: v}
	}
	chain = append(chain, r)
	switch {
	case r.k == k:
		r := duplicate(r)
		r.v = v
		return r
	case r.k < k:
		if r.r == nil {
			chain = rechain(chain)
			r = chain[len(chain)-1]
			r.r = &RBTree[T, U]{k: k, v: v}
			return rebalance(append(chain, r.r))
		}
		return r.r.tree_insert(k, v, chain)

	case r.k > k:
		if r.l == nil {
			chain = rechain(chain)
			r = chain[len(chain)-1]
			r.l = &RBTree[T, U]{k: k, v: v}
			return rebalance(append(chain, r.l))
		}
		return r.l.tree_insert(k, v, chain)
	}
	panic("Not possible.")
}

// Delete returns a new tree that does not contain the key `k`, and
// a boolean indicating whether or not an element was removed.
func (r *RBTree[T, U]) Delete(k T) (*RBTree[T, U], bool) {
	//fmt.Printf("DELETE %d\n", i)
	var chain [64]*RBTree[T, U]
	return r.tree_delete(k, chain[:0])
}

func duplicate[T cmp.Ordered, U any](t *RBTree[T, U]) *RBTree[T, U] {
	if t == nil {
		return nil
	}
	return &RBTree[T, U]{
		k: t.k,
		v: t.v,
		c: t.c,
		l: t.l,
		r: t.r,
	}
}

func rebalance_del[T cmp.Ordered, U any](chain []*RBTree[T, U]) *RBTree[T, U] {
	n, chain := next(chain)
	p, chain := next(chain)

	if p == nil {
		return nil
	}

	var left bool
	if p.l == n {
		left = true
		p.l = nil
	} else if p.r == n {
		left = false
		p.r = nil
	} else {
		panic("1 Wrong parent")
	}

	var s, d, c *RBTree[T, U]
	for {
		// Everything in the chain has already
		// been duplicated, but the siblings and
		// nephews have not been yet. We'll duplicate
		// them here.
		if left {
			s = duplicate(p.r)
			p.r = s
			d = duplicate(s.r)
			s.r = d
			c = duplicate(s.l)
			s.l = c
		} else {
			s = duplicate(p.l)
			p.l = s
			d = duplicate(s.l)
			s.l = d
			c = duplicate(s.r)
			s.r = c
		}
		if s.c == red {
			goto d3
		}
		if d != nil && d.c == red {
			goto d6
		}
		if c != nil && c.c == red {
			goto d5
		}
		if p.c == red {
			goto d4
		}

		// case D2 p, c, s, d all black
		//fmt.Printf("Case d2\n")
		s.c = red
		n = p
		// Not in the original
		p, chain = next(chain)
		if p == nil {
			//fmt.Printf("Case d1\n")
			return n
		}
		if p.l == n {
			left = true
		} else if p.r == n {
			left = false
		} else {
			panic("3 AAAAHHHH")
		}
	}

d3:
	//fmt.Printf("Case d3\n")
	{
		var np, oldp *RBTree[T, U]
		if left {
			np = p.left_rotate()
			oldp = p
			p = np.l

			p.c = red
			np.c = black //s.c = black
			s = c
			// C has been duplicated, but c's kids haven't.
			if s != nil {
				d = duplicate(s.r)
				s.r = d
				c = duplicate(s.l)
				s.l = c
			} else {
				d = nil
				c = nil
			}
		} else {
			np = p.right_rotate()
			oldp = p
			p = np.r
			p.c = red
			np.c = black //s.c = black
			s = c
			if s != nil {
				d = duplicate(s.l)
				s.l = d
				c = duplicate(s.r)
				s.r = c
			} else {
				d = nil
				c = nil
			}
		}
		if len(chain) > 0 {
			g := chain[len(chain)-1]
			if g.l == oldp {
				g.l = np
			} else if g.r == oldp {
				g.r = np
			} else {
				panic("20 Fail")
			}
		}
		// p used to be the root. Now np is the root.
		// Need to push np into the chain.
		chain = append(chain, np)

		if d != nil && d.c == red {
			goto d6
		}
		if c != nil && c.c == red {
			goto d5
		}
	}

d4:
	//fmt.Printf("Case d4\n")
	if s != nil {
		s.c = red
	}
	p.c = black
	if len(chain) > 0 {
		return chain[0]
	}
	return p

d5:
	//fmt.Printf("Case d5\n")
	if left {
		ns := s.right_rotate()
		p.r = ns
		s = ns.r
	} else {
		ns := s.left_rotate()
		p.l = ns
		s = ns.l
	}
	s.c = red
	c.c = black
	d = s
	s = c
d6:
	//fmt.Printf("Case d6\n")
	var np, oldp *RBTree[T, U]
	if left {
		np = p.left_rotate()
		oldp = p
		p = np.l
		s = np
	} else {
		np = p.right_rotate()
		oldp = p
		p = np.r
		s = np
	}
	if len(chain) > 0 {
		g := chain[len(chain)-1]
		if g.l == oldp {
			//fmt.Printf("2 %d left -> %d\n", g.i, np.i)
			g.l = np
		} else if g.r == oldp {
			g.r = np
		} else {
			fmt.Printf("expected %v to gave child %v but it does not.\n", g.k, p.k)
			panic("2 Fail")
		}
	}
	s.c = p.c
	p.c = black
	d.c = black
	if len(chain) > 0 {
		return chain[0]
	}
	return np
}

func replace[T cmp.Ordered, U any](t *RBTree[T, U], i, j T) {
	switch {
	case t.k == i:
		t.k = j
	case t.k < i:
		replace(t.r, i, j)
	case t.k > i:
		replace(t.l, i, j)
	}
	return
}

func (r *RBTree[T, U]) tree_delete(k T, chain []*RBTree[T, U]) (*RBTree[T, U], bool) {
	if r == nil {
		return nil, false
	}
	chain = append(chain, r)
	switch {
	case r.k == k:
		if r.l != nil && r.r != nil {
			// can swap our value with our in-order predecessor
			predec, newTree := r.l.removeLargest(chain)
			// It is safe to modify newTree because the node containing r.i
			// will be new within newTree.
			// This can be optimized in the future so we don't have to
			// re-traverse the tree to find i.
			replace(newTree, r.k, predec.k)
			return newTree, true
		} else if r.l != nil {
			// Only left child
			// replace this node with it's child and color it black.
			chain = rechain(chain)
			r = chain[len(chain)-1]
			r.l = &RBTree[T, U]{
				k: r.l.k,
				v: r.l.v,
				c: r.l.c,
				l: r.l.l,
				r: r.l.r,
			}
			r.l.c = black
			if len(chain) > 1 {
				p := chain[len(chain)-2]
				if p.l == r {
					p.l = r.l
				} else if p.r == r {
					p.r = r.l
				} else {
					panic("foo")
				}
				return chain[0], true
			}
			return r.l, true
		} else if r.r != nil {
			// Only left child
			// replace this node with its child and color it black.
			chain = rechain(chain)
			r = chain[len(chain)-1]
			r.l = &RBTree[T, U]{
				k: r.r.k,
				v: r.r.v,
				c: r.r.c,
				l: r.r.l,
				r: r.r.r,
			}
			r.r.c = black
			if len(chain) > 1 {
				p := chain[len(chain)-2]
				if p.l == r {
					p.l = r.r
				} else if p.r == r {
					p.r = r.r
				} else {
					panic("foo")
				}
				return chain[0], true
			}
			return r.r, true
		} else {
			// No children
			if len(chain) == 1 {
				// No children and we are the root.
				return nil, true
			}
			chain = rechain(chain)
			r = chain[len(chain)-1]
			p := chain[len(chain)-2]
			if r.c == red {
				if p.l == r {
					p.l = nil
				} else if p.r == r {
					p.r = nil
				} else {
					panic("foo2")
				}
				// No children and we are a red node.
				return chain[0], true
			}
			// No children and we are a black node. This will create an imbalance
			// and we need to fix the tree.
			//return rebalance_del(chain), true
			rb := rebalance_del(chain)
			return rb, true
		}
	case r.k > k:
		if nl, ok := r.l.tree_delete(k, chain); ok {
			return nl, true
		}
		return chain[0], false
	case r.k < k:
		if nr, ok := r.r.tree_delete(k, chain); ok {
			return nr, true
		}
		return chain[0], false
	}
	panic("not possible.")
}

func (t *RBTree[T, U]) removeLargest(chain []*RBTree[T, U]) (largest, tree *RBTree[T, U]) {
	if t.r != nil {
		return t.r.removeLargest(append(chain, t))
	}

	// We found the rightmost. We will delete() the current node and return the tree.
	d, _ := t.tree_delete(t.k, chain)
	return t, d

}
