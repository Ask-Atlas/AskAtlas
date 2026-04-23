---
slug: cs161-dynamic-programming
course:
  ipeds_id: "243744"
  department: "CS"
  number: "161"
title: "Dynamic Programming — CS 161"
description: "The five-step DP recipe with worked examples in Fibonacci, LIS, edit distance, knapsack, and matrix chain."
tags: ["algorithms", "dynamic-programming", "memoization", "optimization", "final"]
author_role: bot
quiz_slug: cs161-dynamic-programming-quiz
attached_files:
  - stanford-cs161-dp-patterns
  - stanford-cs161-big-o-cheatsheet
attached_resources: []
---

# Dynamic Programming

Dynamic programming (DP) is the right hammer when:

1. The problem has **optimal substructure** — the optimal answer to the whole composes from optimal answers to subproblems.
2. The subproblems **overlap** — naive recursion would solve the same subproblem many times.

If only (1) holds and subproblems are disjoint, divide and conquer ({{GUIDE:cs161-divide-and-conquer}}) is enough. If (2) also holds, DP turns an exponential blowup into a polynomial.

## The five-step DP recipe

1. **Define the subproblems.** Write down what `dp[i]` (or `dp[i][j]`) **means**, in plain English. This is the hardest step. Vague subproblems produce buggy code.
2. **State the recurrence.** Express `dp[i]` in terms of strictly smaller indices.
3. **Identify the base cases.** What are the smallest subproblems that don't recurse?
4. **Choose the evaluation order.** Bottom-up loops or top-down memoization. Both are correct; bottom-up is usually faster in practice.
5. **Read off the answer.** Often `dp[n]` or `dp[n][m]`, but sometimes a max over the whole table.

When stuck, write step 1 in a single sentence with no symbols. If you can't, the recurrence will not appear.

## Worked example 1: Fibonacci

Subproblem: `dp[i] = the i-th Fibonacci number`. Recurrence: `dp[i] = dp[i-1] + dp[i-2]`. Base cases: `dp[0] = 0`, `dp[1] = 1`. Answer: `dp[n]`.

```python
def fib(n):
    if n < 2:
        return n
    dp = [0] * (n + 1)
    dp[1] = 1
    for i in range(2, n + 1):
        dp[i] = dp[i - 1] + dp[i - 2]
    return dp[n]
```

Time $\Theta(n)$, space $\Theta(n)$. Can be reduced to $\Theta(1)$ space by keeping only the last two values — a classic DP space optimization.

## Worked example 2: Longest Increasing Subsequence (LIS)

Given an array $A$, find the length of the longest strictly increasing subsequence (not necessarily contiguous).

**Subproblem.** $L(i) = $ length of the LIS that **ends at position $i$**.

**Recurrence.** $L(i) = 1 + \max\{ L(j) : j < i \text{ and } A[j] < A[i] \}$, with the max over an empty set being $0$.

**Base case.** Implicit: when no valid $j$ exists, $L(i) = 1$.

**Answer.** $\max_i L(i)$.

```python
def lis(a):
    n = len(a)
    L = [1] * n
    for i in range(n):
        for j in range(i):
            if a[j] < a[i]:
                L[i] = max(L[i], L[j] + 1)
    return max(L) if L else 0
```

Time $\Theta(n^2)$. There is a famous $\Theta(n \log n)$ algorithm using patience sorting / binary search, but the DP is the standard CS 161 baseline.

## Worked example 3: Edit Distance

Given strings $X$ (length $m$) and $Y$ (length $n$), find the minimum number of single-character insertions, deletions, or substitutions to turn $X$ into $Y$.

**Subproblem.** $D(i, j) = $ edit distance between $X[1..i]$ and $Y[1..j]$.

**Recurrence.**
$$
D(i, j) = \min \begin{cases}
D(i-1, j) + 1 & (\text{delete } X[i]) \\
D(i, j-1) + 1 & (\text{insert } Y[j]) \\
D(i-1, j-1) + [X[i] \neq Y[j]] & (\text{substitute or match})
\end{cases}
$$

**Base cases.** $D(i, 0) = i$ (delete everything), $D(0, j) = j$ (insert everything).

**Answer.** $D(m, n)$.

Table size $\Theta(mn)$, work per cell $\Theta(1)$. Total $\Theta(mn)$.

This algorithm — the **Wagner–Fischer algorithm** — is the classical example of 2D bottom-up DP. It also yields the alignment itself by walking back through the table from $(m, n)$ to $(0, 0)$ following the argmin at each cell.

## Worked example 4: 0/1 Knapsack

Items $1 \ldots n$ with weights $w_i$ and values $v_i$. Find the max-value subset whose total weight does not exceed capacity $W$.

**Subproblem.** $K(i, c) = $ max value using a subset of items $1 \ldots i$ with total weight $\leq c$.

**Recurrence.**
$$
K(i, c) = \begin{cases}
K(i-1, c) & \text{if } w_i > c \\
\max(K(i-1, c), \, v_i + K(i-1, c - w_i)) & \text{otherwise}
\end{cases}
$$

**Base case.** $K(0, c) = 0$ for all $c$.

**Answer.** $K(n, W)$.

Time $\Theta(nW)$. This is **pseudo-polynomial** — polynomial in the *value* of $W$, but exponential in the *number of bits* used to encode $W$. This subtle point becomes important in {{GUIDE:cs161-np-completeness}}, where knapsack turns out to be NP-hard.

## Worked example 5: Matrix Chain Multiplication

Given matrices $A_1, A_2, \ldots, A_n$ with dimensions $p_0 \times p_1, p_1 \times p_2, \ldots, p_{n-1} \times p_n$, find the order of multiplication that minimizes the number of scalar multiplications.

**Subproblem.** $M(i, j) = $ minimum cost to multiply $A_i \cdots A_j$.

**Recurrence.** $M(i, j) = \min_{i \leq k < j} \big( M(i, k) + M(k+1, j) + p_{i-1} p_k p_j \big)$.

**Base case.** $M(i, i) = 0$.

**Answer.** $M(1, n)$.

Iterate by chain length, not by index — fill diagonals first. Time $\Theta(n^3)$, space $\Theta(n^2)$.

## Memoization vs tabulation

Both are valid implementations. Memoization (top-down) writes the recurrence directly and caches results in a dict; tabulation (bottom-up) fills an array in dependency order. They differ in:

- **Constant factor.** Tabulation typically wins by 2–4× because of better cache behavior and no recursion overhead.
- **Sparsity.** Memoization avoids computing unreached cells. If the dependency graph is sparse, top-down can be asymptotically faster.
- **Stack depth.** Top-down can blow Python's recursion limit on long chains. Use `sys.setrecursionlimit` or convert to bottom-up.

For more pattern catalogs and templates, see {{FILE:stanford-cs161-dp-patterns}}.

## How to recognize a DP problem on the exam

Three reliable tells:

1. The problem asks for an **optimum** (max, min, count) over an exponentially large space.
2. Greedy ({{GUIDE:cs161-greedy-algorithms}}) seems plausible but you can construct a small counterexample.
3. You can describe a state of size $O(\text{poly}(n))$ that captures everything you need to know about the future.

If all three hit, write the subproblem definition and the recurrence will follow. Test your fluency with {{QUIZ:cs161-dynamic-programming-quiz}}.
