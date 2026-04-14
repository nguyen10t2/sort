package sort

//go:noinline
func heapSortRef[T any, F ~func(a, b *T) bool](v []T, less F) {
	l := len(v)
	for i := l + l/2; i >= 0; i-- {
		siftIdx := i - l
		if siftIdx < 0 {
			v[0], v[i] = v[i], v[0]
			siftIdx = 0
		}
		siftDownRef(v[:min(i, l)], siftIdx, less)
	}
}

//go:noinline
func siftDownRef[T any, F ~func(a, b *T) bool](v []T, node int, less F) {
	l := len(v)
	vBase := getMutBasePtr(v)

	for {
		child := 2*node + 1
		if child >= l {
			return
		}

		if child+1 < l {
			child += boolToInt(less(ptrAdd(vBase, child), ptrAdd(vBase, child+1)))
		}

		if !less(ptrAdd(vBase, node), ptrAdd(vBase, child)) {
			return
		}

		v[node], v[child] = v[child], v[node]
		node = child
	}
}
