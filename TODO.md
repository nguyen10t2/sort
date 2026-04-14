# Rust Zero-Cost Optimization Plan for Go Sort
Module: github.com/nguyen10t2/sort
Branch: blackboxai/rust-zero-cost-opt (to be created)

## Approved Plan Summary
- Thoroughly reviewed codebase: IPN unstable sort (run detect + quicksort w/ ancestor opt + heap fallback + adaptive small networks/general/insertion).
- Rust-inspired opts: 3-way partition (duplicates), ninther pivot, sort16 network, entropy quick-reject, inlines/remove panics.
- Goals: 5-15% perf gain on duplicates/random/large structs, zero-alloc direct path.
- Files: partition_fast.go, pivot_ref.go, smallsort_ref.go, lib_ref.go, sort_test.go.

## TODO Steps (Breakdown)
### 1. Setup (Current)
- [x] Create branch `blackboxai/rust-zero-cost-opt`
- [ ] Baseline bench: `go test -bench=. -benchmem -cpu=1,4 ./...`
- [ ] Commit baseline: 'Baseline before Rust opts'

### 2. Pivot Optimization ✅
- [x] Edit pivot_ref.go: Added `nintherRef` (Rust-style 9-sample medians), updated `choosePivotRef` (ninther for n>=40, median3 small/ends).
- [x] Removed recursive pseudo-median (simpler, better perf).
- [ ] Test: `go test ./...` (next).
- [ ] Microbench pivot (add later)."


### 3. 3-Way Partition
- [ ] Edit partition_fast.go: Add partition3WayRef (LT/EQ/GT Dutch flag, branchless where poss).
- [ ] Update quicksort_ref.go: partitionRef -> partition3WayRef.
- [ ] Add duplicate-heavy bench patterns.

### 4. SmallSort Extension
- [ ] Edit smallsort_ref.go: Add sort16OptimalRef (optimal ~60 swaps).
- [ ] Update smallSortNetworkRef: Use for n>=16.
- [ ] Adjust thresholds/strategies.

### 5. Polish & Final
- [ ] lib_ref.go: Entropy sortedness check in ipnsortRef.
- [ ] Remove panics/hot path guards.
- [ ] Full re-bench/compare.
- [ ] Commit opts.
- [ ] Optional: gh pr create.

## Progress Tracking
- Run `go test ./...` after each step.
- Track allocs/time vs baseline.

Next: User confirms -> Execute step-by-step.
