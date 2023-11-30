// package result is an experiment in monadic types.
// The main type is the Res[T], representing the result of some computation.
// It is useful to use Res in ion.Seq[Res[T]] sequences, since it can
// capture and propagate errors during I/O or other non-pure functions.
package result

import "github.com/knusbaum/ion"

// We can't have a Monad type since U must be
// resolved at instantiation time, meaning Map() cannot
// be polymorphic, unless we use interface{}.
// In other words, a type `someStruct[T any]` cannot have methods (like map)
// which are themselves polymorphic beyond the type parameters given
// upon someStruct's instantiation. So:
//  func (t *someStruct[T]) Hello(e T) {}
// is fine, but
//  func(t *someStruct[T])[U any]map(func(e T) U) T[U]
// is not legal.
// There's also no way to specify that the result of calling Map
// on type T[U], func(U) V is T[V].
//type Monad[T any] interface {
//	ID() T
//	Map(func (e T) U) U
//}
//
// type Monad[T any] interface {
// 	Map(func(e T) U) any
// }
//
// func Map[T, U any, M [T]Monad](e M, f func(e T) U) M[U] {
// 	return e.Map(f)
// }

// Res contains the result of a computation. This can be a value
// of type T, or an error.
// Note: Experimental
type Res[T any] struct {
	e   T
	err error
}

// Get returns the value of type T, and any error held by the Res[T].
// If ther error is not nil, the T is the zero value.
func (r *Res[T]) Get() (T, error) {
	return r.e, r.err
}

// OK returns a new Res[T] containing `e`.
func Ok[T any](e T) Res[T] {
	return Res[T]{
		e: e,
	}
}

// Err returns a new Res[T] containing the error `e`.
func Err[T any](e error) Res[T] {
	return Res[T]{
		err: e,
	}
}

// FMap f func(T) -> func(Res[T]) Res[U]
// FMap takes a function `f` that takes a value of type T and returns
// a value of type U. FMap returns a func that takes a value of type Res[T]
// and applies `f` to the value of type T contained in the result, wrapping
// the value of type U into a Result.
//
// If the Res[T] is an error result, the function `f` is not called on it, and
// the Res[T] is converted to Res[U], its error.
func FMap[T, U any](f func(e T) U) func(Res[T]) Res[U] {
	return func(r Res[T]) Res[U] {
		if r.err != nil {
			return Err[U](r.err)
		}
		return Ok(f(r.e))
	}
}

// Apply f func(T) -> func(Res[T])
// Apply takes a function `f` that takes a value of type T and Apply returns a
// func that takes a value of type Res[T] and applies `f` to the value of type T
// contained in the result. If the Res[T] is an error result, the function `f`
// is not called on it.
func Apply[T any](f func(e T)) func(Res[T]) {
	return func(r Res[T]) {
		if r.err != nil {
			return
		}
		f(r.e)
	}
}

// Handle f func(error) -> func(Res[T]) Res[T]
// Handle takes a function designed to handle errors and returns
// a function that will apply that handler function to a Res[T],
// if the Res[T] is an error result.
func Handle[T any](f func(e error)) func(Res[T]) Res[T] {
	return func(r Res[T]) Res[T] {
		if r.err != nil {
			f(r.err)
		}
		return r
	}
}

// Map s Seq[Res[T]] -> f func(T) U -> Seq[Res[U]]
// Map is analogous to ion.Map, only it maps a function from T to U over
// a Seq[Result[T]], returning Seq[Result[U]].
//
// The resulting sequence contains f applied to the elements of type T
// of the Res[T]s where those Res[T]'s are not errors. Res[T] elements that
// are errors are converted from Res[T] to Res[U] retaining their errors.
//
// Map(s, f) is equivalent to
//
//	ion.Map(s, FMap(f))
func Map[T, U any](s ion.Seq[Res[T]], f func(T) U) ion.Seq[Res[U]] {
	return ion.Map(s, FMap(f))
}
