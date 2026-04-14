package sort

import (
	"cmp"
	"slices"
	"sort"
	"sync"
	"unsafe"
)

// Sort sorts arr using the default ordering for type T.
func Sort[T cmp.Ordered](arr []T) {
	switch v := any(arr).(type) {
	case []int:
		sort.Ints(v)
		return
	case []float64:
		sort.Float64s(v)
		return
	case []string:
		sort.Strings(v)
		return
	}
	slices.Sort(arr)
}

// SortBy sorts arr using a custom comparison function.
func SortBy[T any](arr []T, compare func(a, b T) int) {
	slices.SortFunc(arr, compare)
}

// SortByRef sorts arr using a comparator that takes pointers to elements.
//
// This more closely matches Rust's `FnMut(&T, &T) -> bool` and avoids copying
// large values on each comparison.
//
// Default behavior is standardized to the direct in-place algorithm for
// predictable performance and zero extra O(n) memory.
func SortByRef[T any](arr []T, less func(a, b *T) bool) {
	unstableSortRefDirect(arr, less)
}

// SortByRefDirect forces the direct in-place algorithm (no extra O(n) memory).
func SortByRefDirect[T any](arr []T, less func(a, b *T) bool) {
	unstableSortRefDirect(arr, less)
}

// SortByRefAdaptive may choose an indirect path for very large element sizes.
//
// NOTE: The indirect path allocates O(n) memory and is not always faster.
func SortByRefAdaptive[T any](arr []T, less func(a, b *T) bool) {
	unstableSortRefAdaptive(arr, less)
}

func unstableSortRefAdaptive[T any, F ~func(a, b *T) bool](v []T, less F) {
	var zero T
	sz := unsafe.Sizeof(zero)
	if sz == 0 {
		return
	}
	if len(v) < 2 {
		return
	}

	const (
		maxLenAlwaysInsertionSort = 20
		indirectMinLen            = 4096
		indirectMinSizeBytes      = 1024
	)

	if len(v) <= maxLenAlwaysInsertionSort {
		insertionSortShiftLeftRef(v, 1, less)
		return
	}

	if sz >= indirectMinSizeBytes && len(v) >= indirectMinLen {
		sortByRefIndirect(v, less)
		return
	}

	unstableSortRefDirect(v, less)
}

func unstableSortRefDirect[T any, F ~func(a, b *T) bool](v []T, less F) {
	var zero T
	if unsafe.Sizeof(zero) == 0 {
		return
	}
	if len(v) < 2 {
		return
	}

	const maxLenAlwaysInsertionSort = 20
	if len(v) <= maxLenAlwaysInsertionSort {
		insertionSortShiftLeftRef(v, 1, less)
		return
	}

	ipnsortRef(v, less)
}

func sortByRefIndirect[T any, F ~func(a, b *T) bool](v []T, less F) {
	n := len(v)
	if n < 2 {
		return
	}

	scratch := acquireSortByRefScratch(n)
	defer releaseSortByRefScratch(scratch)

	idx := scratch.idx
	for i := range idx {
		idx[i] = i
	}

	sort.Slice(idx, func(i, j int) bool {
		return less(&v[idx[i]], &v[idx[j]])
	})

	applyPermutationFromSourcesWithVisited(v, idx, scratch.visited)
}

type sortByRefScratch struct {
	idx     []int
	visited []bool
}

var sortByRefScratchPool sync.Pool

func acquireSortByRefScratch(n int) *sortByRefScratch {
	if v := sortByRefScratchPool.Get(); v != nil {
		s := v.(*sortByRefScratch)
		if cap(s.idx) < n {
			s.idx = make([]int, n)
		}
		if cap(s.visited) < n {
			s.visited = make([]bool, n)
		}
		s.idx = s.idx[:n]
		s.visited = s.visited[:n]
		return s
	}
	return &sortByRefScratch{
		idx:     make([]int, n),
		visited: make([]bool, n),
	}
}

func releaseSortByRefScratch(s *sortByRefScratch) {
	s.idx = s.idx[:0]
	s.visited = s.visited[:0]
	sortByRefScratchPool.Put(s)
}

// applyPermutationFromSources transforms v in-place so that:
//
//	v[i] = old[v][src[i]]
func applyPermutationFromSources[T any](v []T, src []int) {
	visited := make([]bool, len(v))
	applyPermutationFromSourcesWithVisited(v, src, visited)
}

func applyPermutationFromSourcesWithVisited[T any](v []T, src []int, visited []bool) {
	n := len(v)
	if n < 2 {
		return
	}
	if len(src) != n {
		panic("applyPermutationFromSources: length mismatch")
	}
	if len(visited) < n {
		panic("applyPermutationFromSourcesWithVisited: visited too short")
	}
	visited = visited[:n]
	clear(visited)
	for i := range n {
		if visited[i] || src[i] == i {
			visited[i] = true
			continue
		}

		tmp := v[i]
		j := i
		for {
			visited[j] = true
			k := src[j]
			if k == i {
				v[j] = tmp
				break
			}
			v[j] = v[k]
			j = k
		}
	}
}

func ipnsortRef[T any, F ~func(a, b *T) bool](v []T, less F) {
	runLen, wasReversed := findExistingRunRef(v, less)
	if runLen == len(v) {
		if wasReversed {
			reverse(v)
		}
		return
	}

	strategy := chooseUnstableSmallSort[T]()
	threshold := smallSortRefThresholdFromStrategy(strategy)
	limit := 2 * ilog2(len(v)|1)
	quicksortRef(v, nil, limit, threshold, strategy, less)
}

func findExistingRunRef[T any, F ~func(a, b *T) bool](v []T, less F) (int, bool) {
	n := len(v)
	if n < 2 {
		return n, false
	}

	vBase := getBasePtr(v)
	runLen := 2
	strictlyDescending := less(ptrAdd(vBase, 1), ptrAdd(vBase, 0))
	if strictlyDescending {
		for runLen < n && less(ptrAdd(vBase, runLen), ptrAdd(vBase, runLen-1)) {
			runLen++
		}
	} else {
		for runLen < n && !less(ptrAdd(vBase, runLen), ptrAdd(vBase, runLen-1)) {
			runLen++
		}
	}
	return runLen, strictlyDescending
}

func insertionSortShiftLeftRef[T any, F ~func(a, b *T) bool](v []T, offset int, less F) {
	n := len(v)
	if offset == 0 || offset > n {
		panic("insertionSortShiftLeftRef: invalid offset")
	}

	for i := offset; i < n; i++ {
		insertTailAtRef(v, i, less)
	}
}

func partialInsertionSortRef[T any, F ~func(a, b *T) bool](v []T, offset int, moveLimit int, less F) bool {
	n := len(v)
	if n < 2 {
		return true
	}
	if offset < 1 {
		offset = 1
	}
	if offset >= n {
		return true
	}

	moves := 0
	for i := offset; i < n; i++ {
		if !less(&v[i], &v[i-1]) {
			continue
		}

		j := i
		for j > 0 && less(&v[j], &v[j-1]) {
			v[j], v[j-1] = v[j-1], v[j]
			j--
		}
		moves += i - j
		if moves > moveLimit {
			return false
		}
	}
	return true
}

// insertTailAtRef inserts v[i] into v[:i] using adjacent swaps.
//
// IMPORTANT: We must not take the address of a stack temporary and pass it to `less`,
// otherwise it will escape and allocate. This version only passes pointers to slice elements.
func insertTailAtRef[T any, F ~func(a, b *T) bool](v []T, i int, less F) {
	for j := i; j > 0 && less(&v[j], &v[j-1]); j-- {
		v[j], v[j-1] = v[j-1], v[j]
	}
}

func reverse[T any](v []T) {
	n := len(v)
	if n < 2 {
		return
	}
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		v[i], v[j] = v[j], v[i]
	}
}

func ilog2(n int) int {
	var log int
	for n > 1 {
		log++
		n >>= 1
	}
	return log
}
