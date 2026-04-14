package sort

func choosePivotRef[T any, F ~func(a, b *T) bool](v []T, less F) int {
	n := len(v)
	if n < 3 {
		return n / 2
	}
	if n < 40 {
		vBase := getBasePtr(v)
		return ptrDiff(median3Ref(vBase, ptrAdd(vBase, n/2), ptrAdd(vBase, n-1), less), vBase)
	}
	return ptrDiff(nintherRef(v, less), getBasePtr(v))
}

func nintherRef[T any, F ~func(a, b *T) bool](v []T, less F) *T {
	n := len(v)
	vBase := getBasePtr(v)
	step := n / 8

	p0 := vBase
	p1 := ptrAdd(vBase, step*1)
	p2 := ptrAdd(vBase, step*3)
	p3 := ptrAdd(vBase, step*4)
	p4 := ptrAdd(vBase, step*5)
	p5 := ptrAdd(vBase, step*7)
	p6 := ptrAdd(vBase, n-3)
	p7 := ptrAdd(vBase, n-2)
	p8 := ptrAdd(vBase, n-1)

	m0 := median3Ref(p0, p1, p2, less)
	m1 := median3Ref(p3, p4, p5, less)
	m2 := median3Ref(p6, p7, p8, less)

	return median3Ref(m0, m1, m2, less)
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

// Deprecated: median3RecRef replaced by nintherRef
// func median3RecRef ... (removed for simplicity, ninther is standard)
