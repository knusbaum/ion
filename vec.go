package ion

import (
	"fmt"
	"io"
	"os"
)

const spanSize = 64

// Vec is a sequence of elements of type T. It is constructed internally
// of spans connected in a somewhat balanced tree structure.
//
// Vec is immutable, meaning operations performed on it return
// new Vecs without modifying the old. Because of the immutable nature
// of the structure, the new Vec shares most of its memory with the
// original, meaning operations can be performed efficiently without
// needing to reconstruct an entirely new vec for every operation.
type Vec[T any] struct {
	leftCount uint64
	height    int8
	l         interface{} // *Vec | *seqLeaf
	r         interface{} // *Vec | *seqLeaf
}

// BuildVec constructs a Vec in an efficient way. It accepts a function,
// `f`, which it executes, passing `f` a function `add`.
// The function `add` can be called repeatedly within the body of `f` in
// order to progressively append elements to the Vec.
//
// For example:
//
//	BuildVec(func(add func(int)) {
//		for i := 0; i < 100; i++ {
//			add(i)
//		}
//	})
//
// This is more efficient than simply Appending to a Vec in a loop.
func BuildVec[T any](f func(add func(T))) *Vec[T] {
	var s *Vec[T]
	f(func(e T) {
		s = s.mutAppend(e)
	})
	return s
}

func (s *Vec[T]) duplicate() *Vec[T] {
	return &Vec[T]{
		leftCount: s.leftCount,
		height:    s.height,
		l:         s.l,
		r:         s.r,
	}
}

func newLeaf[T any]() *seqLeaf[T] {
	l := &seqLeaf[T]{}
	l.seq = l.is[:0]
	return l
}

// Len returns the number of elements in the Vec
func (s *Vec[T]) Len() uint64 {
	if s == nil {
		return 0
	}
	if s.r == nil {
		return s.leftCount
	}

	switch o := s.r.(type) {
	case *Vec[T]:
		return s.leftCount + o.Len()
	case *seqLeaf[T]:
		return s.leftCount + uint64(len(o.seq))
	default:
		panic("BAD TYPE")
	}
}

// Elem implements Seq
func (s *Vec[T]) Elem(idx uint64) (T, bool) {
	// defer func() {
	// 	if p := recover(); p != nil {
	// 		f, _ := os.Create("out.dot")
	// 		s.Dot(f)
	// 		f.Close()
	// 		panic(p)
	// 	}
	// }()
	// fmt.Printf("ELEM %d, LEN: %d\n", idx, s.Len())
	// if idx >= s.Len() {
	// 	panic(fmt.Sprintf("Index %d Out of bounds for Vec of length %d", idx, s.Len()))
	// }
	// //s.Dot(os.Stdout)
	// if s2 := validateVec(s); s2 != nil {
	// 	panic("OH NO!\n")
	// }
	return s.elem(idx)
}

func (s *Vec[T]) elem(idx uint64) (T, bool) {
	if s == nil {
		var r T
		return r, false
		//panic(fmt.Sprintf("Index %d Out of bounds for Vec of length %d", idx, s.Len()))
	}
	if uint64(idx) >= s.leftCount {
		idx -= s.leftCount
		if s.r == nil {
			var r T
			return r, false
			panic("Out of bounds")
		}
		switch o := s.r.(type) {
		case *Vec[T]:
			return o.elem(idx)
		case *seqLeaf[T]:
			if idx >= uint64(len(o.seq)) {
				//s.Dot(os.Stdout)
				var r T
				return r, false
			}
			return o.seq[idx], true
		default:
			panic("Bad Type")
		}
	} else {
		switch o := s.l.(type) {
		case *Vec[T]:
			return o.elem(idx)
		case *seqLeaf[T]:
			return o.seq[idx], true
		default:
			panic("Bad Type")
		}
	}
}

// Iterate implements Seq
func (s *Vec[T]) Iterate(f func(T) bool) {
	s.iterate(f)
}

func (s *Vec[T]) iterate(f func(T) bool) bool {
	if s == nil {
		return true
	}
	if s.l != nil {
		switch o := s.l.(type) {
		case *Vec[T]:
			if !o.iterate(f) {
				return false
			}
		case *seqLeaf[T]:
			for _, e := range o.seq {
				if !f(e) {
					return false
				}
			}
		}
	}
	if s.r != nil {
		switch o := s.r.(type) {
		case *Vec[T]:
			if !o.iterate(f) {
				return false
			}
		case *seqLeaf[T]:
			for _, e := range o.seq {
				if !f(e) {
					return false
				}
			}
		}
	}
	return true
}

// Lazy implements Seq
func (s *Vec[T]) Lazy(f func(func() T) bool) {
	s.lazy(f)
}
func (s *Vec[T]) lazy(f func(func() T) bool) bool {
	if s == nil {
		return true
	}
	if s.l != nil {
		switch o := s.l.(type) {
		case *Vec[T]:
			if !o.lazy(f) {
				return false
			}
		case *seqLeaf[T]:
			for _, e := range o.seq {
				e := e
				if !f(func() T { return e }) {
					return false
				}
			}
		}
	}
	if s.r != nil {
		switch o := s.r.(type) {
		case *Vec[T]:
			if !o.lazy(f) {
				return false
			}
		case *seqLeaf[T]:
			for _, e := range o.seq {
				e := e
				if !f(func() T { return e }) {
					return false
				}
			}
		}
	}
	return true
}

func (s *Vec[T]) mutRebalance() *Vec[T] {
	if bf := s.l.(*Vec[T]).height - s.r.(*Vec[T]).height; bf < -2 {
		// right is taller
		r := s.r.(*Vec[T]).duplicate()
		newLeftCount := s.leftCount + r.leftCount
		s.r = r.l
		r.l = s
		r.leftCount = newLeftCount
		s = r
	} else if bf > 2 {
		// left is taller
		l := s.l.(*Vec[T]).duplicate()
		s.l = l.r
		s.leftCount = s.l.(*Vec[T]).Len()
		l.r = s
		s = l
	}
	return s
}

// Join returns a new Vec[T] that is the contents of the original
// Vec[T] followed by `s2`.
func (s *Vec[T]) Join(s2 *Vec[T]) *Vec[T] {
	if s == nil {
		return s2
	} else if s2 == nil {
		return s
	}

	if bf := s.height - s2.height; bf < -1 {
		//fmt.Printf("Rebalance Right.\n")
		// s2 is taller
		// ignore for now.
		s2 = s2.duplicate()
		s2.l = s.Join(s2.l.(*Vec[T]))
		s2.leftCount = s2.l.(*Vec[T]).Len() // TODO: This is inefficient
		s2 = s2.mutRebalance()
		s2.reheight()
		return s2
	} else if bf > 1 {
		//fmt.Printf("Rebalance Left.\n")
		// s is taller
		s = s.duplicate()
		s.r = s.r.(*Vec[T]).Join(s2)
		s = s.mutRebalance()
		s.reheight()
		return s
	}

	if s.height == 1 && s2.height == 1 {
		// these both have leaf nodes as children. Let's see if we can merge them.
		total := s.Len() + s2.Len()
		if total <= spanSize*2 {
			// These can fit into a single node.

			// TODO: This is terribly inefficient. Just a POC
			var ns *Vec[T]
			s.Iterate(func(i T) bool {
				ns = ns.Append(i)
				return true
			})
			s2.Iterate(func(i T) bool {
				ns = ns.Append(i)
				return true
			})
			return ns
		}
	}

	ns := &Vec[T]{
		leftCount: s.Len(),
		l:         s,
		r:         s2,
	}
	ns.reheight()
	return ns
}

// Split implements Seq
func (s *Vec[T]) Split(idx uint64) (Seq[T], Seq[T]) {
	var ls, rs Seq[T]
	ls, rs = s.split(idx)
	return ls, rs
}

// Take implements Seq
func (s *Vec[T]) Take(idx uint64) Seq[T] {
	// TODO: Don't do the extra work of allocating the rs.
	var ls Seq[T]
	if idx >= s.Len() {
		return s
	}
	ls, _ = s.split(idx)
	return ls
}

func (s *Vec[T]) split(idx uint64) (*Vec[T], *Vec[T]) {
	if s == nil {
		return (*Vec[T])(nil), (*Vec[T])(nil)
		//panic(fmt.Sprintf("Index %d Out of bounds for Vec of length %d", idx, s.Len()))
	}
	if uint64(idx) >= s.leftCount {
		idx -= s.leftCount
		if s.r == nil {
			// TODO: This still allocates in the lower stack frames, when we don't need to
			return s, (*Vec[T])(nil)
			//panic("Out of bounds")
		}
		switch o := s.r.(type) {
		case *Vec[T]:
			s = s.duplicate()
			l, r := o.split(idx)
			s.r = r

			var sl *Vec[T]
			switch o := s.l.(type) {
			case *Vec[T]:
				sl = o
			case *seqLeaf[T]:
				panic("This should not be possible.")
				for _, e := range o.seq {
					fmt.Printf("2Adding %v to pre\n", e)
					sl = sl.Append(e)
				}
			}
			sl = sl.Join(l)

			// l.Iterate(func(i uint64) {
			// 	//fmt.Printf("BEFORE:\n")
			// 	//sl.Dot(os.Stdout)
			// 	fmt.Printf("3Adding %d to pre\n", i)
			// 	sl = sl.Append(i)
			// 	//sl.Dot(os.Stdout)
			// })
			return sl, r
		case *seqLeaf[T]:
			left := s.duplicate()
			if idx == 0 {
				right := s.duplicate()
				right.l = right.r
				right.r = nil
				left.r = nil
				right.leftCount = uint64(len(right.l.(*seqLeaf[T]).seq))
				return left, right
			}

			lo := o.clone()
			if idx >= uint64(len(lo.seq)) {
				return s, nil
			}
			lo.seq = lo.seq[:idx]
			left.r = lo

			ro := o.clone()
			ro.mutCutFront(idx)
			right := &Vec[T]{
				leftCount: uint64(len(ro.seq)),
				height:    1,
				l:         ro,
			}
			return left, right
		default:
			panic("Bad Type")
		}
	} else {
		switch o := s.l.(type) {
		case *Vec[T]:
			s = s.duplicate()
			sl, sr := o.split(idx)
			s.l = sr
			s.leftCount -= idx
			return sl, s
		case *seqLeaf[T]:
			if idx == 0 {
				return nil, s
			}
			right := s.duplicate()
			ro := o.clone()
			ro.mutCutFront(idx)
			right.l = ro
			right.leftCount = uint64(len(ro.seq))

			lo := o.clone()
			lo.seq = lo.seq[:idx]
			left := &Vec[T]{
				leftCount: uint64(len(lo.seq)),
				height:    1,
				l:         lo,
			}
			return left, right
		default:
			fmt.Printf("VAL: %#v\n", s.l)
			s.Dot(os.Stdout)
			panic("Bad Type")
		}
	}
}

// Dot writes out a graphviz dot formatted directed graph to
// the writer `w`. This can be used with graphviz to visualize
// the tree's internal structure.
func (s *Vec[T]) Dot(w io.Writer) {
	fmt.Fprintf(w, "digraph {\n")
	s.dotr(w)
	fmt.Fprintf(w, "}\n")
}

func (s *Vec[T]) dotr(w io.Writer) {
	if s == nil {
		return
	}
	fmt.Fprintf(w, "\t%#p [label=\"left: %d, height: %d\"];\n",
		s, s.leftCount, s.height)
	if s.l != nil {
		switch o := s.l.(type) {
		case *Vec[T]:
			fmt.Fprintf(w, "\t%#p -> %#p;\n", s, o)
			o.dotr(w)
		case *seqLeaf[T]:
			fmt.Fprintf(w, "\t%#p -> %#p;\n", s, o)
			fmt.Fprintf(w, "\t%#p [label=\"%v\"];\n", o, o.seq)
		default:
			panic("BAD TYPE")

		}
	} else {
		fmt.Fprintf(w, "\t%#p -> %#pln;\n", s, s)
		fmt.Fprintf(w, "\t%#pln [label=\"null\"];\n", s)
	}
	if s.r != nil {
		switch o := s.r.(type) {
		case *Vec[T]:
			fmt.Fprintf(w, "\t%#p -> %#p;\n", s, o)
			o.dotr(w)
		case *seqLeaf[T]:
			fmt.Fprintf(w, "\t%#p -> %#p;\n", s, o)
			fmt.Fprintf(w, "\t%#p [label=\"%v\"];\n", o, o.seq)
		default:
			panic("BAD TYPE")

		}
	} else {
		fmt.Fprintf(w, "\t%#p -> %#pln;\n", s, s)
		fmt.Fprintf(w, "\t%#pln [label=\"null\"];\n", s)
	}
}

func (t *Vec[T]) reheight() {
	if t.l != nil && t.r != nil {
		if l, ok := t.l.(*Vec[T]); ok {
			r := t.r.(*Vec[T])
			t.height = max(l.height, r.height) + 1
			return
		}
	}
	t.height = 1
}

// Append returns a new list containing the elements of the original Vec[T]
// with i appended to the end.
func (s *Vec[T]) Append(i T) *Vec[T] {
	//fmt.Printf("Appending %d\n", i)
	if s == nil {
		l := newLeaf[T]()
		l.seq = append(l.seq, i)
		return &Vec[T]{
			leftCount: 1,
			height:    1,
			l:         l,
		}
	}
	if s.r != nil {
		s = s.duplicate()
		switch o := s.r.(type) {
		case *Vec[T]:
			s.r = o.Append(i)
			sl := s.l.(*Vec[T])
			if s.r.(*Vec[T]).height > sl.height {
				sl = sl.duplicate()
				s.l = sl
				s.l = &Vec[T]{
					leftCount: s.leftCount,
					height:    sl.height + 1,
					l:         s.l,
					r:         s.r.(*Vec[T]).l,
				}
				s.leftCount += s.r.(*Vec[T]).leftCount
				sl.reheight()
				s.r = s.r.(*Vec[T]).r
			}
			s.reheight()
		case *seqLeaf[T]:
			if len(o.seq) == spanSize {
				l := newLeaf[T]()
				l.seq = append(l.seq, i)
				s.r = &Vec[T]{
					leftCount: 1,
					height:    1,
					l:         l,
				}
				s.l = &Vec[T]{
					leftCount: s.leftCount,
					height:    1,
					l:         s.l,
					r:         o,
				}
				s.leftCount += spanSize
				s.height = 2
			} else {
				o.seq = append(o.seq, i)
			}
		default:
			panic("BAD TYPE")
		}
		return s
	} else if s.l != nil {
		s = s.duplicate()
		switch o := s.l.(type) {
		case *Vec[T]:
			s.l = o.Append(i)
			s.reheight()
			s.leftCount++
		case *seqLeaf[T]:
			if len(o.seq) == spanSize {
				// s.r must be nil, so we should add a seq to s.r
				l := newLeaf[T]()
				l.seq = append(l.seq, i)
				s.r = l
			} else {
				o = o.clone()
				o.seq = append(o.seq, i)
				s.l = o
				s.leftCount++
			}
		default:
			panic("BAD TYPE")
		}
		return s
	} else {
		s = s.duplicate()
		l := newLeaf[T]()
		l.seq = append(l.seq, i)
		s.l = l
		s.leftCount = 1
		return s
	}
	panic("BAD")
}

func (s *Vec[T]) mutAppend(i T) *Vec[T] {
	if s == nil {
		l := newLeaf[T]()
		l.seq = append(l.seq, i)
		return &Vec[T]{
			leftCount: 1,
			height:    1,
			l:         l,
		}
	}
	if s.r != nil {
		//s = s.duplicate()
		switch o := s.r.(type) {
		case *Vec[T]:
			s.r = o.mutAppend(i)
			sl := s.l.(*Vec[T])
			if s.r.(*Vec[T]).height > sl.height {
				//sl = sl.duplicate()
				//s.l = sl
				s.l = &Vec[T]{
					leftCount: s.leftCount,
					height:    sl.height + 1,
					l:         s.l,
					r:         s.r.(*Vec[T]).l,
				}
				s.leftCount += s.r.(*Vec[T]).leftCount
				sl.reheight()
				s.r = s.r.(*Vec[T]).r
			}
			s.reheight()
		case *seqLeaf[T]:
			if len(o.seq) == spanSize {
				l := newLeaf[T]()
				l.seq = append(l.seq, i)
				s.r = &Vec[T]{
					leftCount: 1,
					height:    1,
					l:         l,
				}
				s.l = &Vec[T]{
					leftCount: s.leftCount,
					height:    1,
					l:         s.l,
					r:         o,
				}
				s.leftCount += spanSize
				s.height = 2
			} else {
				o.seq = append(o.seq, i)
			}
		default:
			panic("BAD TYPE")
		}
		return s
	} else if s.l != nil {
		//s = s.duplicate()
		switch o := s.l.(type) {
		case *Vec[T]:
			s.l = o.mutAppend(i)
			s.reheight()
			s.leftCount++
		case *seqLeaf[T]:
			if len(o.seq) == spanSize {
				// s.r must be nil, so we should add a seq to s.r
				l := newLeaf[T]()
				l.seq = append(l.seq, i)
				s.r = l
			} else {
				//o = o.clone()
				o.seq = append(o.seq, i)
				s.l = o
				s.leftCount++
			}
		default:
			panic("BAD TYPE")
		}
		return s
	} else {
		//s = s.duplicate()
		l := newLeaf[T]()
		l.seq = append(l.seq, i)
		s.l = l
		s.leftCount = 1
		return s
	}
	panic("BAD")
}

type seqLeaf[T any] struct {
	seq []T
	is  [spanSize]T
}

func (s *seqLeaf[T]) clone() *seqLeaf[T] {
	l := &seqLeaf[T]{}
	*l = *s
	l.seq = l.is[:len(s.seq)]
	return l
}

// This function trims a leaf by moving its elements.
// This mutates the VecLeaf so should only be used when
// it is appropriate, e.g. when constructing new values.
func (s *seqLeaf[T]) mutCutFront(idx uint64) {
	s2 := s.seq[idx:]
	s.seq = s.is[:len(s2)]
	copy(s.seq, s2)
}

func validateVec[T any](s *Vec[T]) *Vec[T] {
	if s == nil {
		return nil
	}
	if _, ok := s.l.(*seqLeaf[T]); ok {
		// if s's l is a leaf, s's right must be a leaf.
		if s.r != nil {
			if _, ok := s.r.(*seqLeaf[T]); !ok {
				fmt.Printf("BAD NODE! Left is leaf, right is not leaf.\n")
				return s
			}
		}
	} else {
		if s.r != nil {
			if _, ok := s.r.(*seqLeaf[T]); ok {
				fmt.Printf("BAD NODE! Left is seq, right is leaf.\n")
				return s
			}
		}
	}

	switch o := s.l.(type) {
	case *Vec[uint64]:
		if s.leftCount != o.Len() {
			fmt.Printf("EXPECTED LEFTCOUNT == %d, but was %d\n", o.Len(), s.leftCount)
			return s
		}
	case *seqLeaf[uint64]:
		if s.leftCount != uint64(len(o.seq)) {
			fmt.Printf("(leaf) EXPECTED LEFTCOUNT == %d, but was %d\n", len(o.seq), s.leftCount)
			return s
		}
	}

	if s.l != nil {
		if sl, ok := s.l.(*Vec[T]); ok {
			if bad := validateVec(sl); bad != nil {
				return bad
			}
		}
	}
	if s.r != nil {
		if sr, ok := s.r.(*Vec[T]); ok {
			if bad := validateVec(sr); bad != nil {
				return bad
			}
		}
	}
	return nil
}
