---
slug: cs161-master-theorem
course:
  ipeds_id: "243744"
  department: "CS"
  number: "161"
title: "The Master Theorem — CS 161"
description: "A mechanical recipe for solving divide-and-conquer recurrences of the form T(n) = a T(n/b) + f(n)."
tags: ["algorithms", "master-theorem", "recurrences", "divide-and-conquer", "midterm"]
author_role: bot
quiz_slug: cs161-master-theorem-quiz
attached_files:
  - stanford-cs161-master-theorem-quickref
  - stanford-cs161-divide-and-conquer-slides
attached_resources: []
---

# The Master Theorem

Almost every divide-and-conquer algorithm in CS 161 produces a recurrence of the same shape:

$$
T(n) = a \cdot T(n/b) + f(n)
$$

where $a \geq 1$ is the number of subproblems, $b > 1$ is the factor by which the input shrinks per level, and $f(n)$ is the work done **outside** the recursive calls (typically the divide + combine cost). The Master Theorem gives you the closed-form solution by inspection.

If you only memorize one tool from the recurrence section of the course, memorize this one. Keep the {{FILE:stanford-cs161-master-theorem-quickref}} card near you — it summarizes all three cases.

## The three cases

Let $n^{\log_b a}$ be the **watershed** function: this is exactly the work done at the leaves of the recursion tree.

Compare $f(n)$ to $n^{\log_b a}$ and pick a case:

- **Case 1 — Leaves dominate.** If $f(n) = O(n^{\log_b a - \epsilon})$ for some $\epsilon > 0$, then
  $$T(n) = \Theta(n^{\log_b a}).$$
- **Case 2 — Balanced.** If $f(n) = \Theta(n^{\log_b a} \log^k n)$ for some $k \geq 0$, then
  $$T(n) = \Theta(n^{\log_b a} \log^{k+1} n).$$
- **Case 3 — Root dominates.** If $f(n) = \Omega(n^{\log_b a + \epsilon})$ for some $\epsilon > 0$ AND the **regularity condition** $a f(n/b) \leq c \cdot f(n)$ holds for some $c < 1$ and all sufficiently large $n$, then
  $$T(n) = \Theta(f(n)).$$

The regularity condition almost always holds for polynomial $f$ — it's the formal escape hatch for pathological cases.

## Worked examples

### Mergesort

$T(n) = 2T(n/2) + n$. Here $a = 2$, $b = 2$, so $n^{\log_b a} = n^1 = n$. Compare $f(n) = n$ to $n$: equal, with $k = 0$. **Case 2**: $T(n) = \Theta(n \log n)$.

### Binary search

$T(n) = T(n/2) + 1$. Here $a = 1$, $b = 2$, so $n^{\log_b a} = n^0 = 1$. Compare $f(n) = 1$ to $1$: equal, with $k = 0$. **Case 2**: $T(n) = \Theta(\log n)$.

### Strassen's matrix multiplication

$T(n) = 7T(n/2) + n^2$. Here $a = 7$, $b = 2$, so $n^{\log_b a} = n^{\log_2 7} \approx n^{2.807}$. Compare $f(n) = n^2$ to $n^{2.807}$: $f$ is polynomially smaller. **Case 1**: $T(n) = \Theta(n^{\log_2 7})$.

### A Case-3 example

$T(n) = 3T(n/4) + n^2$. Here $a = 3$, $b = 4$, so $n^{\log_b a} = n^{\log_4 3} \approx n^{0.79}$. Compare $f(n) = n^2$ to $n^{0.79}$: $f$ is polynomially larger. Regularity: $3 \cdot (n/4)^2 = 3n^2/16 \leq (3/16) \cdot n^2$, with $c = 3/16 < 1$. **Case 3**: $T(n) = \Theta(n^2)$.

## When the Master Theorem fails

The MT does **not** apply if $f(n)$ falls into the gap between the cases — for example, $f(n) = \Theta(n^{\log_b a} / \log n)$ is asymptotically smaller than $n^{\log_b a}$ but not polynomially smaller. Likewise, $f(n) = \Theta(n^{\log_b a} \cdot \log n)$ does fit Case 2 (with $k = 1$), but $f(n) = \Theta(n^{\log_b a} \cdot 2^{\sqrt{\log n}})$ does not.

For these gap cases, fall back to the **recursion-tree method** or **Akra–Bazzi** (covered in CS 161 office hours but not on the midterm).

## The recursion-tree intuition

For any $T(n) = a T(n/b) + f(n)$, the recursion tree has:

- depth $\log_b n$,
- $a^i$ nodes at level $i$,
- work $f(n / b^i)$ per node at level $i$.

Total work:

$$
T(n) = \sum_{i=0}^{\log_b n - 1} a^i \cdot f(n / b^i) + \Theta(n^{\log_b a}).
$$

The three Master Theorem cases correspond to which level dominates this sum:

- Case 1: the bottom (leaves) dominates.
- Case 2: every level contributes equally — multiply by depth.
- Case 3: the top (root) dominates.

If you ever forget which case yields which $\Theta$, draw the tree.

## Quick decision flow

```
1. Identify a, b, and f(n).
2. Compute n^(log_b a).
3. Compare f(n) to n^(log_b a):
   - Polynomially smaller -> Case 1, T(n) = Theta(n^(log_b a))
   - Equal up to log^k n  -> Case 2, T(n) = Theta(n^(log_b a) * log^(k+1) n)
   - Polynomially larger and regularity holds -> Case 3, T(n) = Theta(f(n))
4. If none apply, use the recursion-tree method.
```

For the slide deck used in lecture, see {{FILE:stanford-cs161-divide-and-conquer-slides}}.

## What's next

Now that you can solve recurrences mechanically, the {{GUIDE:cs161-divide-and-conquer}} guide builds them from scratch — given a problem, derive the recurrence and apply the MT. The {{GUIDE:cs161-sorting-algorithms}} guide also leans heavily on the MT to compare mergesort, quicksort (expected case), and the $\Omega(n \log n)$ comparison-sort lower bound.

Test your mechanical fluency with {{QUIZ:cs161-master-theorem-quiz}}. The quiz mixes "identify $a, b, f$" with full-solve questions, in the same proportion as midterm problems.
