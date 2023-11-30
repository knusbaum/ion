package ion

import (
	"math"

	"golang.org/x/exp/constraints"
)

// A Seq is a possibly unbounded sequence of elements of type T.
// Seqs are immutable, so any operation modifying a Seq
// (such as Split) must leave the original Seq intact and return
// two new Seqs.
//
// Many operations on Seqs are lazy in nature, such as mapping
// and filtering, meaning it's possible (and useful) to map
// and filter unbounded sequences.
type Seq[T any] interface {
	// Elem returns the element at index i.
	// If the sequence does not contain i elements, it will return
	// the zero value of T and false.
	Elem(i uint64) (T, bool)

	// Split splits a sequence after n elements, returning a Seq
	// containing the first n elements and one containing the
	// remainder of the original Seq. Split must not modify the
	// original Seq. If the seq contains < n elements, the
	// first returned sequence will contain all elements from
	// the original sequence and the second returned sequence
	// will be empty.
	//
	// Note: the first returned sequence may contain < n elements,
	// if the original sequence contained < n elements.
	Split(n uint64) (Seq[T], Seq[T])

	// Take returns a Seq containing the first n elements of the
	// original Seq. The original Seq is not modified.
	// Take can be more efficient than Split since there are
	// more scenarios where evaluation of the Seq elements
	// can be delayed.
	Take(n uint64) Seq[T]

	// Iterate executes a function over every element of the Seq,
	// until the Seq ends or the function returns false.
	Iterate(func(T) bool)

	// Lazy executes a function over every element of the Seq,
	// passing the elements as thunks which will return the
	// element.
	//
	// This is useful to delay the execution of computations
	// such as maps until the execution of the thunk. For instance,
	// this can be used to distribute work over a set of goroutines
	// and have the goroutines themselves incur the cost of mapping
	// the elements is parallel, rather than having the routine
	// executing Lazy incuring the cost as is the case with Iterate.
	Lazy(func(func() T) bool)
}

type mappedSeq[T, U any] struct {
	s Seq[T]
	f func(T) U
}

func (m *mappedSeq[T, U]) Elem(i uint64) (U, bool) {
	if e, ok := m.s.Elem(i); ok {
		return m.f(e), true
	}
	var r U
	return r, false
}

func (m *mappedSeq[T, U]) Split(n uint64) (Seq[U], Seq[U]) {
	sl, sr := m.s.Split(n)
	l := &mappedSeq[T, U]{
		s: sl,
		f: m.f,
	}
	r := &mappedSeq[T, U]{
		s: sr,
		f: m.f,
	}
	return l, r
}

func (m *mappedSeq[T, U]) Take(n uint64) Seq[U] {
	l := &mappedSeq[T, U]{
		s: m.s.Take(n),
		f: m.f,
	}
	return l
}

func (m *mappedSeq[T, U]) Iterate(f func(U) bool) {
	m.s.Iterate(func(e T) bool {
		return f(m.f(e))
	})
}

func (m *mappedSeq[T, U]) Lazy(f func(func() U) bool) {
	m.s.Lazy(func(e func() T) bool {
		return f(func() U {
			return m.f(e())
		})
	})
}

// Map takes a Seq[T] `s` and a func `f` which will be executed
// on every element of `s`, returning a new value of type U.
// It returns a new Seq[U] containing the results of the map.
//
// The mapping is executed lazily, meaning it is safe and useful
// to Map over unbounded sequences.
//
// For example:
//
//	// Create an unbounded list of integers.
//	n := From[int](0,1)
//	// Map the integers by adding 1 and converting to float64, into an unbounded Seq[float64].
//	m := Map(n, func(i int) float64 { return float64(i) + 1 })
//
// Values resulting from the application of `f` to the underlying Seq
// are not retained, and `f` may be called multiple times on a given element.
// This being the case, `f` should ideally be idempotent and stateless.
// If it is not, it may not be safe to Iterate the resulting Seq more than once.
//
// Notes about State:
//
// Although it is ideal to use idempotent functions, it is often very useful to
// iterate stateful functions over stateful values, for instance Map'ing a handler
// function over a sequence of connections (say, net.Conn). In these cases, it is
// important not to "realize" elements of the Seq more than once. The simplest
// and fool-proof way to do this is to never apply more than one operation to
// a seq. e.g. this is ok:
//
//	var conns Seq[net.Conn]
//	conns = listen()
//	n := Map(conns, handle)
//	e := Fold(n, handleErrors)
//
// But this bay be problematic, because we are applying multiple operations to n:
//
//	var conns Seq[net.Conn]
//	conns = listen()
//	n := Map(conns, handle)
//	e := Fold(n, handleErrors)
//	other := n.Iterate(func(e SomeType) {
//	    ...
//	})
//
// The Memo function may be useful in ensuring Map functions are never applied more
// than once to the elements of their underlying Seqs, but keep in mind this means
// the values of these operations are retained in memory.
func Map[T, U any](s Seq[T], f func(T) U) Seq[U] {
	return &mappedSeq[T, U]{
		s: s,
		f: f,
	}
}

type repseq[T any] struct {
	e     T
	limit uint64
}

func (r *repseq[T]) Elem(i uint64) (T, bool) {
	return r.e, true
}

func (r *repseq[T]) Split(n uint64) (Seq[T], Seq[T]) {
	s := &repseq[T]{
		e:     r.e,
		limit: n,
	}

	if r.limit > 0 {
		if n == r.limit {
			return s, (*Vec[T])(nil)
		}
		return s, &repseq[T]{
			e:     r.e,
			limit: r.limit - n,
		}
	}

	return s, r
}

func (r *repseq[T]) Take(n uint64) Seq[T] {
	return &repseq[T]{
		e:     r.e,
		limit: n,
	}
}

func (r *repseq[T]) Iterate(f func(T) bool) {
	if r.limit > 0 {
		for i := uint64(0); i < r.limit; i++ {
			if !f(r.e) {
				return
			}
		}
		return
	} else {
		for {
			if !f(r.e) {
				return
			}
		}
	}
}

func (r *repseq[T]) Lazy(f func(func() T) bool) {
	if r.limit > 0 {
		for i := uint64(0); i < r.limit; i++ {
			if !f(func() T { return r.e }) {
				return
			}
		}
		return
	} else {
		for {
			if !f(func() T { return r.e }) {
				return
			}
		}
	}

	for {
		if !f(func() T { return r.e }) {
			return
		}
	}
}

// Repeatedly returns an unbounded Seq[T] containing e.
//
// Note, e is copied, so it is wise to use non-pointer or
// immutable values.
func Repeatedly[T any](e T) Seq[T] {
	return &repseq[T]{
		e: e,
	}
}

type Number interface {
	constraints.Integer | constraints.Float
}

type genseq[T Number] struct {
	start T
	by    T
	limit uint64
}

func (r *genseq[T]) Elem(i uint64) (T, bool) {
	if r.limit > 0 && i >= r.limit {
		var ret T
		return ret, false
	}
	return r.start + T(i)*r.by, true
}

func (r *genseq[T]) Split(n uint64) (Seq[T], Seq[T]) {
	if r.limit > 0 && n > r.limit {
		return r, (*Vec[T])(nil)
	}
	s := &genseq[T]{
		start: r.start,
		by:    r.by,
		limit: n,
	}

	e, ok := r.Elem(n)
	if !ok {
		return r, (*Vec[T])(nil)
	}
	nr := &genseq[T]{
		start: e,
		by:    r.by,
		limit: r.limit - n,
	}
	return s, nr
}

func (r *genseq[T]) Take(n uint64) Seq[T] {
	lim := n
	if r.limit > 0 && r.limit < n {
		lim = r.limit
	}
	return &genseq[T]{
		start: r.start,
		by:    r.by,
		limit: lim,
	}
}

func (r *genseq[T]) Iterate(f func(T) bool) {
	limit := uint64(math.MaxUint64)
	if r.limit > 0 {
		limit = r.limit
	}
	for i := uint64(0); i < limit; i++ {
		e, _ := r.Elem(i)
		if !f(e) {
			return
		}
	}
}

func (r *genseq[T]) Lazy(f func(func() T) bool) {
	limit := uint64(math.MaxUint64)
	if r.limit > 0 {
		limit = r.limit
	}
	for i := uint64(0); i < limit; i++ {
		j := i
		cont := f(func() T {
			e, _ := r.Elem(j)
			return e
		})
		if !cont {
			return
		}
	}
}

// From creates an unbounded Seq[T] of numeric values (see Number)
// starting at start and increasing by `by`.
func From[T Number](start, by T) Seq[T] {
	return &genseq[T]{
		start: start,
		by:    by,
	}
}

type generateSeq[T, U any] struct {
	f     func(state U) (T, U, bool)
	state U
	limit uint64
}

// Generate takes a func `f` and executes it in order to generate
// values of type T in the resulting Seq[T].
//
// The func `f` takes a state of any type, and should generate a
// value based on that state. `f` should be idempotent, as it may
// be executed multiple times on the same state. The func `f`
// must return a value of type T, and the next state of type U.
func Generate[T, U any](f func(state U) (T, U, bool)) Seq[T] {
	return &generateSeq[T, U]{
		f: f,
	}
}

func GenerateInit[T, U any](state U, f func(state U) (T, U, bool)) Seq[T] {
	return &generateSeq[T, U]{
		f:     f,
		state: state,
	}
}

func (g *generateSeq[T, U]) Elem(i uint64) (T, bool) {
	if g.limit > 0 && i >= g.limit {
		var ret T
		return ret, false
	}
	var res T
	state := g.state
	for j := uint64(0); j <= i; j++ {
		var cont bool
		res, state, cont = g.f(state)
		if !cont {
			return res, false
		}
	}
	return res, true
}

func (g *generateSeq[T, U]) Split(n uint64) (Seq[T], Seq[T]) {
	if g.limit > 0 && n > g.limit {
		return g, (*Vec[T])(nil)
	}

	state := g.state
	l := BuildVec(func(add func(T)) {
		for i := uint64(0); i < n; i++ {
			var e T
			var cont bool
			e, state, cont = g.f(state)
			if !cont {
				return
			}
			add(e)
		}
	})

	r := &generateSeq[T, U]{
		f:     g.f,
		state: state,
		limit: g.limit - n,
	}
	return l, r
}

func (g *generateSeq[T, U]) Take(n uint64) Seq[T] {
	lim := n
	if g.limit > 0 && g.limit < n {
		lim = g.limit
	}

	return &generateSeq[T, U]{
		f:     g.f,
		state: g.state,
		limit: lim,
	}
}

func (g *generateSeq[T, U]) Iterate(f func(T) bool) {
	state := g.state
	if g.limit > 0 {
		var i uint64
		for {
			var e T
			var cont bool
			if i == g.limit {
				return
			}
			i++
			e, state, cont = g.f(state)
			if !cont {
				return
			}
			if !f(e) {
				return
			}
		}
	} else {
		for {
			var e T
			var cont bool
			e, state, cont = g.f(state)
			if !cont {
				return
			}
			if !f(e) {
				return
			}
		}
	}
}

func (g *generateSeq[T, U]) Lazy(f func(func() T) bool) {
	// unfortunately this cannot be lazy, because Elem(i) depends on
	// elem i-1, and we cannot guarantee the execution order of the
	// thunks we return. Instead, we evaluate the current element and
	// return a closure that returns it.
	state := g.state
	if g.limit > 0 {
		var i uint64
		for {
			var e T
			var cont bool
			if i == g.limit {
				return
			}
			i++
			e, state, cont = g.f(state)
			if !cont {
				return
			}
			if !f(func() T { return e }) {
				return
			}
		}
	} else {
		for {
			var e T
			var cont bool
			e, state, cont = g.f(state)
			if !cont {
				return
			}
			if !f(func() T { return e }) {
				return
			}
		}
	}
}

// Fold folds a Seq[T] `s` into a value of type U, based on  the function `f`.
// The function `f` is run over each element of `s`. It accepts a value of
// type U, which is the current value (accumulator) for the fold, and a
// value of type T, which is the current element of `s`. The function `f`
// must return the new accumulator value of type U. Fold returns the final
// accumulator value after `f` has been run over every element of `s`.
//
// Note: Running a fold on an unbounded sequence will never terminate.
// One should usually use Split() or Take() on unbounded sequences first
// to limit the output.
//
// For example, to sum the first 1000 primes (with imaginary isPrime and sum
// functions):
//
//	n := From[int](1,1)
//	n = Filter(n, isPrime)
//	n = n.Take(1000)
//	result := Fold(n, sum)
//
// Or more succinctly:
//
//	result := Fold(Filter(From[int](1,1), isPrime).Take(1000), sum)
func Fold[T, U any](s Seq[T], f func(U, T) U) U {
	var u U
	var i uint64
	s.Iterate(func(e T) bool {
		i++
		u = f(u, e)
		return true
	})
	return u
}

type filterSeq[T any] struct {
	s     Seq[T]
	f     func(T) bool
	limit uint64
}

func (f *filterSeq[T]) Elem(i uint64) (T, bool) {
	if f.limit > 0 && i >= f.limit {
		var ret T
		return ret, false
	}
	var res T
	var ec uint64
	var found bool
	f.s.Iterate(func(e T) bool {
		if f.f(e) {
			if ec == i {
				res = e
				found = true
				return false
			}
			ec++
		}
		return true
	})
	return res, found
}

func (f *filterSeq[T]) Split(n uint64) (Seq[T], Seq[T]) {
	if f.limit > 0 && n > f.limit {
		return f, (*Vec[T])(nil)
	}

	var i uint64
	var split uint64
	l := BuildVec(func(add func(T)) {
		f.s.Iterate(func(e T) bool {
			if i == n {
				return false
			}
			split++
			if f.f(e) {
				add(e)
				i++
			}
			return true
		})
	})
	_, rr := f.s.Split(split)
	r := &filterSeq[T]{
		s:     rr,
		f:     f.f,
		limit: f.limit - n,
	}
	return l, r
}

func (f *filterSeq[T]) Take(n uint64) Seq[T] {
	lim := n
	if f.limit > 0 && f.limit < n {
		lim = f.limit
	}
	return &filterSeq[T]{
		s:     f.s,
		f:     f.f,
		limit: lim,
	}
}

func (f *filterSeq[T]) Iterate(fn func(T) bool) {
	if f.limit > 0 {
		var i uint64
		f.s.Iterate(func(e T) bool {
			if i == f.limit {
				return false
			}
			if f.f(e) {
				i++
				return fn(e)
			}
			return true
		})
	} else {
		f.s.Iterate(func(e T) bool {
			if f.f(e) {
				return fn(e)
			}
			return true
		})
	}
}

func (f *filterSeq[T]) Lazy(fn func(func() T) bool) {
	if f.limit > 0 {
		var i uint64
		f.s.Lazy(func(e func() T) bool {
			if i == f.limit {
				return false
			}
			el := e()
			if f.f(el) {
				i++
				return fn(func() T { return el })
			}
			return true
		})
	} else {
		f.s.Lazy(func(e func() T) bool {
			el := e()
			if f.f(el) {
				return fn(func() T { return el })
			}
			return true
		})
	}
}

// Filter takes a Seq[T] 's' and returns a new Seq[T] which contains only
// the elements for which the func `f` returns true. The func `f` should
// be idempotent, as it may be called multiple times on the same element.
func Filter[T any](s Seq[T], f func(T) bool) Seq[T] {
	return &filterSeq[T]{
		s: s,
		f: f,
	}
}

// ToSlice converts a Seq[T] into a []T.
//
// Note: Running ToSlice on an unbounded sequence will never terminate.
// One should usually use Split() or Take() on unbounded sequences first
// to limit the output.
func ToSlice[T any](s Seq[T]) []T {
	var sl []T
	s.Iterate(func(e T) bool {
		sl = append(sl, e)
		return true
	})
	return sl
}

// Always takes a function accepting an element T which returns nothing,
// and returns a function that does the same thing but always returns true.
//
// This is useful for calls to Iterate:
//
//	seq.Iterate(func(e int) {
//	    f(e)
//	    return true
//	})
//
//	becomes:
//	seq.Iterate(Always(f))
func Always[T any](f func(e T)) func(e T) bool {
	return func(e T) bool {
		f(e)
		return true
	}
}
