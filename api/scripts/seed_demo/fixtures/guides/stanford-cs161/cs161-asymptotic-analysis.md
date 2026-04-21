---
slug: cs161-asymptotic-analysis
course:
  ipeds_id: "243744"
  department: "CS"
  number: "161"
title: "Asymptotic Analysis — CS 161"
description: "Big-O, Big-Omega, and Big-Theta with formal definitions, intuition, and worked recurrences."
tags: ["algorithms", "big-o", "asymptotic", "analysis", "midterm"]
author_role: bot
quiz_slug: cs161-asymptotic-analysis-quiz
attached_files:
  - stanford-cs161-asymptotic-analysis-worked-examples
  - stanford-cs161-big-o-cheatsheet
attached_resources: []
---

# Asymptotic Analysis

CS 161 spends its first two weeks teaching you to **stop counting individual operations** and instead reason about how the running time of an algorithm grows as the input size $n$ grows large. This shift — from concrete to asymptotic — is the single most useful idea you will take away from the course.

## Why we need asymptotics

If algorithm $A$ runs in $100n$ steps and algorithm $B$ runs in $n^2 / 2$ steps, which is faster? It depends on $n$. For $n = 10$, $A$ takes $1000$ steps and $B$ takes $50$. For $n = 10^6$, $A$ takes $10^8$ and $B$ takes $5 \times 10^{11}$. As inputs grow, the **shape** of the function dominates the constants. Asymptotic notation makes that shape the first-class object of analysis.

Three things change between course-1 thinking and CS 161 thinking:

1. We ignore additive lower-order terms — $3n^2 + 5n + 100$ becomes $\Theta(n^2)$.
2. We ignore constant multipliers — $100n$ and $0.001n$ are both $\Theta(n)$.
3. We care only about behavior as $n \to \infty$.

Hardware-level constants matter in practice — but they belong to a different conversation (cache lines, vector instructions, branch prediction). Algorithms class is about the function, not the hardware.

## Formal definitions

For functions $f, g : \mathbb{N} \to \mathbb{R}_{\geq 0}$:

- **$f(n) = O(g(n))$** if there exist constants $c > 0$ and $n_0 \geq 0$ such that $0 \leq f(n) \leq c \cdot g(n)$ for all $n \geq n_0$.
- **$f(n) = \Omega(g(n))$** if there exist $c > 0$ and $n_0 \geq 0$ such that $0 \leq c \cdot g(n) \leq f(n)$ for all $n \geq n_0$.
- **$f(n) = \Theta(g(n))$** iff $f(n) = O(g(n))$ AND $f(n) = \Omega(g(n))$.

Notice the asymmetry. $O$ is an **upper** bound, $\Omega$ is a **lower** bound, $\Theta$ is **tight** — both at once. Saying "this algorithm runs in $O(n^3)$" is technically true for any $O(n^2)$ algorithm; the statement is informative only if you mean the *tightest* bound you know.

There are also strict relatives:

- **$f(n) = o(g(n))$** ("little-o") means $\lim_{n \to \infty} f(n)/g(n) = 0$.
- **$f(n) = \omega(g(n))$** ("little-omega") means $\lim_{n \to \infty} f(n)/g(n) = \infty$.

So $n = o(n^2)$ but $n^2 \neq o(n^2)$.

## A reusable proof template

Most asymptotic proofs follow a fixed shape. To show $f(n) = O(g(n))$:

1. Pick a candidate constant $c$.
2. Pick a threshold $n_0$.
3. Verify $f(n) \leq c \cdot g(n)$ for all $n \geq n_0$.

Example: $3n^2 + 5n + 100 = O(n^2)$.

For $n \geq 100$: $5n \leq n^2$ (since $5 \leq n$ when $n \geq 5$, and certainly when $n \geq 100$) and $100 \leq n^2$ (since $n \geq 10$). So:

$$
3n^2 + 5n + 100 \leq 3n^2 + n^2 + n^2 = 5n^2.
$$

Witness pair: $c = 5$, $n_0 = 100$. The pair is not unique — picking $c = 9$ and $n_0 = 1$ also works. Existence is what matters.

## Common growth classes

From slowest to fastest growing:

```
1  <  log log n  <  log n  <  sqrt(n)  <  n  <  n log n  <  n^2  <  n^3  <  2^n  <  n!  <  n^n
```

Internalize this hierarchy. It tells you, at a glance, whether your $O(n \log n)$ idea beats your friend's $O(n^2)$ idea (it does, asymptotically, every time).

A few easy traps:

- $\log_2 n$ vs. $\log_{10} n$ vs. $\ln n$ — all $\Theta(\log n)$. Bases differ by a constant factor.
- $\log(n!) = \Theta(n \log n)$ by Stirling's approximation. Useful for sorting lower bounds.
- $2^{\log n} = n$ (when $\log = \log_2$). Don't get scared by exponentials over logarithms.
- $n^{1/\log n}$ is constant — namely $2$, when $\log = \log_2$.

## Walking through a real example

Consider this fragment:

```python
def quadratic_pairs(arr):
    n = len(arr)
    count = 0
    for i in range(n):
        for j in range(i + 1, n):
            if arr[i] + arr[j] == 0:
                count += 1
    return count
```

How many iterations of the inner loop run? When $i = 0$: $n - 1$. When $i = 1$: $n - 2$. ... When $i = n-1$: $0$. Sum:

$$
\sum_{i=0}^{n-1} (n - 1 - i) = \frac{n(n-1)}{2} = \Theta(n^2).
$$

We did **not** care that the body is one comparison and one increment. Constant work per iteration multiplied by $\Theta(n^2)$ iterations gives $\Theta(n^2)$ total — done.

For more worked examples in this exact format, see {{FILE:stanford-cs161-asymptotic-analysis-worked-examples}}, and keep the {{FILE:stanford-cs161-big-o-cheatsheet}} taped above your desk for the midterm.

## Common pitfalls

- **"Asymptotic equality" is not symmetric.** Writing $O(n) \subseteq O(n^2)$ is true. The `=` in $f(n) = O(g(n))$ is conventional shorthand for $f \in O(g)$ — it is not transitive in the usual sense.
- **Constants matter for picking algorithms in practice.** Asymptotic comparisons assume large $n$. For tiny $n$, $O(n^2)$ insertion sort beats $O(n \log n)$ mergesort because of cache effects and constant overhead. This is why real-world quicksort implementations switch to insertion sort for small subarrays.
- **Don't ignore the input model.** "Sorting in $O(n \log n)$" assumes comparison-based sorting. Counting sort is $O(n + k)$ — but it's not comparison-based, so it doesn't violate the $\Omega(n \log n)$ lower bound.

## What to practice

After reading the worked examples, attempt:

1. Prove $\frac{1}{2}n^2 - 3n = \Theta(n^2)$ from first principles (witness pair for both $O$ and $\Omega$).
2. Show $n \log n = O(n^{1.5})$ but $n^{1.5} \neq O(n \log n)$.
3. Rank in increasing order of growth: $n^{2/3}, \log^5 n, 2^{\sqrt{\log n}}, n / \log n, n^{0.99}, \sqrt{n} \log n$.

Then take the {{QUIZ:cs161-asymptotic-analysis-quiz}} to verify your fluency. Asymptotics are the language of the rest of the course — when we discuss the {{GUIDE:cs161-master-theorem}} next, every recurrence solution will live inside this notation.
