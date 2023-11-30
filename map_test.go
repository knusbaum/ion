package ion

import (
	"fmt"
	"io"
	"testing"
	"time"
)

func TestSeq(t *testing.T) {

	t.Run("map", func(t *testing.T) {
		m := Repeatedly[int](10)
		m = Map(m, func(x int) int { return x + 1 })
		if e, ok := m.Elem(10); ok == false || e != 11 {
			t.Fatalf("Expected %d == 11", e)
		}
	})

	t.Run("repeat-fold", func(t *testing.T) {
		m := Repeatedly[int](10)
		some, m := m.Split(100)

		ret := Fold(some, func(acc, x int) int {
			return acc + x
		})

		if ret != 1000 {
			t.Fatalf("Expected %d == 1000", ret)
		}
	})

	t.Run("repeat-map-fold", func(t *testing.T) {
		m := Repeatedly[int](10)
		m = Map(m, func(x int) int {
			return x * 10
		})
		some, m := m.Split(100)

		ret := Fold(some, func(acc, x int) int {
			return acc + x
		})

		if ret != 10000 {
			t.Fatalf("Expected %d == 10000", ret)
		}
	})

	t.Run("from", func(t *testing.T) {
		sum := func(acc, x int) int { return acc + x }

		m := From[int](0, 10)
		some, m2 := m.Split(10)
		if s := Fold(some, sum); s != 450 {
			t.Fatalf("Expected %d == 450", s)
		}

		some2, _ := m.Split(10)
		// splitting m again should give the same some.
		if s := Fold(some2, sum); s != 450 {
			t.Fatalf("Expected %d == 450", s)
		}

		some3, _ := m2.Split(10)
		// Using m2 should give the next some.
		if s := Fold(some3, sum); s != 1450 {
			t.Fatalf("Expected %d == 1450", s)
		}
	})

	t.Run("generate", func(t *testing.T) {
		sum := func(acc, x int) int { return acc + x }

		// Unbounded sequence of fibonacci numbers
		fibgen := Generate(func(state [2]int) (int, [2]int, bool) {
			if state[1] == 0 {
				state[1] = 1
				return 0, state, true
			}
			next := state[0] + state[1]
			ret := state[1]
			state[0], state[1] = state[1], next
			return ret, state, true
		})
		// first 10 (0,1,1,2 ..) sum to 88
		some, fg2 := fibgen.Split(10)
		if s := Fold(some, sum); s != 88 {
			t.Fatalf("Expected %d == 88", s)
		}

		// fg2 should give the next 10 (34,55,89...) which sum to
		some2, _ := fg2.Split(10)
		if s := Fold(some2, sum); s != 10857 {
			t.Fatalf("Expected %d == 10857", s)
		}

		// the original fibgen should give the original values *and* be able to
		// continue generating from the sequence.
		// (0,1,1,2,3, ... 55,89,144, ...) Should equal 88 + 10857 == 10945
		some3, _ := fibgen.Split(20)
		if s := Fold(some3, sum); s != 10945 {
			t.Fatalf("Expected %d == 10945", s)
		}
	})

	t.Run("filter", func(t *testing.T) {
		sum := func(acc, x int) int { return acc + x }
		isEven := func(i int) bool { return i%2 == 0 }
		// Unbounded sequence of natural numbers
		naturals := From[int](1, 1)
		// Filter the unbounded list, giving only the evens
		evens := Filter(naturals, isEven)
		// Take 10
		some, _ := evens.Split(10)
		if e, ok := some.Elem(0); ok == false || e != 2 {
			t.Fatalf("Expected element 0 to be 2, but was %d\n", e)
		}
		if l := len(ToSlice(some)); l != 10 {
			t.Fatalf("Expected len(some) == 10, but it was %d\n", l)
		}
		if s := Fold(some, sum); s != 110 {
			t.Fatalf("Expected %d == 110", s)
		}
	})

	t.Run("filter-prime", func(t *testing.T) {
		sum := func(acc, x int) int { return acc + x }
		isPrime := func(n int) bool {
			if n == 1 {
				return false
			}

			i := 2
			// check all integers between 2 and sqrt(n)
			for i*i <= n {
				if n%i == 0 {
					return false
				}
				i += 1
			}
			return true
		}

		// Unbounded sequence of natural numbers
		naturals := From[int](1, 1)
		// Filter the unbounded list, giving only the prime numbers
		primes := Filter(naturals, isPrime)
		// Take 20
		some, _ := primes.Split(20)
		if l := len(ToSlice(some)); l != 20 {
			t.Fatalf("Expected 20 elements, but only got %d\n", l)
		}
		if s := Fold(some, sum); s != 639 {
			t.Fatalf("Expected %d == 639", s)
		}
	})

	t.Run("filter-prime", func(t *testing.T) {
		sum := func(acc, x int) int { return acc + x }
		isPrime := func(n int) bool {
			if n == 1 {
				return false
			}

			i := 2
			// check all integers between 2 and sqrt(n)
			for i*i <= n {
				if n%i == 0 {
					return false
				}
				i += 1
			}
			return true
		}
		result := Fold(Filter(From[int](1, 1), isPrime).Take(1000), sum)
		if result != 3682913 {
			t.Fatalf("Expected %d == 3682913", result)
		}
	})

	t.Run("prime-sieve", func(t *testing.T) {
		sum := func(acc, x int) int { return acc + x }
		primes := Generate(func(state []int) (int, []int, bool) {
			// Sieve of erasthenes (sort of)
			// state is the list of primes followed by the current number
			if len(state) == 0 {
				return 2, []int{2, 3}, true
			}
			i := state[len(state)-1]
			state = state[:len(state)-1]
		outter:
			for {
				for _, p := range state {
					if i%p == 0 {
						i += 1
						continue outter
					}
				}
				state = append(state, i, i+1)
				return i, state, true
			}
		})

		// Take 20
		some, _ := primes.Split(20)
		if l := len(ToSlice(some)); l != 20 {
			t.Fatalf("Expected 20 elements, but got %d\n", l)
		}
		if s := Fold(some, sum); s != 639 {
			t.Fatalf("Expected %d == 639", s)
		}
	})

	t.Run("lazy", func(t *testing.T) {
		primes := Generate(func(state []int) (int, []int, bool) {
			// Sieve of erasthenes (sort of)
			// state is the list of primes followed by the current number
			if len(state) == 0 {
				return 2, []int{2, 3}, true
			}
			i := state[len(state)-1]
			state = state[:len(state)-1]
		outter:
			for {
				for _, p := range state {
					if i%p == 0 {
						i += 1
						continue outter
					}
				}
				state = append(state, i, i+1)
				return i, state, true
			}
		})

		some := primes.Take(100001)
		//some, _ := primes.Split(100001)
		e, ok := some.Elem(0)
		if !ok {
			t.Fatalf("Expected an element.")
		}
		t.Logf("Zero: %d\n", e)
		//fmt.Printf("100000:  %d\n", some.Elem(100000))
		// TODO: Assert laziness somehow.
	})

	t.Run("lazy2", func(t *testing.T) {
		naturals := From[int](1, 1)
		filtered := Filter(naturals, func(i int) bool {
			return i%2 == 0
		})

		//some, _ := filtered.Split(100000001)
		some := filtered.Take(100000001)
		e, ok := some.Elem(0)
		if !ok {
			t.Fatalf("Expected an element.")
		}
		t.Logf("Zero: %d\n", e)
		//fmt.Printf("100000:  %d\n", some.Elem(100000000))
		// TODO: Assert laziness somehow.
	})

	// t.Run("play", func(t *testing.T) {
	// 	primes := Generate(func(state []int) (int, []int) {
	// 		// Sieve of erasthenes (sort of)
	// 		// state is the list of primes followed by the current number
	// 		if len(state) == 0 {
	// 			return 2, []int{2, 3}
	// 		}
	// 		i := state[len(state)-1]
	// 		state = state[:len(state)-1]
	// 	outter:
	// 		for {
	// 			for _, p := range state {
	// 				if i%p == 0 {
	// 					i += 1
	// 					continue outter
	// 				}
	// 			}
	// 			state = append(state, i, i+1)
	// 			return i, state
	// 		}
	// 	})

	// 	some, next := primes.Split(10)
	// 	fmt.Printf("First 10: %v\n", ToSlice(some))

	// 	primes = Map(primes, func(i int) int {
	// 		fmt.Printf("mapping %d\n", i)
	// 		return i + 1
	// 	})

	// 	fmt.Printf("Split\n")
	// 	some2, _ := primes.Split(10)
	// 	fmt.Printf("Done.\n")
	// 	fmt.Printf("First 10 again: %v\n", ToSlice(some2))

	// 	some3, _ := next.Split(10)
	// 	fmt.Printf("Next 10: %v\n", ToSlice(some3))
	// })
}

// Regression test. The "need" calculation in Elem functions was
// using unsigned ints and underflowing, creating needs of close to math.MaxUint64
func TestMapSplitElem(t *testing.T) {
	seq := Generate(func(state int) (int, int, bool) {
		return state + 1, state + 1, true
	})
	seq = Map(seq, func(i int) int { return i + 1 })

	done := make(chan struct{})

	go func() {
		_, tenth := seq.Split(10)
		e, ok := tenth.Elem(0)
		if !ok {
			t.Fatalf("Expected an element.")
		}
		fmt.Fprintf(io.Discard, "tenth: %d\n", e)

		e, ok = seq.Elem(0)
		if !ok {
			t.Fatalf("Expected an element.")
		}
		fmt.Fprintf(io.Discard, "zeroth: %d\n", e)
		close(done)
	}()

	// Expect the element to be calculated in < 1 second.
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out.")
	}
}

func TestMapSplit(t *testing.T) {
	seq := Generate(func(state int) (int, int, bool) {
		return state, state + 1, true
	})

	m := Map(seq, func(i int) int {
		return i + 1
	})

	m1, m2 := m.Split(1000)

	if e, ok := m.Elem(10); ok == false || e != 11 {
		t.Fatalf("Expected m[10] == 11, but was %d\n", e)
	}

	if e, ok := m1.Elem(10); ok == false || e != 11 {
		t.Fatalf("Expected m1[10] == 11, but was %d\n", e)
	}
	if e, ok := m2.Elem(10); ok == false || e != 1011 {
		t.Fatalf("Expected m2[10] == 1011, but was %d\n", e)
	}
}

func BenchmarkPrimes(b *testing.B) {
	b.Run("sieve", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			primes := Generate(func(state []int) (int, []int, bool) {
				// Sieve of erasthenes (sort of)
				// state is the list of primes followed by the current number
				if len(state) == 0 {
					return 2, []int{2, 3}, true
				}
				i := state[len(state)-1]
				state = state[:len(state)-1]
			outter:
				for {
					for _, p := range state {
						if i%p == 0 {
							i += 1
							continue outter
						}
					}
					state = append(state, i, i+1)
					return i, state, true
				}
			})

			// Take 20
			primes.Split(2000)
		}
	})

	b.Run("sqrt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			isPrime := func(n int) bool {
				if n == 1 {
					return false
				}

				i := 2
				// check all integers between 2 and sqrt(n)
				for i*i <= n {
					if n%i == 0 {
						return false
					}
					i += 1
				}
				return true
			}

			// Unbounded sequence of natural numbers
			naturals := From[int](1, 1)
			// Filter the unbounded list, giving only the prime numbers
			primes := Filter(naturals, isPrime)
			// Take 20
			primes.Split(2000)
		}
	})
}
