---
slug: stanford-cs161-asymptotic-analysis-worked-examples
title: "Asymptotic Analysis: Worked Examples"
mime: application/pdf
filename: asymptotic-analysis-worked-examples.pdf
course: stanford/cs161
description: "Step-by-step worked examples of tight Big-O, Big-Theta, and recurrence analyses for common CS 161 patterns."
author_role: bot
---

# Asymptotic Analysis: Worked Examples

The cheatsheet gives the rules. This document walks through the *moves*: how an analysis actually unfolds when a problem lands on your exam. Each example is solved twice — once informally, once with the formal definition — so you can see when each style pays off.

## Example 1: Three Nested Loops with a Dependency

```python
def triangle_work(n: int) -> int:
    count = 0
    for i in range(n):
        for j in range(i, n):
            for k in range(j, n):
                count += 1
    return count
```

**Informal.** The outer loop picks `$i$`, then `$j \ge i$`, then `$k \ge j$`. The number of triples is `$\binom{n+2}{3} \approx n^3/6$`, so the runtime is `$\Theta(n^3)$`.

**Formal.** `$T(n) = \sum_{i=0}^{n-1}\sum_{j=i}^{n-1}(n-j) = \Theta(n^3)$`. Dropping the `$1/6$` constant is legal because it is independent of `$n$`.

## Example 2: Loop with a Non-Constant Step

```python
def doubling_scan(n: int) -> int:
    i = 1
    count = 0
    while i < n:
        count += 1
        i *= 2
    return count
```

The loop runs while `$2^k < n$`, so `$k < \log_2 n$`. Runtime: `$\Theta(\log n)$`.

## Example 3: Mixed Linear and Logarithmic Work

```python
def sorted_scan(arr):
    arr.sort()                     # O(n log n)
    out = []
    for x in arr:                  # O(n)
        if binary_search(arr, x):  # O(log n)
            out.append(x)
    return out
```

Sequential blocks: `$\Theta(n \log n) + \Theta(n \cdot \log n) = \Theta(n \log n)$`. The tempting wrong answer is `$\Theta(n^2 \log n)$` — students confuse the *for* with the *binary search* and multiply when they should add work per iteration.

## Example 4: Recurrence with Unbalanced Split

`$T(n) = T(n/3) + T(2n/3) + \Theta(n)$`.

Master theorem does not apply directly (unequal subproblems). Use a recursion tree: each level does `$\Theta(n)$` work; the longest root-to-leaf path is `$\log_{3/2} n$`. Total: `$\Theta(n \log n)$`. This recurrence appears in worst-case quicksort with a good pivot guarantee.

## Example 5: Recurrence with a Subtractive Reduction

`$T(n) = T(n-1) + \Theta(n)$`.

Expand: `$T(n) = \Theta(n) + \Theta(n-1) + \cdots + \Theta(1) = \Theta(n^2)$`. This is the runtime of naive insertion sort and of the outer structure of selection sort.

## Example 6: Amortized Analysis of a Dynamic Array

A dynamic array doubles on overflow. A single `append` can be `$\Theta(n)$` (copy), but the **amortized** cost is `$\Theta(1)$`. Banker's argument: each element pays `$3$` credits on insertion — `$1$` to be placed, `$1$` to move itself at the next resize, `$1$` to move an earlier element. Total credits cover all resizes.

## Proof Moves to Memorize

1. To prove `$f = O(g)$`, find constants `$c, n_0$` and show `$f(n) \le c g(n)$` for `$n \ge n_0$`.
2. To disprove `$f = O(g)$`, show `$\lim_{n\to\infty} f(n)/g(n) = \infty$`.
3. Polynomials dominate logarithms: `$\log^k n = o(n^\epsilon)$` for any `$\epsilon > 0$`.
4. Exponentials dominate polynomials: `$n^k = o(c^n)$` for any `$c > 1$`.

## Exam Rubric Signals

- Always state the bound you claim (`$O$`, `$\Theta$`, `$\Omega$`).
- If asked for tight bounds, `$O$` alone loses points.
- Justify with either a summation, a recursion tree, or the master theorem — never "by inspection" for anything nontrivial.
