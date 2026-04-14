# Hướng dẫn tối ưu hóa Sort trong Go - Áp dụng kỹ thuật từ Rust ipnsort

## Tổng quan

Tài liệu này phân tích các kỹ thuật tối ưu từ implementation Rust `ipnsort` (Lukas Bergdoll) và hướng dẫn cách áp dụng tương tự trong Go.

---

## 1. Kỹ thuật `#[inline]` trong Rust → `//go:noinline` trong Go

### Rust sử dụng như thế nào?

```rust
#[inline(always)]   // Bắt buộc inline - dùng cho entry points, median3
#[inline(never)]    // Không bao giờ inline - dùng cho heapsort, panic path
// Không attribute   // Để compiler tự quyết định
```

**Chiến lược của Rust:**
- `#[inline(always)]` cho: entry points (`sort()`), hot path (`median3()`, `merge_up()`)
- `#[inline(never)]` cho: fallback algorithms (`heapsort()`), panic paths, code lớn (`sort9_optimal()`)
- **Lý do**: Tránh code bloat, tối ưu i-cache, giảm compile time

### Áp dụng trong Go

Go **KHÔNG có `//go:inline`** (ép inline), nhưng có `//go:noinline` để **ngăn** inline:

```go
//go:noinline
func heapSortRef[T any, F ~func(a, b *T) bool](v []T, less F) {
    // heapsort fallback - không inline để tránh code bloat
}

//go:noinline  
func panicOnOrdViolation() {
    panic("Ord violation")
}

// Không có directive → Go tự quyết inline
func median3[T any, F ~func(a, b *T) bool](a, b, c *T, less F) *T {
    // Go compiler sẽ tự động inline nếu có lợi
}
```

**Best practice:**
1. Dùng `//go:noinline` cho: fallback (heapsort), panic path, functions > 50 dòng
2. Để Go tự inline cho: hot path nhỏ (median3, ptr arithmetic, boolToInt)
3. Kiểm tra inline decision: `go build -gcflags="-m"`

---

## 2. "SIMD" trong Rust → Branchless code trong Go

### Rust làm gì?

Rust ipnsort **KHÔNG dùng SIMD explicit**. Thay vào đó dùng **branchless patterns** để compiler auto-vectorize:

```rust
// Branchless: boolean → integer
state.num_lt += right_is_lt as usize;  // true=1, false=0

// Branchless merge
let is_l = !is_less(&*right_src, &*left_src);
let src = if is_l { left_src } else { right_src };
ptr::copy_nonoverlapping(src, dst, 1);
```

### Go áp dụng tương tự

Go 1.20+ cải thiện auto-vectorization. Viết code branchless để tận dụng:

#### ✅ Pattern 1: Boolean arithmetic

```go
// ❌ Branchy
if less(&a, &b) {
    child++
}

// ✅ Branchless
child += boolToInt(less(&v[child], &v[child+1]))
```

#### ✅ Pattern 2: Branchless select

```go
// Tương tự Rust `select()` cho cmov
func selectInt(cond bool, ifTrue, ifFalse int) int {
    if cond {
        return ifTrue
    }
    return ifFalse
}
```

#### ✅ Pattern 3: Manual loop unrolling

```go
// 2x unroll cho types nhỏ (≤ 16 bytes)
unrollEnd := n &^ 1  // làm tròn xuống cho chẵn
for right := 0; right < unrollEnd; right += 2 {
    if less(&v[right], pivot) {
        v[numLT], v[right] = v[right], v[numLT]
        numLT++
    }
    if less(&v[right+1], pivot) {
        v[numLT], v[right+1] = v[right+1], v[numLT]
        numLT++
    }
}
```

---

## 3. Zero-Allocation Strategies

### Rust: Stack allocation với MaybeUninit

```rust
let mut stack_array = MaybeUninit::<[T; 48]>::uninit();
let scratch = unsafe {
    slice::from_raw_parts_mut(stack_array.as_mut_ptr(), 48)
};
```

### Go: Stack arrays + gap guard

#### ✅ Pattern 1: Stack arrays

```go
// Stack allocation (nếu ≤ 4KB)
var scratch [48]T
smallSortGeneralWithScratch(v, scratch[:], less)
```

**Kiểm tra escape:**
```bash
go build -gcflags="-m" | grep "escapes"
```

#### ✅ Pattern 2: GapGuard cho panic safety

```go
type gapGuard[T any] struct {
    pos   int
    value T
    has   bool
}

func (g *gapGuard[T]) Close(v []T) {
    if g.has {
        v[g.pos] = g.value
    }
}

// Usage
func partition[T any](v []T) {
    var gap gapGuard[T]
    defer gap.Close(v)  // Auto restore nếu panic
    
    // ... partition logic
    gap = gapGuard[T]{pos: right, value: v[left], has: true}
    gap.has = false  // Manual close
}
```

---

## 4. Giảm Overhead

### 4.1. Compile-time dispatch

**Rust:** `const fn` chọn implementation tại compile time

**Go:** Runtime selection (one-time cost)

```go
// Chọn MỘT lần
if unsafe.Sizeof(v[0]) <= 96 {
    partitionFunc = partitionLomutoBranchlessRef[T, F]
} else {
    partitionFunc = partitionHoareRef[T, F]
}
```

### 4.2. Pointer arithmetic giảm bounds checking

```go
// ✅ Dùng pointer thay vì slice indexing
vBase := &v[0]
left := vBase
right := (*T)(unsafe.Add(unsafe.Pointer(vBase), uintptr(n)))

// So pointer thay vì bounds check
for uintptr(unsafe.Pointer(left)) < uintptr(unsafe.Pointer(right)) {
    // ...
}
```

---

## 5. Checklist tối ưu

### ✅ ĐÃ ÁP DỤNG

| Kỹ thuật | File |
|----------|------|
| Branchless partition | `partition_fast.go` |
| Panic-safe gap guard | `partition_panicsafe.go` |
| Pointer arithmetic | `ptrutil.go` |
| Zero-alloc direct sort | `lib_ref.go` |
| Stack scratch buffer | `smallsort_ref.go` |

### 🔧 SẼ CẢI THIỆN

| Kỹ thuật | File mới | Ưu tiên |
|----------|----------|---------|
| `//go:noinline` cho heapsort | `heapsort_ref.go` | Cao |
| Branchless median3 (XOR) | `pivot_optimized.go` | Cao |
| Manual unroll 2x | `partition_optimized.go` | Cao |
| Branchless merge | `smallsort_optimized.go` | Trung bình |
| Inline control docs | `OPTIMIZATION_GUIDE.md` | Cao |

---

## 6. Benchmark commands

```bash
# Chạy tests
go test -v

# Benchmark
go test -bench=. -benchmem -count=5

# Escape analysis
go build -gcflags="-m" 2>&1 | grep "escapes"

# Inline decisions  
go build -gcflags="-m" 2>&1 | grep -E "(inlined|cannot inline)"

# Bounds check elimination
go build -gcflags="-d=ssa/check_bce/debug"
```
