package ion

import (
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
	m.m.Lock()
	defer m.m.Unlock()
	memlen := m.mems.Len()

	var sl, sr Seq[T]
	if n > memlen {
		splits := n - m.mems.Len()
		sl, sr = m.s.Split(splits)
	} else {
		sl = (*Vec[T])(nil)
		sr = m.s
	}
	var ml, mr Seq[T]
	if n >= m.mems.Len() {
		ml = m.mems
		mr = (*Vec[T])(nil)
	} else {
		ml, mr = m.mems.Split(n)
	}
	l := &memo[T]{
		s:    sl,
		mems: ml.(*Vec[T]),
	}
	r := &memo[T]{
		s:    sr,
		mems: mr.(*Vec[T]),
	}
	return l, r
}

func (m *memo[T]) Take(n uint64) Seq[T] {
	m.m.Lock()
	defer m.m.Unlock()
	splits := int64(n) - int64(m.mems.Len())
	var sl Seq[T]
	if splits > 0 {
		sl = m.s.Take(uint64(splits))
	} else {
		sl = (*Vec[T])(nil)
	}
	var ml Seq[T]
	if n >= m.mems.Len() {
		ml = m.mems
	} else {
		ml = m.mems.Take(n)
	}
	l := &memo[T]{
		s:    sl,
		mems: ml.(*Vec[T]),
	}
	return l
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
