package iter

import (
	"iter"
)

// Seq is an alias for iter.Seq for convenience.
type Seq[V any] = iter.Seq[V]

// Seq2 is an alias for iter.Seq2 for convenience.
type Seq2[K, V any] = iter.Seq2[K, V]

// Result represents either a value or an error.
// This is useful for iterators that can produce errors.
type Result[T any] struct {
	Value T
	Err   error
}

// ResultSeq is a convenience type for iterators that produce value-error pairs.
type ResultSeq[T any] = Seq2[T, error]

// Filter returns an iterator that yields only elements satisfying the predicate.
// For Seq2[T, error] iterators, errors are always yielded (not filtered).
func Filter[T any](it Seq2[T, error], pred func(T) bool) Seq2[T, error] {
	return func(yield func(T, error) bool) {
		it(func(v T, err error) bool {
			if err != nil {
				return yield(v, err)
			}
			if pred(v) {
				return yield(v, nil)
			}
			return true // continue iteration
		})
	}
}

// Map transforms each element in the iterator.
// For Seq2[T, error] iterators, errors pass through untransformed.
func Map[T, U any](it Seq2[T, error], fn func(T) U) Seq2[U, error] {
	return func(yield func(U, error) bool) {
		it(func(v T, err error) bool {
			if err != nil {
				var zero U
				return yield(zero, err)
			}
			return yield(fn(v), nil)
		})
	}
}

// MapErr transforms each element, allowing the transform to return an error.
func MapErr[T, U any](it Seq2[T, error], fn func(T) (U, error)) Seq2[U, error] {
	return func(yield func(U, error) bool) {
		it(func(v T, err error) bool {
			if err != nil {
				var zero U
				return yield(zero, err)
			}
			result, resultErr := fn(v)
			return yield(result, resultErr)
		})
	}
}

// Take returns an iterator that yields at most n elements.
func Take[T any](it Seq2[T, error], n int) Seq2[T, error] {
	return func(yield func(T, error) bool) {
		count := 0
		it(func(v T, err error) bool {
			if count >= n {
				return false
			}
			if !yield(v, err) {
				return false
			}
			if err == nil {
				count++
			}
			return count < n
		})
	}
}

// Skip returns an iterator that skips the first n elements.
func Skip[T any](it Seq2[T, error], n int) Seq2[T, error] {
	return func(yield func(T, error) bool) {
		count := 0
		it(func(v T, err error) bool {
			// Always yield errors
			if err != nil {
				return yield(v, err)
			}
			if count < n {
				count++
				return true
			}
			return yield(v, nil)
		})
	}
}

// TakeWhile returns an iterator that yields elements while the predicate is true.
func TakeWhile[T any](it Seq2[T, error], pred func(T) bool) Seq2[T, error] {
	return func(yield func(T, error) bool) {
		it(func(v T, err error) bool {
			if err != nil {
				return yield(v, err)
			}
			if !pred(v) {
				return false
			}
			return yield(v, nil)
		})
	}
}

// SkipWhile returns an iterator that skips elements while the predicate is true.
func SkipWhile[T any](it Seq2[T, error], pred func(T) bool) Seq2[T, error] {
	return func(yield func(T, error) bool) {
		skipping := true
		it(func(v T, err error) bool {
			if err != nil {
				return yield(v, err)
			}
			if skipping && pred(v) {
				return true
			}
			skipping = false
			return yield(v, nil)
		})
	}
}

// Collect gathers all elements from an iterator into a slice.
// Returns the first error encountered, if any.
func Collect[T any](it Seq2[T, error]) ([]T, error) {
	var result []T
	for v, err := range it {
		if err != nil {
			return result, err
		}
		result = append(result, v)
	}
	return result, nil
}

// CollectN gathers at most n elements from an iterator into a slice.
// Returns the first error encountered, if any.
func CollectN[T any](it Seq2[T, error], n int) ([]T, error) {
	return Collect(Take(it, n))
}

// ForEach applies a function to each element in the iterator.
// Returns the first error encountered, if any.
func ForEach[T any](it Seq2[T, error], fn func(T)) error {
	for v, err := range it {
		if err != nil {
			return err
		}
		fn(v)
	}
	return nil
}

// ForEachErr applies a function that can return an error to each element.
// Returns the first error encountered from either the iterator or the function.
func ForEachErr[T any](it Seq2[T, error], fn func(T) error) error {
	for v, err := range it {
		if err != nil {
			return err
		}
		if fnErr := fn(v); fnErr != nil {
			return fnErr
		}
	}
	return nil
}

// Count returns the number of elements in the iterator.
// Returns the first error encountered, if any.
func Count[T any](it Seq2[T, error]) (int, error) {
	count := 0
	for _, err := range it {
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// First returns the first element from the iterator.
// Returns false if the iterator is empty.
func First[T any](it Seq2[T, error]) (T, bool, error) {
	for v, err := range it {
		if err != nil {
			var zero T
			return zero, false, err
		}
		return v, true, nil
	}
	var zero T
	return zero, false, nil
}

// Last returns the last element from the iterator.
// Returns false if the iterator is empty.
func Last[T any](it Seq2[T, error]) (T, bool, error) {
	var last T
	found := false
	for v, err := range it {
		if err != nil {
			return last, found, err
		}
		last = v
		found = true
	}
	return last, found, nil
}

// Find returns the first element satisfying the predicate.
// Returns false if no element matches.
func Find[T any](it Seq2[T, error], pred func(T) bool) (T, bool, error) {
	for v, err := range it {
		if err != nil {
			var zero T
			return zero, false, err
		}
		if pred(v) {
			return v, true, nil
		}
	}
	var zero T
	return zero, false, nil
}

// Any returns true if any element satisfies the predicate.
func Any[T any](it Seq2[T, error], pred func(T) bool) (bool, error) {
	for v, err := range it {
		if err != nil {
			return false, err
		}
		if pred(v) {
			return true, nil
		}
	}
	return false, nil
}

// All returns true if all elements satisfy the predicate.
// Returns true for empty iterators.
func All[T any](it Seq2[T, error], pred func(T) bool) (bool, error) {
	for v, err := range it {
		if err != nil {
			return false, err
		}
		if !pred(v) {
			return false, nil
		}
	}
	return true, nil
}

// None returns true if no element satisfies the predicate.
func None[T any](it Seq2[T, error], pred func(T) bool) (bool, error) {
	return All(it, func(v T) bool { return !pred(v) })
}

// Enumerate adds an index to each element in the iterator.
func Enumerate[T any](it Seq2[T, error]) Seq2[IndexedValue[T], error] {
	return func(yield func(IndexedValue[T], error) bool) {
		idx := 0
		it(func(v T, err error) bool {
			if err != nil {
				return yield(IndexedValue[T]{}, err)
			}
			if !yield(IndexedValue[T]{Index: idx, Value: v}, nil) {
				return false
			}
			idx++
			return true
		})
	}
}

// IndexedValue pairs an index with a value.
type IndexedValue[T any] struct {
	Index int
	Value T
}

// Reduce combines all elements using a binary function.
// Returns the initial value if the iterator is empty.
func Reduce[T, U any](it Seq2[T, error], initial U, fn func(U, T) U) (U, error) {
	acc := initial
	for v, err := range it {
		if err != nil {
			return acc, err
		}
		acc = fn(acc, v)
	}
	return acc, nil
}

// Chunk splits an iterator into chunks of the specified size.
func Chunk[T any](it Seq2[T, error], size int) Seq2[[]T, error] {
	return func(yield func([]T, error) bool) {
		chunk := make([]T, 0, size)
		hadError := false
		it(func(v T, err error) bool {
			if err != nil {
				// Yield partial chunk before error
				if len(chunk) > 0 {
					if !yield(chunk, nil) {
						return false
					}
					chunk = nil
				}
				hadError = true
				return yield(nil, err)
			}
			chunk = append(chunk, v)
			if len(chunk) >= size {
				if !yield(chunk, nil) {
					return false
				}
				chunk = make([]T, 0, size)
			}
			return true
		})
		// Yield remaining elements
		if !hadError && len(chunk) > 0 {
			yield(chunk, nil)
		}
	}
}

// FromSlice creates an iterator from a slice.
func FromSlice[T any](slice []T) Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for i := 0; i < len(slice); i++ {
			if !yield(slice[i], nil) {
				return
			}
		}
	}
}

// Empty returns an empty iterator.
func Empty[T any]() Seq2[T, error] {
	return func(yield func(T, error) bool) {}
}

// Once returns an iterator that yields a single value.
func Once[T any](v T) Seq2[T, error] {
	return func(yield func(T, error) bool) {
		yield(v, nil)
	}
}

// OnceErr returns an iterator that yields a single error.
func OnceErr[T any](err error) Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		yield(zero, err)
	}
}

// Concat concatenates multiple iterators into one.
func Concat[T any](its ...Seq2[T, error]) Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for i := 0; i < len(its); i++ {
			shouldContinue := true
			its[i](func(v T, err error) bool {
				if !yield(v, err) {
					shouldContinue = false
					return false
				}
				return true
			})
			if !shouldContinue {
				return
			}
		}
	}
}

// Flatten flattens an iterator of slices into a single iterator.
func Flatten[T any](it Seq2[[]T, error]) Seq2[T, error] {
	return func(yield func(T, error) bool) {
		it(func(slice []T, err error) bool {
			if err != nil {
				var zero T
				return yield(zero, err)
			}
			for i := 0; i < len(slice); i++ {
				if !yield(slice[i], nil) {
					return false
				}
			}
			return true
		})
	}
}
