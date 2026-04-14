package sort

import "unsafe"

// Small sort strategy type and constants.

type SmallSortStrategy int

const (
	Fallback SmallSortStrategy = iota
	General
	Network
)

const (
	SMALL_SORT_FALLBACK_THRESHOLD = 16
	SMALL_SORT_GENERAL_THRESHOLD  = 32
	SMALL_SORT_NETWORK_THRESHOLD  = 32

	SMALL_SORT_GENERAL_SCRATCH_LEN = SMALL_SORT_GENERAL_THRESHOLD + 16
	SMALL_SORT_NETWORK_SCRATCH_LEN = SMALL_SORT_NETWORK_THRESHOLD

	MAX_STACK_ARRAY_SIZE        = 4096
	PSEUDO_MEDIAN_REC_THRESHOLD = 64
)

func chooseUnstableSmallSort[T any]() SmallSortStrategy {
	var z T
	size := int(unsafe.Sizeof(z))
	if isCopyType[T]() &&
		hasEfficientSwap[T]() &&
		size*SMALL_SORT_NETWORK_SCRATCH_LEN <= MAX_STACK_ARRAY_SIZE {
		return Network
	}
	if size*SMALL_SORT_GENERAL_SCRATCH_LEN <= MAX_STACK_ARRAY_SIZE {
		return General
	}
	return Fallback
}

func smallSortThresholdFromStrategy(strategy SmallSortStrategy) int {
	switch strategy {
	case Network:
		return SMALL_SORT_NETWORK_THRESHOLD
	case General:
		return SMALL_SORT_GENERAL_THRESHOLD
	default:
		return SMALL_SORT_FALLBACK_THRESHOLD
	}
}

func isCopyType[T any]() bool {
	var z T
	switch any(z).(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		float32, float64, complex64, complex128,
		string, bool:
		return true
	default:
		return false
	}
}

func hasEfficientSwap[T any]() bool {
	var z T
	return unsafe.Sizeof(z) <= unsafe.Sizeof(uint64(0))
}

//go:noinline
func panicOnOrdViolation() {
	panic("Ord violation")
}

func selectIndex(cond bool, ifTrue, ifFalse int) int {
	if cond {
		return ifTrue
	}
	return ifFalse
}
