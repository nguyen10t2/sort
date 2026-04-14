package sort

func choosePivotRef[T any, F ~func(a, b *T) bool](v []T, less F) int {
	n := len(v)
	if n < 8 {
		panic("choosePivotRef: input too short")
	}

	vBase := getBasePtr(v)
	n8 := n / 8

	a := vBase
	b := ptrAdd(vBase, n8*4)
	c := ptrAdd(vBase, n8*7)

	if n < PSEUDO_MEDIAN_REC_THRESHOLD {
		return ptrDiff(median3Ref(a, b, c, less), vBase)
	}
	return ptrDiff(median3RecRef(a, b, c, n8, less), vBase)
}

func median3RecRef[T any, F ~func(a, b *T) bool](a, b, c *T, n int, less F) *T {
	if n*8 >= PSEUDO_MEDIAN_REC_THRESHOLD {
		n8 := n / 8
		a = median3RecRef(a, ptrAdd(a, n8*4), ptrAdd(a, n8*7), n8, less)
		b = median3RecRef(b, ptrAdd(b, n8*4), ptrAdd(b, n8*7), n8, less)
		c = median3RecRef(c, ptrAdd(c, n8*4), ptrAdd(c, n8*7), n8, less)
	}
	return median3Ref(a, b, c, less)
}

func median3Ref[T any, F ~func(a, b *T) bool](a, b, c *T, less F) *T {
	x := less(a, b)
	y := less(a, c)
	if x == y {
		z := less(b, c)
		if z != x {
			return c
		}
		return b
	}
	return a
}
