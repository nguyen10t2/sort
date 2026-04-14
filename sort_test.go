package sort

import (
	"math/rand"
	"slices"
	stdsort "sort"
	"strconv"
	"testing"
)

type Big struct {
	Key int
	Pad [256]byte
}

func TestSortMatchesStdSort(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	sizes := []int{0, 1, 2, 3, 7, 8, 9, 15, 16, 17, 31, 32, 33, 63, 64, 65, 127, 128, 1024}
	patterns := []string{"random", "sorted", "reversed", "duplicates", "nearly-sorted", "organ-pipe"}

	for _, size := range sizes {
		for _, pattern := range patterns {
			name := patternName(pattern, size)
			t.Run(name, func(t *testing.T) {
				input := generateInts(size, pattern, rng)

				got := append([]int(nil), input...)
				Sort(got)

				want := append([]int(nil), input...)
				stdsort.Ints(want)

				if !slices.Equal(got, want) {
					t.Fatalf("mismatch for %s", name)
				}
			})
		}
	}
}

func TestSortByRefMatchesStdSort(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))

	sizes := []int{0, 1, 2, 3, 7, 8, 9, 15, 16, 17, 31, 32, 33, 63, 64, 65, 127, 128, 1024}
	patterns := []string{"random", "sorted", "reversed", "duplicates", "nearly-sorted", "organ-pipe"}

	for _, size := range sizes {
		for _, pattern := range patterns {
			name := "ref-" + patternName(pattern, size)
			t.Run(name, func(t *testing.T) {
				input := generateInts(size, pattern, rng)

				got := append([]int(nil), input...)
				SortByRef(got, func(a, b *int) bool { return *a < *b })

				want := append([]int(nil), input...)
				stdsort.Ints(want)

				if !slices.Equal(got, want) {
					t.Fatalf("mismatch for %s", name)
				}
			})
		}
	}
}

func TestSortByCustomComparator(t *testing.T) {
	input := []int{5, 1, 9, 1, 0, 3, 3, 7}

	got := append([]int(nil), input...)
	SortBy(got, func(a, b int) int {
		switch {
		case a > b:
			return -1
		case a < b:
			return 1
		default:
			return 0
		}
	})

	want := append([]int(nil), input...)
	slices.SortFunc(want, func(a, b int) int {
		switch {
		case a > b:
			return -1
		case a < b:
			return 1
		default:
			return 0
		}
	})

	if !slices.Equal(got, want) {
		t.Fatalf("descending mismatch: got=%v want=%v", got, want)
	}
}

func TestSortByRefAvoidsCopies(t *testing.T) {
	type Big struct {
		Key int
		Pad [256]byte
	}
	input := []Big{{Key: 5}, {Key: 1}, {Key: 9}, {Key: 1}, {Key: 0}, {Key: 3}}

	got := append([]Big(nil), input...)
	SortByRef(got, func(a, b *Big) bool { return a.Key < b.Key })

	want := append([]Big(nil), input...)
	slices.SortFunc(want, func(a, b Big) int {
		switch {
		case a.Key < b.Key:
			return -1
		case a.Key > b.Key:
			return 1
		default:
			return 0
		}
	})

	if len(got) != len(want) {
		t.Fatalf("len mismatch")
	}
	for i := range got {
		if got[i].Key != want[i].Key {
			t.Fatalf("mismatch at %d: got=%d want=%d", i, got[i].Key, want[i].Key)
		}
	}
}

func BenchmarkBigStructSorters(b *testing.B) {
	benchCases := []struct {
		name    string
		n       int
		pattern string
	}{
		{name: "random-8k", n: 8_000, pattern: "random"},
		{name: "duplicates-8k", n: 8_000, pattern: "duplicates"},
		{name: "nearly-sorted-8k", n: 8_000, pattern: "nearly-sorted"},
	}

	sorters := []struct {
		name string
		fn   func([]Big)
	}{
		{
			name: "SortByRef",
			fn: func(v []Big) {
				SortByRef(v, func(a, b *Big) bool { return a.Key < b.Key })
			},
		},
		{
			name: "SortByRefDirect",
			fn: func(v []Big) {
				SortByRefDirect(v, func(a, b *Big) bool { return a.Key < b.Key })
			},
		},
		{
			name: "SortByRefIndirect",
			fn: func(v []Big) {
				SortByRefIndirect(v, func(a, b *Big) bool { return a.Key < b.Key })
			},
		},
		{
			name: "SortBy",
			fn: func(v []Big) {
				SortBy(v, func(a, b Big) int {
					switch {
					case a.Key < b.Key:
						return -1
					case a.Key > b.Key:
						return 1
					default:
						return 0
					}
				})
			},
		},
		{
			name: "slices.SortFunc",
			fn: func(v []Big) {
				slices.SortFunc(v, func(a, b Big) int {
					switch {
					case a.Key < b.Key:
						return -1
					case a.Key > b.Key:
						return 1
					default:
						return 0
					}
				})
			},
		},
		{
			name: "sort.Slice",
			fn: func(v []Big) {
				stdsort.Slice(v, func(i, j int) bool { return v[i].Key < v[j].Key })
			},
		},
	}

	rng := rand.New(rand.NewSource(2026))
	for _, bc := range benchCases {
		base := generateBigs(bc.n, bc.pattern, rng)
		b.Run(bc.name, func(b *testing.B) {
			for _, s := range sorters {
				b.Run(s.name, func(b *testing.B) {
					b.ReportAllocs()
					work := make([]Big, len(base))
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						copy(work, base)
						s.fn(work)
					}
				})
			}
		})
	}
}

func generateBigs(n int, pattern string, rng *rand.Rand) []Big {
	out := make([]Big, n)
	switch pattern {
	case "sorted":
		for i := range out {
			out[i].Key = i
			out[i].Pad[0] = byte(i)
		}
	case "reversed":
		for i := range out {
			out[i].Key = n - i
			out[i].Pad[0] = byte(i)
		}
	case "duplicates":
		for i := range out {
			k := rng.Intn(32)
			out[i].Key = k
			out[i].Pad[0] = byte(k)
		}
	case "nearly-sorted":
		for i := range out {
			out[i].Key = i
			out[i].Pad[0] = byte(i)
		}
		swaps := n / 20
		for range swaps {
			a := rng.Intn(n)
			b := rng.Intn(n)
			out[a], out[b] = out[b], out[a]
		}
	default: // random
		for i := range out {
			k := rng.Int()
			out[i].Key = k
			out[i].Pad[0] = byte(k)
		}
	}
	return out
}

func BenchmarkSortVsGoDefault(b *testing.B) {
	benchCases := []struct {
		name    string
		n       int
		pattern string
	}{
		{name: "random-1k", n: 1_000, pattern: "random"},
		{name: "random-16k", n: 16_000, pattern: "random"},
		{name: "duplicates-16k", n: 16_000, pattern: "duplicates"},
		{name: "nearly-sorted-16k", n: 16_000, pattern: "nearly-sorted"},
	}

	sorters := []struct {
		name string
		fn   func([]int)
	}{
		{name: "ipnsort", fn: func(v []int) { Sort(v) }},
		{name: "go-sort-ints", fn: func(v []int) { stdsort.Ints(v) }},
		{name: "go-slices-sort", fn: func(v []int) { slices.Sort(v) }},
	}

	rng := rand.New(rand.NewSource(99))
	for _, bc := range benchCases {
		base := generateInts(bc.n, bc.pattern, rng)
		b.Run(bc.name, func(b *testing.B) {
			for _, s := range sorters {
				b.Run(s.name, func(b *testing.B) {
					b.ReportAllocs()
					work := make([]int, len(base))
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						copy(work, base)
						s.fn(work)
					}
				})
			}
		})
	}
}

func patternName(pattern string, size int) string {
	return pattern + "-" + strconv.Itoa(size)
}

func generateInts(n int, pattern string, rng *rand.Rand) []int {
	out := make([]int, n)
	switch pattern {
	case "sorted":
		for i := range out {
			out[i] = i
		}
	case "reversed":
		for i := range out {
			out[i] = n - i
		}
	case "duplicates":
		for i := range out {
			out[i] = rng.Intn(32)
		}
	case "nearly-sorted":
		for i := range out {
			out[i] = i
		}
		swaps := n / 20
		for range swaps {
			a := rng.Intn(n)
			b := rng.Intn(n)
			out[a], out[b] = out[b], out[a]
		}
	case "organ-pipe":
		mid := n / 2
		for i := range n {
			if i <= mid {
				out[i] = i
			} else {
				out[i] = n - i
			}
		}
	default: // random
		for i := range out {
			out[i] = rng.Int()
		}
	}
	return out
}
