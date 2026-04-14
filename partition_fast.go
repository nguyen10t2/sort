package sort

func partitionLomutoBranchlessRef[T any, F ~func(a, b *T) bool](v []T, pivot *T, less F, reverse bool) int {
	n := len(v)
	numLT := 0

	if !reverse {
		for right := range n {
			if less(&v[right], pivot) {
				v[numLT], v[right] = v[right], v[numLT]
				numLT++
			}
		}
		return numLT
	}

	// reverse: a <= pivot  <=>  !less(pivot, a)
	for right := range n {
		if !less(pivot, &v[right]) {
			v[numLT], v[right] = v[right], v[numLT]
			numLT++
		}
	}
	return numLT
}

func partitionLomutoBranchlessRefInfo[T any, F ~func(a, b *T) bool](v []T, pivot *T, less F, reverse bool) (int, bool) {
	n := len(v)
	numLT := 0
	alreadyPartitioned := true

	if !reverse {
		for right := range n {
			if less(&v[right], pivot) {
				if numLT != right {
					alreadyPartitioned = false
					v[numLT], v[right] = v[right], v[numLT]
				}
				numLT++
			}
		}
		return numLT, alreadyPartitioned
	}

	// reverse: a <= pivot  <=>  !less(pivot, a)
	for right := range n {
		if !less(pivot, &v[right]) {
			if numLT != right {
				alreadyPartitioned = false
				v[numLT], v[right] = v[right], v[numLT]
			}
			numLT++
		}
	}
	return numLT, alreadyPartitioned
}

func partitionHoareRef[T any, F ~func(a, b *T) bool](v []T, pivot *T, less F, reverse bool) int {
	n := len(v)
	if n == 0 {
		return 0
	}

	vBase := getMutBasePtr(v)
	left := vBase
	right := ptrAdd(vBase, n)

	hasGap := false
	var gapPos *T
	var gapVal T

	if !reverse {
		for {
			for ptrLess(left, right) && less(left, pivot) {
				left = ptrAdd(left, 1)
			}

			for {
				right = ptrSub(right, 1)
				if !ptrLess(left, right) || less(right, pivot) {
					break
				}
			}

			if !ptrLess(left, right) {
				break
			}

			if !hasGap {
				hasGap = true
				gapPos = right
				gapVal = *left
			} else {
				*gapPos = *left
				gapPos = right
			}

			*left = *right
			left = ptrAdd(left, 1)
		}
	} else {
		// reverse: a <= pivot  <=>  !less(pivot, a)
		for {
			for ptrLess(left, right) && !less(pivot, left) {
				left = ptrAdd(left, 1)
			}

			for {
				right = ptrSub(right, 1)
				if !ptrLess(left, right) || !less(pivot, right) {
					break
				}
			}

			if !ptrLess(left, right) {
				break
			}

			if !hasGap {
				hasGap = true
				gapPos = right
				gapVal = *left
			} else {
				*gapPos = *left
				gapPos = right
			}

			*left = *right
			left = ptrAdd(left, 1)
		}
	}

	if hasGap {
		*gapPos = gapVal
	}

	return ptrDiff(left, vBase)
}

func partitionHoareRefInfo[T any, F ~func(a, b *T) bool](v []T, pivot *T, less F, reverse bool) (int, bool) {
	n := len(v)
	if n == 0 {
		return 0, true
	}

	vBase := getMutBasePtr(v)
	left := vBase
	right := ptrAdd(vBase, n)

	hasGap := false
	var gapPos *T
	var gapVal T

	if !reverse {
		for {
			for ptrLess(left, right) && less(left, pivot) {
				left = ptrAdd(left, 1)
			}

			for {
				right = ptrSub(right, 1)
				if !ptrLess(left, right) || less(right, pivot) {
					break
				}
			}

			if !ptrLess(left, right) {
				break
			}

			if !hasGap {
				hasGap = true
				gapPos = right
				gapVal = *left
			} else {
				*gapPos = *left
				gapPos = right
			}

			*left = *right
			left = ptrAdd(left, 1)
		}
	} else {
		// reverse: a <= pivot  <=>  !less(pivot, a)
		for {
			for ptrLess(left, right) && !less(pivot, left) {
				left = ptrAdd(left, 1)
			}

			for {
				right = ptrSub(right, 1)
				if !ptrLess(left, right) || !less(pivot, right) {
					break
				}
			}

			if !ptrLess(left, right) {
				break
			}

			if !hasGap {
				hasGap = true
				gapPos = right
				gapVal = *left
			} else {
				*gapPos = *left
				gapPos = right
			}

			*left = *right
			left = ptrAdd(left, 1)
		}
	}

	if hasGap {
		*gapPos = gapVal
	}

	return ptrDiff(left, vBase), !hasGap
}
