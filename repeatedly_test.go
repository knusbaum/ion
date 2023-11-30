package ion

import (
	"testing"
)

func TestRepeatedlyInf(t *testing.T) {
	// This one we can't use testInfSeq.
	s := Repeatedly(1)
	t.Run(t.Name()+"/elem", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			i, ok := s.Elem(100)
			if !ok {
				t.Fatalf("Expected s[100] to return a value, but got nothing.\n")
			}
			if i != 1 {
				t.Fatalf("Expected s[100] == 1, but got %d\n", i)
			}
		}
	})
	t.Run("/split-end", func(t *testing.T) {
		l, _ := s.Split(100)
		l, r := l.Split(100)

		ret := Fold(l, func(acc, i int) int {
			return acc + i
		})
		if ret != 100 {
			t.Fatalf("Expected ret == 100, but got %d\n", ret)
		}
		ret = Fold(r, func(acc, i int) int {
			return acc + i
		})
		if ret != 0 {
			t.Fatalf("Expected ret == 0, but got %d\n", ret)
		}
	})
	t.Run(t.Name()+"/split", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			l, r := s.Split(10000)

			// Make sure l contains what we expect
			res := Fold(l, func(acc, i int) int {
				return acc + i
			})
			if res != 10000 {
				t.Fatalf("Expected res == 10000, but was %d", res)
			}

			var j int
			l.Iterate(func(i int) bool {
				j += i
				return true
			})
			if j != 10000 {
				t.Fatalf("Expected res == 10000, but was %d", j)
			}

			// Make sure l only contains what we expect.
			e, ok := l.Elem(9999)
			if !ok {
				t.Fatalf("Expected l[9999] to return a value, but got nothing.\n")
			}
			if e != 1 {
				t.Fatalf("Expected l[9999] == 1, but got %d\n", e)
			}

			// Make sure r contains what we expect.
			e, ok = r.Elem(0)
			if !ok {
				t.Fatalf("Expected r[0] to return a value, but got nothing.\n")
			}
			if e != 1 {
				t.Fatalf("Expected r[0] == 1, but got %d\n", e)
			}

			// Make sure we can split l again
			ll, lr := l.Split(10)
			res = Fold(ll, func(acc, i int) int {
				return acc + i
			})
			if res != 10 {
				t.Fatalf("Expected res == 10, but was %d", res)
			}

			res = Fold(lr, func(acc, i int) int {
				return acc + i
			})
			if res != 9990 {
				t.Fatalf("Expected res == 9990, but was %d", res)
			}

			// Make sure we can split r again
			rl, rr := r.Split(10)
			res = Fold(rl, func(acc, i int) int {
				return acc + i
			})
			if res != 10 {
				t.Fatalf("Expected res == 10, but was %d", res)
			}

			e, ok = rr.Elem(0)
			if !ok {
				t.Fatalf("Expected rr[0] to return a value, but got nothing.\n")
			}
			if e != 1 {
				t.Fatalf("Expected rr[0] == 1, but got %d\n", e)
			}
		}
	})

	t.Run(t.Name()+"/take", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			l := s.Take(10000)

			// Make sure l contains what we expect
			res := Fold(l, func(acc, i int) int {
				return acc + i
			})
			if res != 10000 {
				t.Fatalf("Expected res == 10000, but was %d", res)
			}
		}
	})

	t.Run(t.Name()+"/iterate", func(t *testing.T) {
		for j := 0; j < 2; j++ {
			var i, res int

			s.Iterate(func(e int) bool {
				res += e
				i++
				if i == 10000 {
					return false
				}
				return true
			})

			// Make sure l contains what we expect
			if res != 10000 {
				t.Fatalf("Expected res == 10000, but was %d", res)
			}
		}
	})

	t.Run(t.Name()+"/lazy", func(t *testing.T) {
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
			if res != 10000 {
				t.Fatalf("Expected res == 10000, but was %d", res)
			}
		}
	})
}
