---
slug: stanford-cs106a-recursion-primer
title: "Recursion Primer: Base Cases and Recursive Cases"
mime: application/pdf
filename: recursion-primer.pdf
course: stanford/cs106a
description: "Primer on recursion in Python: identifying base cases, writing recursive cases, and classic examples."
author_role: bot
---

# Recursion Primer

Recursion is a function that calls itself with a smaller version of the same problem. Every correct recursive function has two parts: a **base case** that stops the recursion, and a **recursive case** that reduces the problem toward the base case.

## The Mental Model

Think of recursion as *trust*. You assume the function already works on a smaller input, and you use that result to build the answer for the current input.

```python
def countdown(n: int) -> None:
    if n <= 0:          # BASE CASE
        print("Go!")
        return
    print(n)
    countdown(n - 1)    # RECURSIVE CASE
```

## Example 1: Factorial

`n! = n * (n-1)!`, with `0! = 1`.

```python
def factorial(n: int) -> int:
    if n == 0:
        return 1
    return n * factorial(n - 1)
```

Trace `factorial(4)`:

```
factorial(4) = 4 * factorial(3)
             = 4 * 3 * factorial(2)
             = 4 * 3 * 2 * factorial(1)
             = 4 * 3 * 2 * 1 * factorial(0)
             = 4 * 3 * 2 * 1 * 1 = 24
```

## Example 2: Fibonacci

`fib(n) = fib(n-1) + fib(n-2)`, with `fib(0)=0` and `fib(1)=1`.

```python
def fib(n: int) -> int:
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)
```

Naive `fib` is exponential (`O(2^n)`) because it recomputes the same sub-problems. Memoization fixes this:

```python
from functools import lru_cache

@lru_cache(maxsize=None)
def fib_fast(n: int) -> int:
    if n < 2:
        return n
    return fib_fast(n - 1) + fib_fast(n - 2)
```

## Example 3: Sum of a List

```python
def total(nums: list[int]) -> int:
    if not nums:          # base: empty list
        return 0
    return nums[0] + total(nums[1:])
```

## Recursion vs. Iteration

| Aspect         | Recursion                        | Iteration                         |
|----------------|----------------------------------|-----------------------------------|
| Readability    | Matches naturally recursive data | Clear for linear sequences        |
| Stack cost     | Each call adds a stack frame     | Constant stack use                |
| Termination    | Base case must be reachable      | Loop condition must eventually fail |
| Best for       | Trees, divide-and-conquer        | Counting, accumulating            |

## Common Bugs

1. **Missing base case** — leads to `RecursionError: maximum recursion depth exceeded`.
2. **Base case never reached** — e.g. recursing with `n` instead of `n - 1`.
3. **Forgetting to `return`** — the recursive call's value is computed and thrown away.

## Checklist Before You Submit

- Does the function return on its base case?
- Does every recursive call move strictly closer to the base case?
- Have you tested the smallest inputs (`0`, `""`, `[]`)?
- Have you traced at least one non-trivial call by hand?

Recursion feels magical until you've traced a few calls on paper. Do that once per week, and it clicks.
