package iter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	input := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	evens := Filter(input, func(n int) bool { return n%2 == 0 })

	result, err := Collect(evens)
	require.NoError(t, err)
	assert.Equal(t, []int{2, 4, 6, 8, 10}, result)
}

func TestFilter_WithError(t *testing.T) {
	testErr := errors.New("test error")
	it := func(yield func(int, error) bool) {
		yield(1, nil)
		yield(2, nil)
		yield(0, testErr)
		yield(3, nil) // Won't reach this
	}

	filtered := Filter(it, func(n int) bool { return n%2 == 0 })
	result, err := Collect(filtered)

	require.ErrorIs(t, err, testErr)
	assert.Equal(t, []int{2}, result)
}

func TestMap(t *testing.T) {
	input := FromSlice([]int{1, 2, 3})
	doubled := Map(input, func(n int) int { return n * 2 })

	result, err := Collect(doubled)
	require.NoError(t, err)
	assert.Equal(t, []int{2, 4, 6}, result)
}

func TestMapErr(t *testing.T) {
	input := FromSlice([]int{1, 2, 3, 4, 5})
	testErr := errors.New("too big")

	// Transform with potential error
	transformed := MapErr(input, func(n int) (int, error) {
		if n > 3 {
			return 0, testErr
		}
		return n * 2, nil
	})

	result, err := Collect(transformed)
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, []int{2, 4, 6}, result)
}

func TestTake(t *testing.T) {
	input := FromSlice([]int{1, 2, 3, 4, 5})
	first3 := Take(input, 3)

	result, err := Collect(first3)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, result)
}

func TestTake_LessThanN(t *testing.T) {
	input := FromSlice([]int{1, 2})
	first5 := Take(input, 5)

	result, err := Collect(first5)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, result)
}

func TestSkip(t *testing.T) {
	input := FromSlice([]int{1, 2, 3, 4, 5})
	afterFirst3 := Skip(input, 3)

	result, err := Collect(afterFirst3)
	require.NoError(t, err)
	assert.Equal(t, []int{4, 5}, result)
}

func TestTakeWhile(t *testing.T) {
	input := FromSlice([]int{1, 2, 3, 4, 5, 1, 2})
	lessThan4 := TakeWhile(input, func(n int) bool { return n < 4 })

	result, err := Collect(lessThan4)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, result)
}

func TestSkipWhile(t *testing.T) {
	input := FromSlice([]int{1, 2, 3, 4, 5, 1, 2})
	afterLessThan4 := SkipWhile(input, func(n int) bool { return n < 4 })

	result, err := Collect(afterLessThan4)
	require.NoError(t, err)
	assert.Equal(t, []int{4, 5, 1, 2}, result)
}

func TestCollect(t *testing.T) {
	input := FromSlice([]string{"a", "b", "c"})
	result, err := Collect(input)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestCollectN(t *testing.T) {
	input := FromSlice([]int{1, 2, 3, 4, 5})
	result, err := CollectN(input, 3)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, result)
}

func TestForEach(t *testing.T) {
	input := FromSlice([]int{1, 2, 3})
	var sum int
	err := ForEach(input, func(n int) { sum += n })
	require.NoError(t, err)
	assert.Equal(t, 6, sum)
}

func TestForEachErr(t *testing.T) {
	input := FromSlice([]int{1, 2, 3, 4, 5})
	testErr := errors.New("too big")

	var processed []int
	err := ForEachErr(input, func(n int) error {
		if n > 3 {
			return testErr
		}
		processed = append(processed, n)
		return nil
	})

	require.ErrorIs(t, err, testErr)
	assert.Equal(t, []int{1, 2, 3}, processed)
}

func TestCount(t *testing.T) {
	input := FromSlice([]int{1, 2, 3, 4, 5})
	count, err := Count(input)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestFirst(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		input := FromSlice([]int{1, 2, 3})
		v, found, err := First(input)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, 1, v)
	})

	t.Run("empty", func(t *testing.T) {
		input := Empty[int]()
		v, found, err := First(input)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Zero(t, v)
	})
}

func TestLast(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		input := FromSlice([]int{1, 2, 3})
		v, found, err := Last(input)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, 3, v)
	})

	t.Run("empty", func(t *testing.T) {
		input := Empty[int]()
		v, found, err := Last(input)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Zero(t, v)
	})
}

func TestFind(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		input := FromSlice([]int{1, 2, 3, 4, 5})
		v, found, err := Find(input, func(n int) bool { return n > 3 })
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, 4, v)
	})

	t.Run("not found", func(t *testing.T) {
		input := FromSlice([]int{1, 2, 3})
		v, found, err := Find(input, func(n int) bool { return n > 10 })
		require.NoError(t, err)
		assert.False(t, found)
		assert.Zero(t, v)
	})
}

func TestAny(t *testing.T) {
	t.Run("some match", func(t *testing.T) {
		input := FromSlice([]int{1, 2, 3, 4, 5})
		result, err := Any(input, func(n int) bool { return n%2 == 0 })
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("none match", func(t *testing.T) {
		input := FromSlice([]int{1, 3, 5})
		result, err := Any(input, func(n int) bool { return n%2 == 0 })
		require.NoError(t, err)
		assert.False(t, result)
	})
}

func TestAll(t *testing.T) {
	t.Run("all match", func(t *testing.T) {
		input := FromSlice([]int{2, 4, 6})
		result, err := All(input, func(n int) bool { return n%2 == 0 })
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("some don't match", func(t *testing.T) {
		input := FromSlice([]int{2, 4, 5, 6})
		result, err := All(input, func(n int) bool { return n%2 == 0 })
		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("empty", func(t *testing.T) {
		input := Empty[int]()
		result, err := All(input, func(n int) bool { return n%2 == 0 })
		require.NoError(t, err)
		assert.True(t, result)
	})
}

func TestNone(t *testing.T) {
	t.Run("none match", func(t *testing.T) {
		input := FromSlice([]int{1, 3, 5})
		result, err := None(input, func(n int) bool { return n%2 == 0 })
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("some match", func(t *testing.T) {
		input := FromSlice([]int{1, 2, 3})
		result, err := None(input, func(n int) bool { return n%2 == 0 })
		require.NoError(t, err)
		assert.False(t, result)
	})
}

func TestEnumerate(t *testing.T) {
	input := FromSlice([]string{"a", "b", "c"})
	enumerated := Enumerate(input)

	result := make([]IndexedValue[string], 0, 3)
	for v, err := range enumerated {
		require.NoError(t, err)
		result = append(result, v)
	}

	assert.Equal(t, []IndexedValue[string]{
		{Index: 0, Value: "a"},
		{Index: 1, Value: "b"},
		{Index: 2, Value: "c"},
	}, result)
}

func TestReduce(t *testing.T) {
	input := FromSlice([]int{1, 2, 3, 4, 5})
	sum, err := Reduce(input, 0, func(acc, n int) int { return acc + n })
	require.NoError(t, err)
	assert.Equal(t, 15, sum)
}

func TestChunk(t *testing.T) {
	t.Run("even chunks", func(t *testing.T) {
		input := FromSlice([]int{1, 2, 3, 4, 5, 6})
		chunked := Chunk(input, 2)

		result, err := Collect(chunked)
		require.NoError(t, err)
		assert.Equal(t, [][]int{{1, 2}, {3, 4}, {5, 6}}, result)
	})

	t.Run("uneven chunks", func(t *testing.T) {
		input := FromSlice([]int{1, 2, 3, 4, 5})
		chunked := Chunk(input, 2)

		result, err := Collect(chunked)
		require.NoError(t, err)
		assert.Equal(t, [][]int{{1, 2}, {3, 4}, {5}}, result)
	})
}

func TestFromSlice(t *testing.T) {
	slice := []int{1, 2, 3}
	result, err := Collect(FromSlice(slice))
	require.NoError(t, err)
	assert.Equal(t, slice, result)
}

func TestEmpty(t *testing.T) {
	result, err := Collect(Empty[int]())
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestOnce(t *testing.T) {
	result, err := Collect(Once(42))
	require.NoError(t, err)
	assert.Equal(t, []int{42}, result)
}

func TestOnceErr(t *testing.T) {
	testErr := errors.New("test error")
	result, err := Collect(OnceErr[int](testErr))
	require.ErrorIs(t, err, testErr)
	assert.Empty(t, result)
}

func TestConcat(t *testing.T) {
	a := FromSlice([]int{1, 2})
	b := FromSlice([]int{3, 4})
	c := FromSlice([]int{5})

	result, err := Collect(Concat(a, b, c))
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, result)
}

func TestFlatten(t *testing.T) {
	input := FromSlice([][]int{{1, 2}, {3, 4, 5}, {6}})
	flattened := Flatten(input)

	result, err := Collect(flattened)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, result)
}

func TestChainedOperations(t *testing.T) {
	// Test chaining multiple operations
	input := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	// Get even numbers, double them, take first 3
	result, err := Collect(
		Take(
			Map(
				Filter(input, func(n int) bool { return n%2 == 0 }),
				func(n int) int { return n * 2 },
			),
			3,
		),
	)

	require.NoError(t, err)
	assert.Equal(t, []int{4, 8, 12}, result) // 2*2=4, 4*2=8, 6*2=12
}

func TestEarlyTermination(t *testing.T) {
	// Test that iterators properly stop when break is called
	callCount := 0
	input := func(yield func(int, error) bool) {
		for i := 1; i <= 100; i++ {
			callCount++
			if !yield(i, nil) {
				return
			}
		}
	}

	// Take only first 5
	result, err := Collect(Take(input, 5))
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, result)
	assert.Equal(t, 5, callCount, "iterator should stop after 5 elements")
}

// Error propagation tests

func errAfter(n int, errVal error) Seq2[int, error] {
	return func(yield func(int, error) bool) {
		for i := 0; i < n; i++ {
			if !yield(i, nil) {
				return
			}
		}
		yield(0, errVal)
	}
}

func TestMap_WithSourceError(t *testing.T) {
	testErr := errors.New("source error")
	it := errAfter(2, testErr)
	mapped := Map(it, func(n int) int { return n * 2 })

	result, err := Collect(mapped)
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, []int{0, 2}, result)
}

func TestMapErr_WithSourceError(t *testing.T) {
	testErr := errors.New("source error")
	it := errAfter(2, testErr)
	mapped := MapErr(it, func(n int) (int, error) { return n * 2, nil })

	result, err := Collect(mapped)
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, []int{0, 2}, result)
}

func TestTake_WithErrorInFirstN(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(1, testErr) // yields 0, then error
	taken := Take(it, 5)

	result, err := Collect(taken)
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, []int{0}, result)
}

func TestSkip_WithErrorDuringSkip(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(1, testErr) // yields 0, then error
	skipped := Skip(it, 5)    // trying to skip 5, but error after 1

	_, err := Collect(skipped)
	require.ErrorIs(t, err, testErr)
}

func TestTakeWhile_WithError(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(2, testErr)
	taken := TakeWhile(it, func(n int) bool { return true })

	result, err := Collect(taken)
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, []int{0, 1}, result)
}

func TestSkipWhile_WithErrorDuringSkip(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(2, testErr)
	skipped := SkipWhile(it, func(n int) bool { return true }) // skip everything

	_, err := Collect(skipped)
	require.ErrorIs(t, err, testErr)
}

func TestCount_WithError(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(3, testErr)

	count, err := Count(it)
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, 3, count)
}

func TestFirst_WithError(t *testing.T) {
	testErr := errors.New("error")
	it := OnceErr[int](testErr)

	_, found, err := First(it)
	require.ErrorIs(t, err, testErr)
	assert.False(t, found)
}

func TestLast_WithErrorMidIteration(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(2, testErr)

	v, found, err := Last(it)
	require.ErrorIs(t, err, testErr)
	assert.True(t, found)
	assert.Equal(t, 1, v) // Last valid value before error
}

func TestFind_WithErrorBeforeMatch(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(2, testErr)

	_, found, err := Find(it, func(n int) bool { return n == 99 })
	require.ErrorIs(t, err, testErr)
	assert.False(t, found)
}

func TestAny_WithError(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(2, testErr)

	result, err := Any(it, func(n int) bool { return n == 99 })
	require.ErrorIs(t, err, testErr)
	assert.False(t, result)
}

func TestAll_WithError(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(2, testErr)

	result, err := All(it, func(n int) bool { return true })
	require.ErrorIs(t, err, testErr)
	assert.False(t, result)
}

func TestNone_WithError(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(2, testErr)

	result, err := None(it, func(n int) bool { return false })
	require.ErrorIs(t, err, testErr)
	assert.False(t, result)
}

func TestEnumerate_WithError(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(1, testErr)
	enumerated := Enumerate(it)

	var results []IndexedValue[int]
	var collectedErr error
	for v, err := range enumerated {
		if err != nil {
			collectedErr = err
			break
		}
		results = append(results, v)
	}
	require.ErrorIs(t, collectedErr, testErr)
	assert.Len(t, results, 1)
}

func TestReduce_WithError(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(3, testErr)

	sum, err := Reduce(it, 0, func(acc, n int) int { return acc + n })
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, 3, sum) // 0+1+2
}

func TestChunk_WithError(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(3, testErr)
	chunked := Chunk(it, 2)

	result, err := Collect(chunked)
	require.ErrorIs(t, err, testErr)
	// Should have yielded [0,1] as a full chunk, then [2] as partial before error
	assert.Len(t, result, 2)
}

func TestFlatten_WithErrorInOuter(t *testing.T) {
	testErr := errors.New("error")
	// Create an iterator of slices that errors after first
	outer := func(yield func([]int, error) bool) {
		if !yield([]int{1, 2}, nil) {
			return
		}
		yield(nil, testErr)
	}
	flattened := Flatten(outer)

	result, err := Collect(flattened)
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, []int{1, 2}, result)
}

func TestConcat_WithErrorInFirst(t *testing.T) {
	testErr := errors.New("error")
	first := errAfter(1, testErr)
	second := FromSlice([]int{10, 20})

	result, err := Collect(Concat(first, second))
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, []int{0}, result)
}

func TestConcat_WithErrorInSecond(t *testing.T) {
	testErr := errors.New("error")
	first := FromSlice([]int{1, 2})
	second := errAfter(0, testErr) // error immediately

	result, err := Collect(Concat(first, second))
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, []int{1, 2}, result)
}

func TestForEach_WithSourceError(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(2, testErr)

	var sum int
	err := ForEach(it, func(n int) { sum += n })
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, 1, sum) // 0+1
}

func TestForEachErr_WithSourceError(t *testing.T) {
	testErr := errors.New("error")
	it := errAfter(2, testErr)

	var sum int
	err := ForEachErr(it, func(n int) error {
		sum += n
		return nil
	})
	require.ErrorIs(t, err, testErr)
	assert.Equal(t, 1, sum)
}

func TestCollect_Empty(t *testing.T) {
	result, err := Collect(Empty[int]())
	require.NoError(t, err)
	assert.Nil(t, result)
}
