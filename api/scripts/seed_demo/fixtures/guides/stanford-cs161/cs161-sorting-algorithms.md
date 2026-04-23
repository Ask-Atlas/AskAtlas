---
slug: cs161-sorting-algorithms
course:
  ipeds_id: "243744"
  department: "CS"
  number: "161"
title: "Sorting Algorithms — CS 161"
description: "Mergesort, quicksort, heapsort, and the comparison-sort lower bound — when to pick which."
tags: ["algorithms", "sorting", "quicksort", "mergesort", "heapsort", "midterm"]
author_role: bot
attached_files:
  - stanford-cs161-divide-and-conquer-slides
  - stanford-cs161-big-o-cheatsheet
attached_resources: []
---

# Sorting Algorithms

Sorting is the canonical problem for studying algorithm design. Every paradigm — divide and conquer, randomization, amortized analysis, lower bounds — shows up in the sorting universe. CS 161 expects fluency in four algorithms and the lower-bound proof.

## Insertion sort

```python
def insertion_sort(a):
    for i in range(1, len(a)):
        x = a[i]
        j = i - 1
        while j >= 0 and a[j] > x:
            a[j + 1] = a[j]
            j -= 1
        a[j + 1] = x
```

- **Worst case**: $\Theta(n^2)$ (reverse-sorted input).
- **Best case**: $\Theta(n)$ (already sorted — the inner loop never executes).
- **Space**: $\Theta(1)$ extra.
- **Stable**: yes.

Insertion sort is what production sorting libraries fall back to for small subarrays (typically $n \leq 16$). Its constant factors are tiny because the inner loop is just a compare + move.

## Mergesort

See {{GUIDE:cs161-divide-and-conquer}} for the derivation. Summary:

- $\Theta(n \log n)$ in **all** cases (best, average, worst).
- $\Theta(n)$ extra space (the merge buffer).
- Stable.
- Performance is independent of input distribution — predictable runtime is its biggest selling point.

## Quicksort

```python
def quicksort(a, lo, hi):
    if lo >= hi:
        return
    p = partition(a, lo, hi)
    quicksort(a, lo, p - 1)
    quicksort(a, p + 1, hi)

def partition(a, lo, hi):
    pivot = a[hi]
    i = lo - 1
    for j in range(lo, hi):
        if a[j] <= pivot:
            i += 1
            a[i], a[j] = a[j], a[i]
    a[i + 1], a[hi] = a[hi], a[i + 1]
    return i + 1
```

- **Worst case**: $\Theta(n^2)$ (e.g., always-rightmost pivot on a sorted input).
- **Average case** with random pivot: $\Theta(n \log n)$ in expectation. The expected number of comparisons is $\approx 1.39 n \log n$.
- **Space**: $O(\log n)$ recursion stack on average.
- **Stable**: no.

The expected analysis: for any pair of elements at ranks $i < j$ in the final sorted order, they are compared exactly once iff one of them is chosen as a pivot before any element strictly between them. This happens with probability $2 / (j - i + 1)$. Summing across all pairs gives $\Theta(n \log n)$.

This is one of the most beautiful expected-running-time arguments in the course. It does **not** require any assumption about the input distribution — just that the pivot is chosen uniformly at random from the current subarray.

## Heapsort

Build a max-heap in $\Theta(n)$ (bottom-up `heapify`), then repeatedly extract the max and place it at the end of the array. The running time is $\Theta(n \log n)$ in all cases, $\Theta(1)$ extra space, but **not** stable.

The bottom-up heap construction is the surprising part. Inserting $n$ elements one at a time into an initially empty heap is $\Theta(n \log n)$. But heapifying an array in place — running `sift_down` on positions $n/2, n/2 - 1, \ldots, 1$ — is only $\Theta(n)$, because most positions are near the bottom and have very small subtrees.

## The comparison-sort lower bound

**Theorem.** Any deterministic comparison-based sorting algorithm requires $\Omega(n \log n)$ comparisons in the worst case.

**Proof sketch.** A comparison-based algorithm can be modeled as a binary decision tree where each internal node is a comparison and each leaf is a permutation of the input. There are $n!$ permutations, so the tree has at least $n!$ leaves and depth at least $\log_2(n!) = \Theta(n \log n)$ by Stirling. The worst-case number of comparisons equals the depth of the deepest leaf.

**Consequence.** No comparison-based algorithm can asymptotically beat $O(n \log n)$. Mergesort and heapsort are optimal in this model.

## Beating the lower bound: linear-time sorts

If the input has structure beyond comparability, you can sort faster. Two important non-comparison sorts:

- **Counting sort.** When keys are integers in $\{0, 1, \ldots, k\}$: count occurrences in $\Theta(n + k)$, then write back. Stable, but only useful when $k$ is small.
- **Radix sort.** Sort by least significant digit first using counting sort as the inner loop. Total time $\Theta(d \cdot (n + b))$ where $d$ is the number of digits and $b$ is the base. For 32-bit integers with $b = 256$, this is $\Theta(n)$ in practice.

These do not violate the lower bound — they sidestep it by leaving the comparison model.

## Choosing in practice

| If you have... | Use |
|---|---|
| Tiny array ($n \leq 16$) | Insertion sort |
| Untrusted/adversarial input | Heapsort or mergesort |
| Large array, random-looking input | Randomized quicksort |
| Memory-constrained, in-place required | Heapsort |
| Stability required | Mergesort |
| Small-integer keys | Counting / radix sort |

For visual reference, see the lecture deck in {{FILE:stanford-cs161-divide-and-conquer-slides}} and the cheatsheet in {{FILE:stanford-cs161-big-o-cheatsheet}}.

## Practice

For complementary practice, take the {{QUIZ:cs161-master-theorem-quiz}} (mergesort recurrence) and {{QUIZ:cs161-asymptotic-analysis-quiz}} (Big-O reasoning). The next pivot in the course narrative is {{GUIDE:cs161-dynamic-programming}}, which leaves divide-and-conquer behind for problems with overlapping subproblems.
