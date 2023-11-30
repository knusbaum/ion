// package ion provides basic immutable data structures and functions on them,
// enabling writing Go programs in a functional style.
//
// This includes immutable trees and sequences, Map, Fold, Filter functions
// and a few other things.
//
// Computations such as those done by Map, Filter, and others are made as
// lazy as possible. This means that most of the operations on a Seq[T] are
// not actually executed until one or more elements of that sequence are
// realized with Elem or Iterate.
//
// The main functionality of this package are the functions operating on
// the Seq[T] type. Take a look at type Seq in the index below.
//
// See the programs in cmd/ for examples demonstrating the usage.
package ion
