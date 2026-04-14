package sort

import "unsafe"

type partitionInfoRef struct {
	numLT              int
	alreadyPartitioned bool
}

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

func quicksortHybridRef[T any, F ~func(a, b *T) bool](v []T, ancestorPivot *T, limit int, smallThreshold int, smallStrategy SmallSortStrategy, less F) {
	const (
		highlyUnbalancedRatioDiv = 8
		partialInsertionLimit    = 8
		partialInsertionMaxLen   = 2048
	)

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
			info := partitionRefInfo(v, pivotPos, less, true)
			v = v[info.numLT+1:]
			ancestorPivot = nil
			continue
		}

		info := partitionRefInfo(v, pivotPos, less, false)
		numLT := info.numLT
		left := v[:numLT]
		rightWithPivot := v[numLT:]
		pivot := &rightWithPivot[0]
		right := rightWithPivot[1:]

		lSize := len(left)
		rSize := len(right)
		highlyUnbalanced := lSize < len(v)/highlyUnbalancedRatioDiv || rSize < len(v)/highlyUnbalancedRatioDiv
		if highlyUnbalanced {
			breakPatternsRef(v, numLT)
		} else if info.alreadyPartitioned && len(v) <= partialInsertionMaxLen {
			if partialInsertionSortRef(left, 1, partialInsertionLimit, less) &&
				partialInsertionSortRef(right, 1, partialInsertionLimit, less) {
				return
			}
		}

		quicksortHybridRef(left, ancestorPivot, limit, smallThreshold, smallStrategy, less)
		v = right
		ancestorPivot = pivot
	}
}

func partitionRefInfo[T any, F ~func(a, b *T) bool](v []T, pivotIdx int, less F, reverse bool) partitionInfoRef {
	n := len(v)
	if n == 0 {
		return partitionInfoRef{}
	}
	if pivotIdx < 0 || pivotIdx >= n {
		panic("partitionRefInfo: pivot out of bounds")
	}

	v[0], v[pivotIdx] = v[pivotIdx], v[0]
	vWithoutPivot := v[1:]

	var zero T
	const maxBranchlessPartitionSize = 96
	if unsafe.Sizeof(zero) <= maxBranchlessPartitionSize {
		numLT, alreadyPartitioned := partitionLomutoBranchlessRefInfo(vWithoutPivot, &v[0], less, reverse)
		v[0], v[numLT] = v[numLT], v[0]
		return partitionInfoRef{
			numLT:              numLT,
			alreadyPartitioned: alreadyPartitioned,
		}
	}

	numLT, alreadyPartitioned := partitionHoareRefInfo(vWithoutPivot, &v[0], less, reverse)
	v[0], v[numLT] = v[numLT], v[0]
	return partitionInfoRef{
		numLT:              numLT,
		alreadyPartitioned: alreadyPartitioned,
	}
}

func breakPatternsRef[T any](v []T, pivotPos int) {
	const (
		insertionSortThreshold = 24
		nintherThreshold       = 128
	)

	size := len(v)
	if size <= insertionSortThreshold || pivotPos <= 0 || pivotPos >= size {
		return
	}

	lSize := pivotPos
	rSize := size - (pivotPos + 1)

	if lSize >= insertionSortThreshold {
		i0, i1 := 0, lSize/4
		if i0 != i1 {
			v[i0], v[i1] = v[i1], v[i0]
		}
		i2, i3 := pivotPos-1, pivotPos-lSize/4
		if i2 != i3 {
			v[i2], v[i3] = v[i3], v[i2]
		}

		if lSize > nintherThreshold {
			a, b := 1, lSize/4+1
			if a != b {
				v[a], v[b] = v[b], v[a]
			}
			c, d := 2, lSize/4+2
			if c != d {
				v[c], v[d] = v[d], v[c]
			}
			e, f := pivotPos-2, pivotPos-(lSize/4+1)
			if e != f {
				v[e], v[f] = v[f], v[e]
			}
			g, h := pivotPos-3, pivotPos-(lSize/4+2)
			if g != h {
				v[g], v[h] = v[h], v[g]
			}
		}
	}

	if rSize >= insertionSortThreshold {
		i0, i1 := pivotPos+1, pivotPos+(1+rSize/4)
		if i0 != i1 {
			v[i0], v[i1] = v[i1], v[i0]
		}
		i2, i3 := size-1, size-rSize/4
		if i2 != i3 {
			v[i2], v[i3] = v[i3], v[i2]
		}

		if rSize > nintherThreshold {
			a, b := pivotPos+2, pivotPos+(2+rSize/4)
			if a != b {
				v[a], v[b] = v[b], v[a]
			}
			c, d := pivotPos+3, pivotPos+(3+rSize/4)
			if c != d {
				v[c], v[d] = v[d], v[c]
			}
			e, f := size-2, size-(1+rSize/4)
			if e != f {
				v[e], v[f] = v[f], v[e]
			}
			g, h := size-3, size-(2+rSize/4)
			if g != h {
				v[g], v[h] = v[h], v[g]
			}
		}
	}
}
