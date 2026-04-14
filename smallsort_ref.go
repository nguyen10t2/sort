package sort

import "unsafe"

func smallSortRefThresholdFromStrategy(strategy SmallSortStrategy) int {
	return smallSortThresholdFromStrategy(strategy)
}

func smallSortRefWithStrategy[T any, F ~func(a, b *T) bool](v []T, strategy SmallSortStrategy, less F) {
	switch strategy {
	case Network:
		smallSortNetworkRef(v, less)
	case General:
		smallSortGeneralRef(v, less)
	default:
		if len(v) >= 2 {
			insertionSortShiftLeftRef(v, 1, less)
		}
	}
}

func smallSortGeneralRef[T any, F ~func(a, b *T) bool](v []T, less F) {
	var scratch [SMALL_SORT_GENERAL_SCRATCH_LEN]T
	smallSortGeneralWithScratchRef(v, scratch[:], less)
}

func smallSortGeneralWithScratchRef[T any, F ~func(a, b *T) bool](v []T, scratch []T, less F) {
	n := len(v)
	if n < 2 {
		return
	}
	if len(scratch) < n+16 {
		panic("smallSortGeneralWithScratchRef: scratch too short")
	}

	mid := n / 2
	presorted := 1
	var zero T
	if unsafe.Sizeof(zero) <= 16 && n >= 16 {
		sort8StableRef(v[:8], scratch[:8], scratch[n:n+8], less)
		sort8StableRef(v[mid:mid+8], scratch[mid:mid+8], scratch[n+8:n+16], less)
		presorted = 8
	} else if n >= 8 {
		sort4StableRef(v[:4], scratch[:4], less)
		sort4StableRef(v[mid:mid+4], scratch[mid:mid+4], less)
		presorted = 4
	} else {
		scratch[0] = v[0]
		scratch[mid] = v[mid]
	}

	leftRegion := scratch[:mid]
	for i := presorted; i < mid; i++ {
		leftRegion[i] = v[i]
		insertTailAtRef(leftRegion, i, less)
	}

	rightLen := n - mid
	rightRegion := scratch[mid:n]
	for i := presorted; i < rightLen; i++ {
		rightRegion[i] = v[mid+i]
		insertTailAtRef(rightRegion, i, less)
	}

	bidirectionalMergeRef(scratch[:mid], scratch[mid:n], v, less)
}

func smallSortNetworkRef[T any, F ~func(a, b *T) bool](v []T, less F) {
	n := len(v)
	if n < 2 {
		return
	}
	if n > SMALL_SORT_NETWORK_SCRATCH_LEN {
		panic("smallSortNetworkRef: input too large")
	}

	if n < 18 {
		presorted := 1
		if n >= 13 {
			sort13OptimalRef(v, less)
			presorted = 13
		} else if n >= 9 {
			sort9OptimalRef(v, less)
			presorted = 9
		}
		if presorted < n {
			insertionSortShiftLeftRef(v, presorted, less)
		}
		return
	}

	mid := n / 2
	left := v[:mid]
	right := v[mid:]

	presortRegionRef(left, less)
	presortRegionRef(right, less)

	var scratch [SMALL_SORT_NETWORK_SCRATCH_LEN]T
	bidirectionalMergeRef(left, right, scratch[:n], less)
	copy(v, scratch[:n])
}

func presortRegionRef[T any, F ~func(a, b *T) bool](region []T, less F) {
	presorted := 1
	if len(region) >= 13 {
		sort13OptimalRef(region, less)
		presorted = 13
	} else if len(region) >= 9 {
		sort9OptimalRef(region, less)
		presorted = 9
	}
	if presorted < len(region) {
		insertionSortShiftLeftRef(region, presorted, less)
	}
}

func sort4StableRef[T any, F ~func(a, b *T) bool](v, dst []T, less F) {
	if len(v) < 4 || len(dst) < 4 {
		panic("sort4StableRef: input too short")
	}

	c1 := less(&v[1], &v[0])
	c2 := less(&v[3], &v[2])
	a, b := 0, 1
	c, d := 2, 3
	if c1 {
		a, b = 1, 0
	}
	if c2 {
		c, d = 3, 2
	}

	c3 := less(&v[c], &v[a])
	c4 := less(&v[d], &v[b])
	minIdx := selectIndex(c3, c, a)
	maxIdx := selectIndex(c4, b, d)
	unknownLeft := selectIndex(c3, a, selectIndex(c4, c, b))
	unknownRight := selectIndex(c4, d, selectIndex(c3, b, c))

	c5 := less(&v[unknownRight], &v[unknownLeft])
	lo := selectIndex(c5, unknownRight, unknownLeft)
	hi := selectIndex(c5, unknownLeft, unknownRight)

	dst[0] = v[minIdx]
	dst[1] = v[lo]
	dst[2] = v[hi]
	dst[3] = v[maxIdx]
}

func sort8StableRef[T any, F ~func(a, b *T) bool](v, dst, scratch []T, less F) {
	if len(v) < 8 || len(dst) < 8 || len(scratch) < 8 {
		panic("sort8StableRef: input too short")
	}

	sort4StableRef(v[:4], scratch[:4], less)
	sort4StableRef(v[4:8], scratch[4:8], less)
	bidirectionalMergeRef(scratch[:4], scratch[4:8], dst[:8], less)
}

func swapIfLessRef[T any, F ~func(a, b *T) bool](v []T, aPos, bPos int, less F) {
	if less(&v[bPos], &v[aPos]) {
		v[aPos], v[bPos] = v[bPos], v[aPos]
	}
}

func sort9OptimalRef[T any, F ~func(a, b *T) bool](v []T, less F) {
	if len(v) < 9 {
		panic("sort9OptimalRef: input too short")
	}

	swapIfLessRef(v, 0, 3, less)
	swapIfLessRef(v, 1, 7, less)
	swapIfLessRef(v, 2, 5, less)
	swapIfLessRef(v, 4, 8, less)
	swapIfLessRef(v, 0, 7, less)
	swapIfLessRef(v, 2, 4, less)
	swapIfLessRef(v, 3, 8, less)
	swapIfLessRef(v, 5, 6, less)
	swapIfLessRef(v, 0, 2, less)
	swapIfLessRef(v, 1, 3, less)
	swapIfLessRef(v, 4, 5, less)
	swapIfLessRef(v, 7, 8, less)
	swapIfLessRef(v, 1, 4, less)
	swapIfLessRef(v, 3, 6, less)
	swapIfLessRef(v, 5, 7, less)
	swapIfLessRef(v, 0, 1, less)
	swapIfLessRef(v, 2, 4, less)
	swapIfLessRef(v, 3, 5, less)
	swapIfLessRef(v, 6, 8, less)
	swapIfLessRef(v, 2, 3, less)
	swapIfLessRef(v, 4, 5, less)
	swapIfLessRef(v, 6, 7, less)
	swapIfLessRef(v, 1, 2, less)
	swapIfLessRef(v, 3, 4, less)
	swapIfLessRef(v, 5, 6, less)
}

func sort13OptimalRef[T any, F ~func(a, b *T) bool](v []T, less F) {
	if len(v) < 13 {
		panic("sort13OptimalRef: input too short")
	}

	swapIfLessRef(v, 0, 12, less)
	swapIfLessRef(v, 1, 10, less)
	swapIfLessRef(v, 2, 9, less)
	swapIfLessRef(v, 3, 7, less)
	swapIfLessRef(v, 5, 11, less)
	swapIfLessRef(v, 6, 8, less)
	swapIfLessRef(v, 1, 6, less)
	swapIfLessRef(v, 2, 3, less)
	swapIfLessRef(v, 4, 11, less)
	swapIfLessRef(v, 7, 9, less)
	swapIfLessRef(v, 8, 10, less)
	swapIfLessRef(v, 0, 4, less)
	swapIfLessRef(v, 1, 2, less)
	swapIfLessRef(v, 3, 6, less)
	swapIfLessRef(v, 7, 8, less)
	swapIfLessRef(v, 9, 10, less)
	swapIfLessRef(v, 11, 12, less)
	swapIfLessRef(v, 4, 6, less)
	swapIfLessRef(v, 5, 9, less)
	swapIfLessRef(v, 8, 11, less)
	swapIfLessRef(v, 10, 12, less)
	swapIfLessRef(v, 0, 5, less)
	swapIfLessRef(v, 3, 8, less)
	swapIfLessRef(v, 4, 7, less)
	swapIfLessRef(v, 6, 11, less)
	swapIfLessRef(v, 9, 10, less)
	swapIfLessRef(v, 0, 1, less)
	swapIfLessRef(v, 2, 5, less)
	swapIfLessRef(v, 6, 9, less)
	swapIfLessRef(v, 7, 8, less)
	swapIfLessRef(v, 10, 11, less)
	swapIfLessRef(v, 1, 3, less)
	swapIfLessRef(v, 2, 4, less)
	swapIfLessRef(v, 5, 6, less)
	swapIfLessRef(v, 9, 10, less)
	swapIfLessRef(v, 1, 2, less)
	swapIfLessRef(v, 3, 4, less)
	swapIfLessRef(v, 5, 7, less)
	swapIfLessRef(v, 6, 8, less)
	swapIfLessRef(v, 2, 3, less)
	swapIfLessRef(v, 4, 5, less)
	swapIfLessRef(v, 6, 7, less)
	swapIfLessRef(v, 8, 9, less)
	swapIfLessRef(v, 3, 4, less)
	swapIfLessRef(v, 5, 6, less)
}

func bidirectionalMergeRef[T any, F ~func(a, b *T) bool](left, right, dst []T, less F) {
	if len(dst) != len(left)+len(right) {
		panic("bidirectionalMergeRef: invalid destination length")
	}
	if len(dst) == 0 {
		return
	}

	leftLo, rightLo, dstLo := 0, 0, 0
	leftHi, rightHi, dstHi := len(left)-1, len(right)-1, len(dst)-1

	for i := 0; i < len(dst)/2; i++ {
		leftLo, rightLo, dstLo = mergeUpRef(left, right, dst, leftLo, rightLo, dstLo, less)
		leftHi, rightHi, dstHi = mergeDownRef(left, right, dst, leftHi, rightHi, dstHi, less)
	}

	if len(dst)%2 != 0 {
		leftNonEmpty := leftLo <= leftHi
		rightNonEmpty := rightLo <= rightHi
		if leftNonEmpty && (!rightNonEmpty || !less(&right[rightLo], &left[leftLo])) {
			dst[dstLo] = left[leftLo]
			leftLo++
		} else {
			dst[dstLo] = right[rightLo]
			rightLo++
		}
		dstLo++
	}

	if leftLo != leftHi+1 || rightLo != rightHi+1 {
		panicOnOrdViolation()
	}
}

func mergeUpRef[T any, F ~func(a, b *T) bool](left, right, dst []T, leftSrc, rightSrc, out int, less F) (int, int, int) {
	leftExhausted := leftSrc >= len(left)
	rightExhausted := rightSrc >= len(right)
	if leftExhausted {
		dst[out] = right[rightSrc]
		return leftSrc, rightSrc + 1, out + 1
	}
	if rightExhausted {
		dst[out] = left[leftSrc]
		return leftSrc + 1, rightSrc, out + 1
	}

	if !less(&right[rightSrc], &left[leftSrc]) {
		dst[out] = left[leftSrc]
		return leftSrc + 1, rightSrc, out + 1
	}
	dst[out] = right[rightSrc]
	return leftSrc, rightSrc + 1, out + 1
}

func mergeDownRef[T any, F ~func(a, b *T) bool](left, right, dst []T, leftSrc, rightSrc, out int, less F) (int, int, int) {
	leftExhausted := leftSrc < 0
	rightExhausted := rightSrc < 0
	if leftExhausted {
		dst[out] = right[rightSrc]
		return leftSrc, rightSrc - 1, out - 1
	}
	if rightExhausted {
		dst[out] = left[leftSrc]
		return leftSrc - 1, rightSrc, out - 1
	}

	if !less(&right[rightSrc], &left[leftSrc]) {
		dst[out] = right[rightSrc]
		return leftSrc, rightSrc - 1, out - 1
	}
	dst[out] = left[leftSrc]
	return leftSrc - 1, rightSrc, out - 1
}
