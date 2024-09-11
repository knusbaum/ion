package ion

import (
	"math"
	"testing"
)

// testInfSeq expects s to be an unbounded Seq[int] starting at 0 and incrementing by 1
func testInfSeq(t *testing.T, s Seq[int]) {
	t.Run("elem", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			i, ok := s.Elem(100)
			if !ok {
				t.Fatalf("Expected s[100] to return a value, but got nothing.\n")
			}
			if i != 100 {
				t.Fatalf("Expected s[100] == 100, but got %d\n", i)
			}
		}
	})
	t.Run("split-end", func(t *testing.T) {
		l, _ := s.Split(100)
		l, r := l.Split(100)

		ret := Fold(l, func(acc, i int) int {
			return acc + i
		})
		if ret != 4950 {
			t.Fatalf("Expected ret == 4950, but got %d\n", ret)
		}

		ret = Fold(r, func(acc, i int) int {
			return acc + i
		})
		if ret != 0 {
			t.Fatalf("Expected ret == 0, but got %d\n", ret)
		}
	})
	t.Run("split", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			l, r := s.Split(10000)

			// Make sure l contains what we expect
			res := Fold(l, func(acc, i int) int {
				return acc + i
			})
			if res != 49995000 {
				t.Fatalf("Expected res == 49995000, but was %d", res)
			}

			var j int
			l.Iterate(func(i int) bool {
				j += i
				return true
			})
			if j != 49995000 {
				t.Fatalf("Expected res == 49995000, but was %d", j)
			}

			// Make sure l only contains what we expect.
			e, ok := l.Elem(9999)
			if !ok {
				t.Fatalf("Expected l[9999] to return a value, but got nothing.\n")
			}
			if e != 9999 {
				t.Fatalf("Expected l[9999] == 9999, but got %d\n", e)
			}

			// Make sure r contains what we expect.
			e, ok = r.Elem(0)
			if !ok {
				t.Fatalf("Expected r[0] to return a value, but got nothing.\n")
			}
			if e != 10000 {
				t.Fatalf("Expected r[0] == 10000, but got %d\n", e)
			}

			// Make sure we can split l again
			ll, lr := l.Split(10)
			res = Fold(ll, func(acc, i int) int {
				return acc + i
			})
			if res != 45 {
				t.Fatalf("Expected res == 45, but was %d", res)
			}

			res = Fold(lr, func(acc, i int) int {
				return acc + i
			})
			if res != 49994955 {
				t.Fatalf("Expected res == 49994955, but was %d", res)
			}

			// Make sure we can split r again
			rl, rr := r.Split(10)
			res = Fold(rl, func(acc, i int) int {
				return acc + i
			})
			if res != 100045 {
				t.Fatalf("Expected res == 100045, but was %d", res)
			}

			e, ok = rr.Elem(0)
			if !ok {
				t.Fatalf("Expected rr[0] to return a value, but got nothing.\n")
			}
			if e != 10010 {
				t.Fatalf("Expected rr[0] == 10010, but got %d\n", e)
			}
		}
	})

	t.Run("take", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			l := s.Take(10000)

			// Make sure l contains what we expect
			res := Fold(l, func(acc, i int) int {
				return acc + i
			})
			if res != 49995000 {
				t.Fatalf("Expected res == 49995000, but was %d", res)
			}
		}
	})

	t.Run("iterate", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			var i, res int
			s.Iterate(func(e int) bool {
				res += e
				i++
				if i >= 10000 {
					return false
				}
				return true
			})

			// Make sure l contains what we expect
			if res != 49995000 {
				t.Fatalf("Expected res == 49995000, but was %d", res)
			}
		}
	})

	t.Run("lazy", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			var i, res int

			s.Lazy(func(e func() int) bool {
				res += e()
				i++
				if i == 10000 {
					return false
				}
				return true
			})

			// Make sure l contains what we expect
			if res != 49995000 {
				t.Fatalf("Expected res == 49995000, but was %d", res)
			}
		}
	})
}

// testFinSeq expects s to be a Seq[int] of length 10,000 starting at 0 and incrementing by 1
func testFinSeq(t *testing.T, s Seq[int]) {
	t.Run("elem", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			i, ok := s.Elem(9_999)
			if !ok {
				t.Fatalf("Expected s[9_999] to return a value, but got nothing.\n")
			}
			if i != 9_999 {
				t.Fatalf("Expected s[9_999] == 9_999, but got %d\n", i)
			}

			i, ok = s.Elem(10_000)
			if ok {
				t.Fatalf("Expected s[10_000] to return no value, but got %d.\n", i)
			}
			if i != 0 {
				t.Fatalf("Expected s[10_000] == 0, but got %d\n", i)
			}
		}
	})
	t.Run("split-end", func(t *testing.T) {
		l, _ := s.Split(10000)
		l, r := l.Split(10000)

		ret := Fold(l, func(acc, i int) int {
			return acc + i
		})
		if ret != 49995000 {
			t.Fatalf("Expected ret == 49995000, but got %d\n", ret)
		}
		ret = Fold(r, func(acc, i int) int {
			return acc + i
		})
		if ret != 0 {
			t.Fatalf("Expected ret == 0, but got %d\n", ret)
		}
	})
	t.Run("split", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			l, r := s.Split(1000)

			// Make sure l contains what we expect
			res := Fold(l, func(acc, i int) int {
				return acc + i
			})
			if res != 499500 {
				t.Fatalf("Expected res == 499500, but was %d", res)
			}

			var j int
			l.Iterate(func(i int) bool {
				j += i
				return true
			})
			if j != 499500 {
				t.Fatalf("Expected res == 499500, but was %d", j)
			}

			// Make sure l only contains what we expect.
			e, ok := l.Elem(9999)
			if ok {
				t.Fatalf("Expected l[9999] to return no value, but got %d.\n", e)
			}
			if e != 0 {
				t.Fatalf("Expected l[9999] == 0, but got %d\n", e)
			}

			// Make sure r contains what we expect.
			e, ok = r.Elem(0)
			if !ok {
				t.Fatalf("Expected r[0] to return a value, but got nothing.\n")
			}
			if e != 1000 {
				t.Fatalf("Expected r[0] == 1000, but got %d\n", e)
			}

			// Make sure we can split l again
			ll, lr := l.Split(10)
			res = Fold(ll, func(acc, i int) int {
				return acc + i
			})
			if res != 45 {
				t.Fatalf("Expected res == 45, but was %d", res)
			}

			res = Fold(lr, func(acc, i int) int {
				return acc + i
			})
			if res != 499455 {
				t.Fatalf("Expected res == 499455, but was %d", res)
			}

			// Make sure we can split r again
			rl, rr := r.Split(10)
			res = Fold(rl, func(acc, i int) int {
				return acc + i
			})
			if res != 10045 {
				t.Fatalf("Expected res == 10045, but was %d", res)
			}

			e, ok = rr.Elem(0)
			if !ok {
				t.Fatalf("Expected rr[0] to return a value, but got nothing.\n")
			}
			if e != 1010 {
				t.Fatalf("Expected rr[0] == 1010, but got %d\n", e)
			}

			sum := func(s Seq[int]) int {
				return Fold(s, func(acc, i int) int {
					return acc + i
				})
			}

			// Make sure s split on idx > len(s) does not panic.
			sbig, snothing := s.Split(200_000)
			if res := sum(sbig); res != 49995000 {
				t.Fatalf("Expected sum(sbig) == 49995000, but was %d", res)
			}
			if res := sum(snothing); res != 0 {
				t.Fatalf("Expected sum(snothing) == 0, but was %d", res)
			}

			// Make sure l split on idx > len(l) does not panic.
			lbig, lnothing := l.Split(2000)
			if res := sum(lbig); res != 499500 {
				t.Fatalf("Expected sum(lbig) == 499500, but was %d", res)
			}
			if res := sum(lnothing); res != 0 {
				t.Fatalf("Expected sum(lnothing) == 0, but was %d", res)
			}

			// Make sure r split on idx > len(r) does not panic.
			re, ok := r.Elem(0)
			if !ok {
				t.Fatalf("Expected r[0] to return a value, but got nothing.\n")
			}
			if re != 1000 {
				t.Fatalf("Expected r == 1000, but was %d\n", re)
			}
			rbig, rnothing := r.Split(10000)
			if res := sum(rbig); res != 49495500 {
				t.Fatalf("Expected sum(rbig) == 49495500, but was %d", res)
			}
			if res := sum(rnothing); res != 0 {
				t.Fatalf("Expected sum(rnothing) == 0, but was %d", res)
			}
		}
	})

	t.Run("take", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			l := s.Take(200000)
			var s int
			// Make sure l contains what we expect
			res := Fold(l, func(acc, i int) int {
				s += i
				return acc + i
			})
			if res != 49995000 {
				t.Fatalf("Expected res == 49995000, but was %d", res)
			}
		}
	})

	t.Run("iterate", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			var res int
			s.Iterate(func(e int) bool {
				res += e
				return true
			})

			// Make sure l contains what we expect
			if res != 49995000 {
				t.Fatalf("Expected res == 49995000, but was %d", res)
			}
		}
	})

	t.Run("lazy", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			var i, res int

			s.Lazy(func(e func() int) bool {
				res += e()
				i++
				//fmt.Printf("I: %d\n", i)
				if i > 10000 {
					t.Fatalf("OVER")
				}
				return true
			})

			// Make sure l contains what we expect
			if res != 49995000 {
				t.Fatalf("Expected res == 49995000, but was %d", res)
			}
		}
	})
}

func TestFilter(t *testing.T) {
	g := Generate(func(state [2]int) (int, [2]int, bool) {
		// state[0] is next. state[1] starts as int max.
		// state[0] and state[1] alternate, generating a sequence that looks like:
		// 0, max, 1, max, 2, max, 3, max
		// This lets the Filter remove the max int values, leaving the expected sequence
		// 0, 1, 2, 3, ...

		if state[0] == 0 {
			// initial
			state[1] = math.MaxInt
		}

		ret := state[0]
		if state[0] != math.MaxInt {
			state[0]++
		}
		state[0], state[1] = state[1], state[0]
		return ret, state, true

	})
	t.Run("inf", func(t *testing.T) {
		testInfSeq(t, Filter(g, func(i int) bool {
			return i != math.MaxInt
		}))
	})
	t.Run("fin1", func(t *testing.T) {
		testFinSeq(t, Filter(g, func(i int) bool {
			return i != math.MaxInt
		}).Take(10000))
	})
	t.Run("fin2", func(t *testing.T) {
		l, _ := Filter(g, func(i int) bool {
			return i != math.MaxInt
		}).Split(10000)
		testFinSeq(t, l)
	})
}

func TestFrom(t *testing.T) {
	t.Run("inf", func(t *testing.T) {
		testInfSeq(t, From[int](0, 1))
	})
	t.Run("fin1", func(t *testing.T) {
		testFinSeq(t, From[int](0, 1).Take(10000))
	})
	t.Run("fin2", func(t *testing.T) {
		l, _ := From[int](0, 1).Split(10000)
		testFinSeq(t, l)
	})
}

func TestGenerate(t *testing.T) {
	t.Run("inf", func(t *testing.T) {
		testInfSeq(t, Generate(func(state int) (int, int, bool) {
			return state, state + 1, true
		}))
	})
	t.Run("fin1", func(t *testing.T) {
		s := Generate(func(state int) (int, int, bool) {
			if state >= 10000 {
				return 0, state, false
			}
			return state, state + 1, true
		})
		testFinSeq(t, s)
	})
	t.Run("fin2", func(t *testing.T) {
		testFinSeq(t, Generate(func(state int) (int, int, bool) {
			return state, state + 1, true
		}).Take(10000))
	})
	t.Run("fin3", func(t *testing.T) {
		l, _ := Generate(func(state int) (int, int, bool) {
			return state, state + 1, true
		}).Split(10000)
		testFinSeq(t, l)
	})
}

func TestGenerateInit(t *testing.T) {
	t.Run("inf", func(t *testing.T) {
		testInfSeq(t, GenerateInit(100, func(state int) (int, int, bool) {
			return state - 100, state + 1, true
		}))
	})
	t.Run("fin1", func(t *testing.T) {
		s := GenerateInit(100, func(state int) (int, int, bool) {
			if state-100 >= 10000 {
				return 0, state, false
			}
			return state - 100, state + 1, true
		})
		testFinSeq(t, s)
	})
	t.Run("fin2", func(t *testing.T) {
		testFinSeq(t, GenerateInit(100, func(state int) (int, int, bool) {
			return state - 100, state + 1, true
		}).Take(10000))
	})
	t.Run("fin3", func(t *testing.T) {
		l, _ := GenerateInit(100, func(state int) (int, int, bool) {
			return state - 100, state + 1, true
		}).Split(10000)
		testFinSeq(t, l)
	})
}

func TestMap(t *testing.T) {
	s := Generate(func(state int) (int, int, bool) {
		return state + 1, state + 1, true
	})

	t.Run("inf", func(t *testing.T) {
		testInfSeq(t, Map(s, func(i int) int {
			return i - 1
		}))
	})
	t.Run("fin1", func(t *testing.T) {
		testFinSeq(t, Map(s, func(i int) int {
			return i - 1
		}).Take(10000))
	})
	t.Run("fin2", func(t *testing.T) {
		l, _ := Map(s, func(i int) int {
			return i - 1
		}).Split(10000)
		testFinSeq(t, l)
	})
}

func TestStateGen(t *testing.T) {
	t.Run("inf", func(t *testing.T) {
		var i int
		testInfSeq(t, StateGen(func() (int, bool) {
			ret := i
			i++
			return ret, true
		}))
	})
	t.Run("fin1", func(t *testing.T) {
		var i int
		testFinSeq(t, StateGen(func() (int, bool) {
			ret := i
			i++
			return ret, true
		}).Take(10000))
	})
	t.Run("fin2", func(t *testing.T) {
		var i int
		l, _ := StateGen(func() (int, bool) {
			ret := i
			i++
			return ret, true
		}).Split(10000)
		testFinSeq(t, l)
	})
}

func TestMemo(t *testing.T) {
	t.Run("inf", func(t *testing.T) {
		var i int
		testInfSeq(t, Memo(StateGen(func() (int, bool) {
			ret := i
			i++
			return ret, true
		})))
	})
	t.Run("fin1", func(t *testing.T) {
		var i int
		testFinSeq(t, Memo(StateGen(func() (int, bool) {
			ret := i
			i++
			return ret, true
		})).Take(10000))
	})
	t.Run("fin2", func(t *testing.T) {
		var i int
		l, _ := Memo(StateGen(func() (int, bool) {
			ret := i
			i++
			return ret, true
		})).Split(10000)
		testFinSeq(t, l)
	})
	t.Run("fin3", func(t *testing.T) {
		var i int
		testFinSeq(t, Memo(StateGen(func() (int, bool) {
			ret := i
			i++
			return ret, true
		}).Take(10000)))
	})
	t.Run("fin4", func(t *testing.T) {
		var i int
		l, _ := StateGen(func() (int, bool) {
			ret := i
			i++
			return ret, true
		}).Split(10000)
		testFinSeq(t, Memo(l))
	})
}
