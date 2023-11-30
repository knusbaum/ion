package ion

import (
	"cmp"
	"fmt"
	"io"

	"golang.org/x/exp/constraints"
)

// AVLTree is a tree-based map of keys of type T to values of type U.
//
// AVLTree is immutable, meaning operations performed on it return
// new trees without modifying the old. Because of the immutable nature
// of the structure, the new tree shares most of its memory with the
// original, meaning operations can be performed efficiently without
// needing to reconstruct an entirely new tree for every operation.
type AVLTree[T cmp.Ordered, U any] struct {
	height int8
	k      T
	v      U
	l      *AVLTree[T, U]
	r      *AVLTree[T, U]
}

func (t *AVLTree[T, U]) bf() int8 {
	if t.r == nil {
		if t.l == nil {
			return 0
		} else {
			return 0 - t.l.height
		}
	} else {
		if t.l == nil {
			return t.r.height
		} else {
			return t.r.height - t.l.height
		}
	}
}

func max[T constraints.Ordered](x, y T) T {
	if x > y {
		return x
	}
	return y
}

// Size returns the number of elements present in the tree.
func (t *AVLTree[T, U]) Size() uint64 {
	if t == nil {
		return 0
	}
	return t.l.Size() + t.r.Size() + 1
}

// Insert returns a new tree, consisting of the original tree with the
// key/value pair `k`/`v` added to it.
func (t *AVLTree[T, U]) Insert(k T, v U) *AVLTree[T, U] {
	if t == nil {
		return &AVLTree[T, U]{k: k, v: v, height: 1}
	}

	switch {
	case t.k < k:
		if t.r == nil {
			//t.r = &AVLTree{i: i, height: 1}
			t = &AVLTree[T, U]{
				k:      t.k,
				v:      t.v,
				height: t.height,
				l:      t.l,
				r:      &AVLTree[T, U]{k: k, v: v, height: 1},
			}
			if t.l == nil {
				t.height = 2
			}
		} else {
			tr := t.r.Insert(k, v)
			t = &AVLTree[T, U]{
				k: t.k,
				v: t.v,
				l: t.l,
				r: tr,
			}
			t.reheight()
		}
	case t.k > k:
		if t.l == nil {
			//t.l = &AVLTree{i: i, height: 1}
			t = &AVLTree[T, U]{
				k:      t.k,
				v:      t.v,
				height: t.height,
				r:      t.r,
				l:      &AVLTree[T, U]{k: k, v: v, height: 1},
			}
			if t.r == nil {
				t.height = 2
			}
		} else {
			tl := t.l.Insert(k, v)
			t = &AVLTree[T, U]{
				k: t.k,
				v: t.v,
				l: tl,
				r: t.r,
			}
			t.reheight()
		}
	case t.k == k:
		t := &AVLTree[T, U]{
			k:      k,
			v:      v,
			height: t.height,
			r:      t.r,
			l:      t.l,
		}
		return t
	}

	return t.balance()
}

func (t *AVLTree[T, U]) removeLargest() (largest, tree *AVLTree[T, U]) {
	if t.r != nil {
		t = &AVLTree[T, U]{
			k:      t.k,
			v:      t.v,
			height: t.height,
			l:      t.l,
			r:      t.r,
		}
		l, newright := t.r.removeLargest()
		t.r = newright
		t.reheight()
		//t.balance()
		return l, t.balance()
	}
	if t.l != nil {
		// This is the farthest right node.
		// t = largest, remaining tree is t.l
		t = &AVLTree[T, U]{
			k:      t.k,
			v:      t.v,
			height: t.height,
			l:      t.l,
			r:      t.r,
		}
		ret := t.l
		t.l = nil
		t.height = 1
		return t, ret
	} else {
		// This is the farthest right node.
		// t = largest, remaining tree is nil.
		return t, nil
	}
}

// Delete returns a new tree that does not contain the key `k`, and
// a boolean indicating whether or not an element was removed.
func (t *AVLTree[T, U]) Delete(k T) (*AVLTree[T, U], bool) {
	if t == nil {
		return nil, false
	}

	switch {
	case t.k == k:
		if t.l != nil {
			if t.r != nil {
				// We have two kids.
				// Replace the local node with the largest of the
				// left subtree.
				t = &AVLTree[T, U]{
					k:      t.k,
					v:      t.v,
					height: t.height,
					l:      t.l,
					r:      t.r,
				}
				largest, tree := t.l.removeLargest()
				t.l = tree
				t.k = largest.k
				t.v = largest.v
				t.reheight()
				return t.balance(), true
			} else {
				// We only have a left node. Replace ourselves with  the left node.
				return t.l, true
			}
		} else if t.r != nil {
			// We only have a right node. Replace ourselves with the right node.
			return t.r, true
		} else {
			return nil, true
		}

	case t.k < k:
		if t.r == nil {
			return t, false
		} else {
			if tr, ok := t.r.Delete(k); ok {
				t = &AVLTree[T, U]{
					k:      t.k,
					v:      t.v,
					height: t.height,
					l:      t.l,
					r:      t.r,
				}
				t.r = tr
				t.reheight()
				return t.balance(), true
			}
			return t, false
		}
	case t.k > k:
		if t.l == nil {
			return t, false
		} else {
			if tl, ok := t.l.Delete(k); ok {
				t = &AVLTree[T, U]{
					k:      t.k,
					v:      t.v,
					height: t.height,
					l:      t.l,
					r:      t.r,
				}
				t.l = tl
				t.reheight()
				return t.balance(), true
			}
			return t, false
		}
	}
	return t.balance(), true
}

func (t *AVLTree[T, U]) balance() *AVLTree[T, U] {
	bf := t.bf()
	switch {
	case bf <= -2:
		// left child insert - rebalance
		zbf := t.l.bf()
		if zbf <= 0 {
			// left left
			return t.right_rotate()
		} else if zbf > 0 {
			// left right
			t = &AVLTree[T, U]{
				k:      t.k,
				v:      t.v,
				height: t.height,
				l:      t.l.left_rotate(),
				r:      t.r,
			}
			//t.l = t.l.left_rotate()
			// TODO: right_rotate throws away t, so we can probably
			// avoid allocating t to begin with.
			return t.right_rotate()
		}
	case bf >= 2:
		// right child insert - rebalance
		zbf := t.r.bf()
		if zbf >= 0 {
			// right right
			return t.left_rotate()
		} else if zbf < 0 {
			// right left
			t = &AVLTree[T, U]{
				k:      t.k,
				v:      t.v,
				height: t.height,
				l:      t.l,
				r:      t.r.right_rotate(),
			}
			//t.r = t.r.right_rotate()
			// TODO: left_rotate throws away t, so we can probably
			// avoid allocating t to begin with.
			return t.left_rotate()
		}
	}
	return t
}

func (t *AVLTree[T, U]) getNode(k T) *AVLTree[T, U] {
	if t == nil {
		return nil
	}

	switch {
	case t.k == k:
		return t
	case t.k < k:
		return t.r.getNode(k)
	case t.k > k:
		return t.l.getNode(k)
	}
	panic("Not possible.")
}

// Get looks up the element in the map associated with `k`.
// It also returns a boolean indicating whether the value was found.
func (t *AVLTree[T, U]) Get(k T) (U, bool) {
	if t == nil {
		var r U
		return r, false
	}

	switch {
	case t.k == k:
		return t.v, true
	case t.k < k:
		return t.r.Get(k)
	case t.k > k:
		return t.l.Get(k)
	}
	panic("Not possible.")
}

func (t *AVLTree[T, U]) reheight() {
	if t.l != nil {
		if t.r != nil {
			t.height = max(t.l.height, t.r.height) + 1
		} else {
			t.height = t.l.height + 1
		}
	} else if t.r != nil {
		t.height = t.r.height + 1
	} else {
		t.height = 1
	}
}

func (t *AVLTree[T, U]) right_rotate() *AVLTree[T, U] {
	if t.l == nil {
		// Nothing on the left to
		// rotate to the right.
		return t
	}

	//y := t
	//x := t.l
	y := &AVLTree[T, U]{
		k:      t.k,
		v:      t.v,
		height: t.height,
		l:      t.l,
		r:      t.r,
	}
	x := &AVLTree[T, U]{
		k:      t.l.k,
		v:      t.l.v,
		height: t.l.height,
		l:      t.l.l,
		r:      t.l.r,
	}

	a := t.l.l
	b := t.l.r
	c := t.r

	x.l = a
	x.r = y
	y.l = b
	y.r = c

	if y.l != nil {
		if y.r != nil {
			y.height = max(y.l.height, y.r.height) + 1
		} else {
			y.height = y.l.height + 1
		}
	} else if y.r != nil {
		y.height = y.r.height + 1
	} else {
		y.height = 1
	}

	// We rotated right, so x.r != nil
	if x.l != nil {
		x.height = max(x.l.height, x.r.height) + 1
	} else {
		x.height = x.r.height + 1
	}

	return x
}

func (t *AVLTree[T, U]) left_rotate() *AVLTree[T, U] {
	if t.r == nil {
		// Nothing on the right to
		// rotate to the left.
		return t
	}

	//x := t
	//y := t.r
	x := &AVLTree[T, U]{
		k:      t.k,
		v:      t.v,
		height: t.height,
		l:      t.l,
		r:      t.r,
	}
	y := &AVLTree[T, U]{
		k:      t.r.k,
		v:      t.r.v,
		height: t.r.height,
		l:      t.r.l,
		r:      t.r.r,
	}

	a := t.l
	b := t.r.l
	c := t.r.r

	y.l = x
	y.r = c
	x.l = a
	x.r = b

	// We rotated left, so y now definitely has a left.
	if x.l != nil {
		if x.r != nil {
			x.height = max(x.l.height, x.r.height) + 1
		} else {
			x.height = x.l.height + 1
		}
	} else if x.r != nil {
		x.height = x.r.height + 1
	} else {
		x.height = 1
	}

	if y.r != nil {
		y.height = max(y.l.height, y.r.height) + 1
	} else {
		y.height = y.l.height + 1
	}
	return y
}

// Dot writes out a graphviz dot formatted directed graph to
// the writer `w`. This can be used with graphviz to visualize
// the tree's internal structure.
func (r *AVLTree[T, U]) Dot(w io.Writer) {
	fmt.Fprintf(w, "digraph tree {\n")
	r.rdot(w)
	fmt.Fprintf(w, "}\n")
}

func (r *AVLTree[T, U]) rdot(w io.Writer) {
	if r == nil {
		return
	}

	fmt.Fprintf(w, "\t%v.%d;\n", r.k, r.height)
	if r.l != nil {
		fmt.Fprintf(w, "\t%v.%d -> %v.%d;\n",
			r.k, r.height, r.l.k, r.l.height)
	}
	if r.r != nil {
		fmt.Fprintf(w, "\t%v.%d -> %v.%d;\n",
			r.k, r.height, r.r.k, r.r.height)
	}
	r.l.rdot(w)
	r.r.rdot(w)
}
