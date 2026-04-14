package sort

import "unsafe"

// Pointer arithmetic utilities for working with slices of arbitrary types.

func getBasePtr[T any](v []T) *T {
	if len(v) == 0 {
		return nil
	}
	return &v[0]
}

func getMutBasePtr[T any](v []T) *T {
	return getBasePtr(v)
}

//go:nosplit
func ptrAdd[T any](p *T, offset int) *T {
	return (*T)(unsafe.Add(unsafe.Pointer(p), uintptr(offset)*unsafe.Sizeof(*p)))
}

//go:nosplit
func ptrSub[T any](p *T, offset int) *T {
	return (*T)(unsafe.Add(unsafe.Pointer(p), uintptr(-offset)*unsafe.Sizeof(*p)))
}

func ptrDiff[T any](a, b *T) int {
	return int((uintptr(unsafe.Pointer(a)) - uintptr(unsafe.Pointer(b))) / unsafe.Sizeof(*a))
}

func ptrLess[T any](a, b *T) bool {
	return uintptr(unsafe.Pointer(a)) < uintptr(unsafe.Pointer(b))
}

func ptrRead[T any](p *T) T {
	return *p
}

func ptrCopyNonoverlapping[T any](dst, src *T, count int) {
	if count == 1 {
		*dst = *src
		return
	}
	srcSlice := unsafe.Slice(src, count)
	dstSlice := unsafe.Slice(dst, count)
	copy(dstSlice, srcSlice)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
