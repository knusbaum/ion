package ion_test

import (
	"fmt"

	"github.com/knusbaum/ion"
)

func ExampleMap() {

	// Create an unbounded list of integers
	n := ion.From[int](0, 1)

	// Map the integers to themselves mod 3
	n = ion.Map(n, func(i int) int {
		return i % 3
	})

	// Get the first 10 and print them
	n.Take(10).Iterate(func(i int) bool {
		fmt.Printf("%d ", i)
		return true
	})

	// Output: 0 1 2 0 1 2 0 1 2 0
}

func ExampleFilter() {
	// Create an unbounded list of natural numbers
	n := ion.From[int](1, 1)

	// Filter the even numbers
	n = ion.Filter(n, func(i int) bool {
		return i%2 == 0
	})

	// Get the first 10 and print them
	n.Take(10).Iterate(func(i int) bool {
		fmt.Printf("%d ", i)
		return true
	})

	// Output: 2 4 6 8 10 12 14 16 18 20
}

func ExampleGenerate() {
	// primes is a Seq[int] of prime numbers created by this generator
	primes := ion.Generate(func(state []int) (int, []int, bool) {
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

	// Grab the first 10 primes and print them
	primes.Take(10).Iterate(func(i int) bool {
		fmt.Printf("%d ", i)
		return true
	})

	// Output: 2 3 5 7 11 13 17 19 23 29
}
