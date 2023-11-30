package ion

import (
	"sync"
)

type stateGen[T any] struct {
	f    func() (T, bool)
	mems *Vec[T]
	m    sync.Mutex
}

func (g *stateGen[T]) Elem(i uint64) (T, bool) {
	g.m.Lock()
	defer g.m.Unlock()

	memlen := g.mems.Len()
	if i >= memlen {
		var finished bool
		need := i - memlen + 1
		//fmt.Printf("Taking %d new elements.\n", need)
		newmems := BuildVec(func(add func(e T)) {
			for j := uint64(0); j < need; j++ {
				next, cont := g.f()
				if !cont {
					finished = false
					return
				}
				add(next)
			}
			finished = true
		})
		g.mems = g.mems.Join(newmems)
		if !finished {
			var r T
			return r, false
		}
	}
	// fmt.Printf("g.mems len: %d\n", g.mems.Len())
	// //fmt.Printf("g.mems: %v\n", ToSlice(g.mems))
	// var ii, res int
	// g.mems.Iterate(func(e T) bool {
	// 	var ei any
	// 	ei = e
	// 	res += ei.(int)
	// 	ii++
	// 	if ii == 10000 {
	// 		return false
	// 	}
	// 	return true
	// })
	// fmt.Printf("g.mems sum: %d\n", res)
	return g.mems.Elem(i)
}

func (g *stateGen[T]) Split(n uint64) (Seq[T], Seq[T]) {
	g.Elem(n)
	g.m.Lock()
	defer g.m.Unlock()

	//l, r := g.mems.Split(n)
	l := g.mems.Take(n)
	spl := &splitStateGen[T]{
		g:     g,
		start: n,
	}
	return l, spl
}

func (g *stateGen[T]) Take(n uint64) Seq[T] {
	g.Elem(n)
	g.m.Lock()
	defer g.m.Unlock()

	return g.mems.Take(n)
}

func (g *stateGen[T]) Iterate(f func(T) bool) {
	// TODO: lame only one thread can iterate an immutable struct at a time.
	// need to figure this out.
	g.m.Lock()
	defer g.m.Unlock()
	quit := false
	g.mems.Iterate(func(e T) bool {
		if !f(e) {
			quit = true
			return false
		}
		return true
	})
	if quit {
		return
	}
	for {
		e, cont := g.f()
		if !cont {
			return
		}
		g.mems = g.mems.Append(e)
		if !f(e) {
			return
		}
	}
}

func (g *stateGen[T]) Lazy(f func(func() T) bool) {
	// TODO: lame only one thread can iterate an immutable seq at a time.
	// need to figure this out.
	g.m.Lock()
	defer g.m.Unlock()
	quit := false
	g.mems.Lazy(func(ef func() T) bool {
		if !f(ef) {
			quit = true
			return false
		}
		return true
	})
	if quit {
		return
	}
	for {
		e, cont := g.f()
		if !cont {
			return
		}
		g.mems = g.mems.Append(e)
		if !f(func() T { return e }) {
			return
		}
	}
}

type splitStateGen[T any] struct {
	g     *stateGen[T]
	start uint64
}

func (g *splitStateGen[T]) Elem(i uint64) (T, bool) {
	return g.g.Elem(g.start + i)
}

func (g *splitStateGen[T]) Split(n uint64) (Seq[T], Seq[T]) {
	ll := g.g.Take(g.start + n)
	_, l := ll.Split(g.start)
	return l, &splitStateGen[T]{
		g:     g.g,
		start: g.start + n,
	}

}

func (g *splitStateGen[T]) Take(n uint64) Seq[T] {
	ll := g.g.Take(g.start + n)
	_, l := ll.Split(g.start)
	return l
}

func (g *splitStateGen[T]) Iterate(f func(T) bool) {
	_, og := g.Split(g.start)
	og.Iterate(f)
}

func (g *splitStateGen[T]) Lazy(f func(func() T) bool) {
	_, og := g.Split(g.start)
	og.Lazy(f)
}

// StateGen takes a func `f` and executes it in order to generate
// values of type T in the resulting Seq[T].
//
// The func `f` should be a clojure containing whatever state it
// needs to generate values. Each call to `f` should generate the
// next element in the sequence.
func StateGen[T any](f func() (T, bool)) Seq[T] {
	return &stateGen[T]{
		f: f,
	}
}
