package sort

import "unsafe"

func quicksortRef[T any, F ~func(a, b *T) bool](v []T, ancestorPivot *T, limit int, smallThreshold int, smallStrategy SmallSortStrategy, less F) {
	for {
		if len(v) <= smallThreshold {
			smallSortRefWithStrategy(v, smallStrategy, less)
			return
		}

		if limit == 0 {
			heapSortRef(v, less)
			return
		}
		limit--

		pivotPos := choosePivotRef(v, less)

		if ancestorPivot != nil && !less(ancestorPivot, ptrAdd(getBasePtr(v), pivotPos)) {
			numLT := partitionRef(v, pivotPos, less, true)
			v = v[numLT+1:]
			ancestorPivot = nil
			continue
		}

		numLT := partitionRef(v, pivotPos, less, false)

		left := v[:numLT]
		rightWithPivot := v[numLT:]
		pivot := &rightWithPivot[0]
		right := rightWithPivot[1:]

		quicksortRef(left, ancestorPivot, limit, smallThreshold, smallStrategy, less)

		v = right
		ancestorPivot = pivot
	}
}

func partitionRef[T any, F ~func(a, b *T) bool](v []T, pivotIdx int, less F, reverse bool) int {
	n := len(v)
	if n == 0 {
		return 0
	}
	if pivotIdx < 0 || pivotIdx >= n {
		panic("partitionRef: pivot out of bounds")
	}

	v[0], v[pivotIdx] = v[pivotIdx], v[0]
	vWithoutPivot := v[1:]

	var zero T
	const maxBranchlessPartitionSize = 96
	if unsafe.Sizeof(zero) <= maxBranchlessPartitionSize {
		numLT := partitionLomutoBranchlessRef(vWithoutPivot, &v[0], less, reverse)
		v[0], v[numLT] = v[numLT], v[0]
		return numLT
	}

	numLT := partitionHoareRef(vWithoutPivot, &v[0], less, reverse)
	v[0], v[numLT] = v[numLT], v[0]
	return numLT
}
