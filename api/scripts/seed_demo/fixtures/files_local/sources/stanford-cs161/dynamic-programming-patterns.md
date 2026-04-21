---
slug: stanford-cs161-dp-patterns
title: "Dynamic Programming Patterns: Memoization, Tabulation, and Classic Problems"
mime: application/pdf
filename: dp-patterns.pdf
course: stanford/cs161
description: "Core DP techniques illustrated with knapsack, LCS, and edit distance. Covers both memoization and tabulation."
author_role: bot
---

# Dynamic Programming Patterns

Dynamic programming (DP) solves problems with **optimal substructure** (optimal solution composed of optimal subsolutions) and **overlapping subproblems** (the same subproblems recur).

## Two Implementation Styles

| Style | Direction | Pros | Cons |
|-------|-----------|------|------|
| **Memoization** | Top-down (recursion + cache) | Only computes needed states; mirrors recurrence | Recursion overhead; stack depth |
| **Tabulation** | Bottom-up (iterative fill) | No recursion; predictable order; often faster | May compute unused states |

## Fibonacci: The Canonical Example

Naive recursion: `$T(n) = T(n-1) + T(n-2) + O(1) = \Theta(\varphi^n)$`.

```python
def fib_memo(n, cache=None):
    cache = cache if cache is not None else {}
    if n < 2:
        return n
    if n in cache:
        return cache[n]
    cache[n] = fib_memo(n - 1, cache) + fib_memo(n - 2, cache)
    return cache[n]

def fib_tab(n):
    if n < 2:
        return n
    dp = [0] * (n + 1)
    dp[1] = 1
    for i in range(2, n + 1):
        dp[i] = dp[i - 1] + dp[i - 2]
    return dp[n]
```

Both run in `$\Theta(n)$`. Tabulation can be reduced to `$O(1)$` space.

## 0/1 Knapsack

Given items with weights `$w_i$` and values `$v_i$`, capacity `$W$`, maximize total value without exceeding `$W$`.

**Recurrence:**

`$dp[i][w] = \max(dp[i-1][w], \; dp[i-1][w - w_i] + v_i)$`

```python
def knapsack(weights, values, W):
    n = len(weights)
    dp = [[0] * (W + 1) for _ in range(n + 1)]
    for i in range(1, n + 1):
        for w in range(W + 1):
            dp[i][w] = dp[i - 1][w]
            if weights[i - 1] <= w:
                take = dp[i - 1][w - weights[i - 1]] + values[i - 1]
                dp[i][w] = max(dp[i][w], take)
    return dp[n][W]
```

Time: `$\Theta(nW)$` (pseudo-polynomial). Space can be reduced to `$O(W)$` with a rolling 1D array.

## Longest Common Subsequence (LCS)

Given strings `$X$` and `$Y$`, find the longest subsequence appearing in both.

**Recurrence:**

- If `$X_i = Y_j$`: `$dp[i][j] = dp[i-1][j-1] + 1$`
- Else: `$dp[i][j] = \max(dp[i-1][j], dp[i][j-1])$`

```python
def lcs(X, Y):
    m, n = len(X), len(Y)
    dp = [[0] * (n + 1) for _ in range(m + 1)]
    for i in range(1, m + 1):
        for j in range(1, n + 1):
            if X[i - 1] == Y[j - 1]:
                dp[i][j] = dp[i - 1][j - 1] + 1
            else:
                dp[i][j] = max(dp[i - 1][j], dp[i][j - 1])
    return dp[m][n]
```

Time and space: `$\Theta(mn)$`.

## Edit Distance (Levenshtein)

Minimum insertions, deletions, and substitutions to transform `$X$` into `$Y$`.

**Recurrence:**

- `$dp[i][0] = i$`, `$dp[0][j] = j$`
- If `$X_i = Y_j$`: `$dp[i][j] = dp[i-1][j-1]$`
- Else: `$dp[i][j] = 1 + \min(dp[i-1][j], \; dp[i][j-1], \; dp[i-1][j-1])$`

Time and space: `$\Theta(mn)$`.

## DP Problem-Solving Recipe

1. Identify the decision at each step.
2. Define state variables capturing enough history.
3. Write the recurrence and base cases.
4. Decide top-down or bottom-up.
5. Analyze time and space; reduce dimensions if possible.
6. Reconstruct the solution by walking back pointers or parent arrays.
