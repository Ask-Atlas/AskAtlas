---
slug: stanford-cs161-big-o-cheatsheet
title: "Big-O Cheatsheet: Asymptotic Notation and Common Complexities"
mime: application/pdf
filename: big-o-cheatsheet.pdf
course: stanford/cs161
description: "Quick reference for Big-O, Theta, and Omega notation with common complexities and recurrence solutions."
author_role: bot
---

# Big-O Cheatsheet

## Asymptotic Notation

Three notations describe how runtime scales with input size `n`:

- **Big-O** `$O(g(n))$`: upper bound. `$f(n) = O(g(n))$` iff there exist constants `$c, n_0 > 0$` such that `$0 \le f(n) \le c \cdot g(n)$` for all `$n \ge n_0$`.
- **Big-Omega** `$\Omega(g(n))$`: lower bound. `$f(n) = \Omega(g(n))$` iff `$f(n) \ge c \cdot g(n)$` eventually.
- **Big-Theta** `$\Theta(g(n))$`: tight bound. `$f(n) = \Theta(g(n))$` iff `$f(n) = O(g(n))$` and `$f(n) = \Omega(g(n))$`.

Little-o (`$o$`) and little-omega (`$\omega$`) are strict versions: `$f(n) = o(g(n))$` means `$f$` grows strictly slower than `$g$`.

## Complexity Hierarchy

From fastest to slowest growth:

| Class | Name | Example |
|-------|------|---------|
| `$O(1)$` | Constant | Array index lookup |
| `$O(\log n)$` | Logarithmic | Binary search |
| `$O(\sqrt{n})$` | Root | Primality trial division |
| `$O(n)$` | Linear | Linear scan, max finding |
| `$O(n \log n)$` | Linearithmic | Merge sort, heap sort |
| `$O(n^2)$` | Quadratic | Bubble sort, insertion sort |
| `$O(n^3)$` | Cubic | Naive matrix multiply |
| `$O(2^n)$` | Exponential | Subset enumeration |
| `$O(n!)$` | Factorial | Traveling salesman brute force |

## Common Data Structure Operations

| Structure | Access | Search | Insert | Delete |
|-----------|--------|--------|--------|--------|
| Array | `$O(1)$` | `$O(n)$` | `$O(n)$` | `$O(n)$` |
| Linked List | `$O(n)$` | `$O(n)$` | `$O(1)$` | `$O(1)$` |
| Hash Table | N/A | `$O(1)$` avg | `$O(1)$` avg | `$O(1)$` avg |
| BST (balanced) | `$O(\log n)$` | `$O(\log n)$` | `$O(\log n)$` | `$O(\log n)$` |
| Binary Heap | N/A | `$O(n)$` | `$O(\log n)$` | `$O(\log n)$` |

## Solving Recurrences

### Master Theorem

For `$T(n) = aT(n/b) + f(n)$` with `$a \ge 1$`, `$b > 1$`:

- If `$f(n) = O(n^{\log_b a - \epsilon})$`, then `$T(n) = \Theta(n^{\log_b a})$`.
- If `$f(n) = \Theta(n^{\log_b a})$`, then `$T(n) = \Theta(n^{\log_b a} \log n)$`.
- If `$f(n) = \Omega(n^{\log_b a + \epsilon})$` and regularity holds, then `$T(n) = \Theta(f(n))$`.

### Common Recurrences

- `$T(n) = 2T(n/2) + O(n) \Rightarrow \Theta(n \log n)$` (merge sort)
- `$T(n) = T(n-1) + O(1) \Rightarrow \Theta(n)$` (linear recursion)
- `$T(n) = T(n/2) + O(1) \Rightarrow \Theta(\log n)$` (binary search)
- `$T(n) = 2T(n-1) + O(1) \Rightarrow \Theta(2^n)$` (Fibonacci naive)
- `$T(n) = T(\sqrt{n}) + O(1) \Rightarrow \Theta(\log \log n)$`

## Worked Example

```python
def pair_sum(arr):
    # Outer loop: n iterations
    for i in range(len(arr)):
        # Inner loop: up to n iterations
        for j in range(i + 1, len(arr)):
            if arr[i] + arr[j] == 0:
                return True
    return False
```

Work: `$\sum_{i=0}^{n-1}(n - i - 1) = \frac{n(n-1)}{2} = \Theta(n^2)$`.

## Rules of Thumb

1. Drop constants: `$3n + 5 = O(n)$`.
2. Drop lower-order terms: `$n^2 + n = O(n^2)$`.
3. Multiplication inside loops compounds: nested loops multiply.
4. Sequential blocks add, then take the max.
5. Recursion: count calls and work per call.
