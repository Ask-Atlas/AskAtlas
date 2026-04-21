---
slug: cs161-divide-and-conquer
course:
  ipeds_id: "243744"
  department: "CS"
  number: "161"
title: "Divide and Conquer — CS 161"
description: "Mergesort, Karatsuba multiplication, closest-pair, and the divide-conquer-combine template."
tags: ["algorithms", "divide-and-conquer", "mergesort", "recursion", "midterm"]
author_role: bot
attached_files:
  - stanford-cs161-divide-and-conquer-slides
  - stanford-cs161-master-theorem-quickref
attached_resources: []
---

# Divide and Conquer

The divide-and-conquer paradigm is the engine behind mergesort, fast Fourier transforms, integer multiplication, matrix multiplication, and a long list of geometric algorithms. The template is always the same:

1. **Divide** the input into smaller subproblems.
2. **Conquer** each subproblem recursively.
3. **Combine** the results.

The "magic" is that the divide and combine steps are usually cheap ($O(n)$ or $O(n \log n)$), and recursion compresses the work by a factor of $b$ at each level. The Master Theorem ({{GUIDE:cs161-master-theorem}}) then turns the recurrence into a closed-form running time.

## Mergesort revisited

Mergesort is the canonical example. Pseudocode:

```python
def mergesort(arr):
    n = len(arr)
    if n <= 1:
        return arr
    mid = n // 2
    left  = mergesort(arr[:mid])    # Conquer
    right = mergesort(arr[mid:])    # Conquer
    return merge(left, right)       # Combine

def merge(a, b):
    out = []
    i = j = 0
    while i < len(a) and j < len(b):
        if a[i] <= b[j]:
            out.append(a[i]); i += 1
        else:
            out.append(b[j]); j += 1
    out.extend(a[i:])
    out.extend(b[j:])
    return out
```

Recurrence: $T(n) = 2T(n/2) + \Theta(n)$. Master Theorem Case 2 with $k = 0$: $T(n) = \Theta(n \log n)$.

Mergesort's defining property is that it is **stable** (equal keys preserve their input order) and **comparison-based**. Its main downside is the $\Theta(n)$ extra space for the merge buffer.

## Karatsuba multiplication

Naive $n$-digit integer multiplication takes $\Theta(n^2)$ digit-multiplications. Karatsuba (1960) reduces this to $\Theta(n^{\log_2 3}) \approx \Theta(n^{1.585})$.

Split each $n$-digit integer $x$ into halves: $x = x_1 \cdot 10^{n/2} + x_0$. Naively, $x \cdot y$ requires four products: $x_1 y_1, x_1 y_0, x_0 y_1, x_0 y_0$. Karatsuba's trick uses three:

$$
P_1 = x_1 y_1, \quad P_2 = x_0 y_0, \quad P_3 = (x_1 + x_0)(y_1 + y_0).
$$

Then $x_1 y_0 + x_0 y_1 = P_3 - P_1 - P_2$, and the final answer assembles in $\Theta(n)$ additions/shifts.

Recurrence: $T(n) = 3T(n/2) + \Theta(n)$. Master Theorem: $n^{\log_2 3} \approx n^{1.585}$, $f(n) = n$, polynomially smaller, **Case 1** — $T(n) = \Theta(n^{\log_2 3})$.

## Closest pair of points

Given $n$ points in the plane, find the two closest. Naive comparison is $\Theta(n^2)$. Divide and conquer gets $\Theta(n \log n)$:

1. Sort points by $x$-coordinate (one-time $\Theta(n \log n)$).
2. Recurse on left and right halves; let $\delta$ be the smaller of the two minimum distances.
3. Examine the **strip** of width $2\delta$ around the dividing $x$-coordinate. A geometric argument shows that for any point in this strip, only its 7 nearest $y$-neighbors in the strip can possibly be closer than $\delta$.

Recurrence: $T(n) = 2T(n/2) + \Theta(n)$. Same as mergesort: $\Theta(n \log n)$.

The clever step is the **constant 7** — without it, the combine step would be $\Theta(n^2)$ in the worst case and the recurrence would be $T(n) = 2T(n/2) + n^2$, giving $\Theta(n^2)$.

## When divide-and-conquer is the wrong tool

D&C shines when the divide is balanced and the combine is cheap. It struggles when:

- The combine step is intrinsically expensive (e.g., merging two arbitrary heaps).
- There's no natural way to split the input into independent subproblems (e.g., shortest paths in a graph — subproblems share state).
- The subproblems overlap so heavily that recomputation dominates. That's the signal to switch to dynamic programming — see {{GUIDE:cs161-dynamic-programming}}.

## Designing your own divide-and-conquer algorithm

A reliable recipe:

1. Identify the natural "halving" structure (sorted index, geometric coordinate, bit position, range of values, ...).
2. Write the recurrence: $T(n) = a T(n/b) + f(n)$.
3. Apply the Master Theorem to predict the runtime.
4. Decide whether the predicted runtime beats the best known alternative. If not, look for a way to reduce $a$ (Karatsuba's trick) or shrink $f(n)$ (the strip argument).

For the slide deck, see {{FILE:stanford-cs161-divide-and-conquer-slides}}. For an MT refresher, see {{FILE:stanford-cs161-master-theorem-quickref}}.

## Worked recurrence catalog

| Algorithm | Recurrence | Solution |
|---|---|---|
| Binary search | $T(n) = T(n/2) + 1$ | $\Theta(\log n)$ |
| Mergesort / Closest pair | $T(n) = 2T(n/2) + n$ | $\Theta(n \log n)$ |
| Karatsuba | $T(n) = 3T(n/2) + n$ | $\Theta(n^{\log_2 3})$ |
| Strassen | $T(n) = 7T(n/2) + n^2$ | $\Theta(n^{\log_2 7})$ |
| Naive matrix mul | $T(n) = 8T(n/2) + n^2$ | $\Theta(n^3)$ |
| Median of medians (select) | $T(n) = T(n/5) + T(7n/10) + n$ | $\Theta(n)$ |

Notice how a single change — going from 8 subproblems to 7 in matrix multiplication — gives a polynomially better algorithm. That observation has motivated 50+ years of subsequent research into matrix multiplication exponents.

## Practice and next steps

For more practice, work through {{GUIDE:cs161-master-theorem}} (recurrence drills) and {{GUIDE:cs161-sorting-algorithms}}, which puts mergesort head-to-head with quicksort, heapsort, and the $\Omega(n \log n)$ comparison-sort lower bound.
