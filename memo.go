package ion

import (
	"math"
	"sync"
)

type memo[T any] struct {
	s    Seq[T]
	mems *Vec[T]
	m    sync.Mutex
}

func (m *memo[T]) Elem(i uint64) (T, bool) {
	m.m.Lock()
	defer m.m.Unlock()
	len := m.mems.Len()
	if i >= len {
		need := i - len + 1
		t, next := m.s.Split(uint64(need))
		m.s = next
		news := BuildVec(func(add func(e T)) {
			t.Iterate(func(e T) bool {
				add(e)
				return true
			})
		})
		m.mems = m.mems.Join(news)
	}
	return m.mems.Elem(i)
}

func (m *memo[T]) Split(n uint64) (Seq[T], Seq[T]) {
	left := &memoPart[T]{
		underlying: m,
		lower:      0,
		upper:      n,
	}
	right := &memoPart[T]{
		underlying: m,
		lower:      n,
		upper:      math.MaxUint64,
	}
	return left, right

}

func (m *memo[T]) Take(n uint64) Seq[T] {
	return &memoPart[T]{
		underlying: m,
		lower:      0,
		upper:      n,
	}
}

func (m *memo[T]) Iterate(f func(T) bool) {
	for i := uint64(0); ; i++ { // TODO: This will only iterate up to math.MaxUint64 elements.
		e, ok := m.Elem(i)
		if !ok || !f(e) {
			return
		}
	}
}

func (m *memo[T]) Lazy(f func(func() T) bool) {
	quit := false
	m.mems.Lazy(func(ef func() T) bool {
		if !f(ef) {
			quit = true
			return false
		}
		return true
	})
	if quit {
		return
	}
	m.s.Lazy(f) // TODO: This is a problem, because we're not memoizing the results.
}

var _ Seq[int] = &memoPart[int]{}

// memoPart is a part of a memo object, used to avoid the need
// to realize memo elements when doing operations on a memo
// such as Take and Split.
//
// memoPart just keeps offsets into the memo Seq and forwards
// operations to the underlying memo instance.
type memoPart[T any] struct {
	underlying *memo[T]
	lower      uint64
	upper      uint64
}

func (m *memoPart[T]) Elem(n uint64) (T, bool) {
	if n >= (m.upper - m.lower) {
		var ret T
		return ret, false
	}
	return m.underlying.Elem(n + m.lower)
}

func (m *memoPart[T]) Iterate(f func(T) bool) {
	// TODO: This will only iterate up to math.MaxUint64 elements.
	for i := m.lower; i < m.upper; i++ {
		e, ok := m.underlying.Elem(i)
		if !ok || !f(e) {
			return
		}
	}
}

func (m *memoPart[T]) Lazy(f func(func() T) bool) {
	for i := m.lower; i < m.upper; i++ {
		j := i
		cont := f(func() T {
			e, ok := m.underlying.Elem(j)
			if !ok {
				// TODO: This shouldn't happen, but indicates
				// something wrong in the API.
				var ret T
				return ret
			}
			return e
		})
		if !cont {
			return
		}
	}
}

func (m *memoPart[T]) Split(n uint64) (Seq[T], Seq[T]) {
	if n > m.upper-m.lower {
		return m, (*Vec[T])(nil)
	}
	left := &memoPart[T]{
		underlying: m.underlying,
		lower:      m.lower,
		upper:      m.lower + n,
	}
	right := &memoPart[T]{
		underlying: m.underlying,
		lower:      m.lower + n,
		upper:      m.upper,
	}
	return left, right
}

func (m *memoPart[T]) Take(n uint64) Seq[T] {
	if n > m.upper-m.lower {
		return m
	}
	left := &memoPart[T]{
		underlying: m.underlying,
		lower:      m.lower,
		upper:      m.lower + n,
	}
	return left
}

// Memo takes a Seq[T] `s` and returns a new memoized Seq[T] which is identical,
// except any computations involved in producing elements of `s` are cached,
// and subsequent accesses of those elements return the cached value.
//
// Keep in mind that Memo'd Seq's keep their values in memory, so Memo'ing
// unbounded sequences can lead to unbounded memory usage if you use them
// carelessly. Only Memo sequences when you have a specific reason to do so.
func Memo[T any](s Seq[T]) Seq[T] {
	return &memo[T]{
		s: s,
	}
}
