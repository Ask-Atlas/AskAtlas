---
slug: stanford-cs161-divide-and-conquer-slides
title: "Divide and Conquer: Merge Sort, Quicksort, and Karatsuba"
mime: application/vnd.openxmlformats-officedocument.presentationml.presentation
filename: divide-and-conquer-slides.pptx
course: stanford/cs161
description: "Slide deck on divide and conquer with merge sort, quicksort pivot analysis, and Karatsuba integer multiplication."
author_role: bot
---

# Divide and Conquer

## The D&C Template

Three steps:

1. **Divide** the problem into smaller subproblems of the same form.
2. **Conquer** each subproblem recursively.
3. **Combine** subsolutions into a final answer.

## Why It Works

- Reduces problem size quickly, often geometrically.
- Yields clean recurrences solvable by the master theorem.
- Enables parallelism — subproblems are independent.

## Merge Sort: Setup

Input: array `$A$` of length `$n$`. Output: sorted array.

Divide at the midpoint, sort each half recursively, merge in linear time.

## Merge Sort: Pseudocode

```python
def merge_sort(arr):
    if len(arr) <= 1:
        return arr
    mid = len(arr) // 2
    left = merge_sort(arr[:mid])
    right = merge_sort(arr[mid:])
    return merge(left, right)

def merge(left, right):
    out, i, j = [], 0, 0
    while i < len(left) and j < len(right):
        if left[i] <= right[j]:
            out.append(left[i]); i += 1
        else:
            out.append(right[j]); j += 1
    out.extend(left[i:]); out.extend(right[j:])
    return out
```

## Merge Sort: Analysis

Recurrence: `$T(n) = 2T(n/2) + \Theta(n)$`.

By master theorem case 2: `$T(n) = \Theta(n \log n)$`.

## Merge Sort: Properties

- **Stable**: equal elements keep relative order.
- **Not in-place**: requires `$O(n)$` auxiliary memory.
- **Worst case**: `$\Theta(n \log n)$` — no bad inputs.

## Quicksort: Setup

Pick a pivot, partition so smaller elements go left and larger go right, recurse on each side.

## Quicksort: Pseudocode

```python
def quicksort(arr, lo=0, hi=None):
    if hi is None:
        hi = len(arr) - 1
    if lo < hi:
        p = partition(arr, lo, hi)
        quicksort(arr, lo, p - 1)
        quicksort(arr, p + 1, hi)

def partition(arr, lo, hi):
    pivot = arr[hi]
    i = lo - 1
    for j in range(lo, hi):
        if arr[j] <= pivot:
            i += 1
            arr[i], arr[j] = arr[j], arr[i]
    arr[i + 1], arr[hi] = arr[hi], arr[i + 1]
    return i + 1
```

## Quicksort: Balanced Pivot

If every pivot lands near the median: `$T(n) = 2T(n/2) + \Theta(n) = \Theta(n \log n)$`.

## Quicksort: Worst Case

Always picking the smallest or largest element: `$T(n) = T(n-1) + \Theta(n) = \Theta(n^2)$`.

Mitigation: randomized pivot selection or median-of-three.

## Quicksort: Expected Analysis

With a random pivot, expected comparisons: `$\Theta(n \log n)$`. The probability of picking a pivot between the 25th and 75th percentile is `$1/2$`, so depth stays `$O(\log n)$` w.h.p.

## Quicksort vs Merge Sort

| Property | Merge Sort | Quicksort |
|----------|------------|-----------|
| Worst case | `$\Theta(n \log n)$` | `$\Theta(n^2)$` |
| Avg case | `$\Theta(n \log n)$` | `$\Theta(n \log n)$` |
| Space | `$O(n)$` | `$O(\log n)$` stack |
| Stable | Yes | No (typically) |
| In-place | No | Yes |

## Karatsuba: Motivation

Multiplying two `$n$`-digit numbers naively costs `$\Theta(n^2)$`. Can we do better?

## Karatsuba: The Trick

Split each number: `$x = x_1 \cdot 10^{n/2} + x_0$`, `$y = y_1 \cdot 10^{n/2} + y_0$`.

Product: `$xy = x_1 y_1 \cdot 10^n + (x_1 y_0 + x_0 y_1) \cdot 10^{n/2} + x_0 y_0$`.

Naive: 4 recursive multiplications. Karatsuba uses 3:

- `$A = x_1 y_1$`
- `$B = x_0 y_0$`
- `$C = (x_1 + x_0)(y_1 + y_0) - A - B$`

## Karatsuba: Recurrence

`$T(n) = 3T(n/2) + \Theta(n) = \Theta(n^{\log_2 3}) \approx \Theta(n^{1.585})$`.

## Takeaways

- D&C recurrences are solved by the master theorem.
- The combine step often dominates; optimizing it (merge, partition) is key.
- Randomization turns worst cases into expected cases.
- Clever algebraic identities (Karatsuba, Strassen) beat naive bounds.
