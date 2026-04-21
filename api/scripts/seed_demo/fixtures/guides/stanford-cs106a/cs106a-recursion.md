---
slug: cs106a-recursion
course:
  ipeds_id: "243744"
  department: "CS"
  number: "106A"
title: "Recursion — CS 106A"
description: "How to think recursively in Python: base cases, recursive cases, and the call stack."
tags: ["python", "recursion", "algorithms", "midterm"]
author_role: bot
quiz_slug: cs106a-recursion-quiz
attached_files:
  - stanford-cs106a-recursion-primer
attached_resources: []
---

# Recursion

Recursion is the topic CS 106A students fear most and the topic that, in
a way, separates "I write code" from "I think computationally". A
recursive function calls itself on a smaller version of the same problem,
trusting that the smaller call will work, and combining the result with
whatever it can do at the current level.

The primer at `{{FILE:stanford-cs106a-recursion-primer}}` walks through
the call-stack diagrams; this guide focuses on the patterns.

## The two-piece template

Every recursive function has exactly two parts:

1. A **base case** — a problem so small you can answer it directly,
   no recursion needed.
2. A **recursive case** — solve a smaller version of the same problem,
   then combine its result with whatever you do at this level.

Forget the base case and you get infinite recursion → `RecursionError`.
Forget to make the problem smaller and you get the same.

## Canonical example: factorial

```python
def factorial(n: int) -> int:
    """Return n! for n >= 0."""
    if n == 0:                  # base case
        return 1
    return n * factorial(n - 1) # recursive case
```

Trace `factorial(3)`:

```
factorial(3)
= 3 * factorial(2)
= 3 * (2 * factorial(1))
= 3 * (2 * (1 * factorial(0)))
= 3 * (2 * (1 * 1))
= 6
```

Each call sits on the **call stack** until the base case returns. Then
the stack unwinds, multiplying as it goes.

## The leap of faith

The single hardest mental move in recursion is to **assume the recursive
call works** without tracing into it. Don't simulate the whole call
stack in your head — that scales poorly.

Instead, ask:

> "If I had a magic helper that could solve `factorial(n - 1)` for me,
> how would I use its answer to solve `factorial(n)`?"

The helper IS the same function. The leap of faith is trusting that as
long as your base case is right and your recursive case shrinks toward
it, the function will work.

## String recursion

Strings are sequences, so the same template applies:

```python
def reverse(s: str) -> str:
    if len(s) <= 1:
        return s
    return reverse(s[1:]) + s[0]
```

Trace `reverse("abc")`:

```
reverse("abc") = reverse("bc") + "a"
              = (reverse("c") + "b") + "a"
              = ("c" + "b") + "a"
              = "cba"
```

A small efficiency note: each `s[1:]` creates a new string, so this is
O(n^2) in time. For CS 106A grading purposes that's fine. Real-world
string reversal uses `s[::-1]`.

## List recursion

```python
def list_sum(nums: list[int]) -> int:
    if not nums:               # base case: empty list
        return 0
    return nums[0] + list_sum(nums[1:])
```

Same shape: shrink the input, combine the head with the recursive result
on the tail.

## Recursion vs iteration

Anything you can do with recursion you can do with iteration, and vice
versa. The choice is about **which form makes the algorithm clearer**.

| Best as recursion | Best as iteration |
|---|---|
| Tree / nested-structure traversals | Counting up to N |
| Divide-and-conquer (binary search, merge sort) | Accumulating a single running value |
| Naturally recursive math (factorial, Fibonacci) | Stepping through a fixed sequence |
| Mutual recursion / state machines | Anything `for x in xs:` handles cleanly |

CS 106A grades both correctness and clarity. Don't force recursion onto
problems that are obviously iterative.

## The call stack and recursion depth

Each pending call uses a stack frame. Python defaults to a recursion
limit of 1000:

```python
import sys
sys.getrecursionlimit()    # 1000
```

Bust it and you get:

```
RecursionError: maximum recursion depth exceeded
```

In CS 106A, every problem is small enough that you'll never hit this in
correct code. If you do hit it, your base case is wrong.

## Helper functions and accumulators

Sometimes the natural recursive shape needs an extra parameter. The
clean idiom is a wrapper + helper pair:

```python
def list_sum(nums: list[int]) -> int:
    return _sum_from(nums, 0)

def _sum_from(nums: list[int], i: int) -> int:
    if i >= len(nums):
        return 0
    return nums[i] + _sum_from(nums, i + 1)
```

This avoids the `nums[1:]` copy on every call — the index `i` does the
shrinking instead. The leading underscore signals "internal helper".

## Multiple recursive calls

When a single problem decomposes into several smaller subproblems:

```python
def count_paths(rows: int, cols: int) -> int:
    """Number of paths from (0,0) to (rows-1, cols-1) moving only down/right."""
    if rows == 1 or cols == 1:
        return 1
    return count_paths(rows - 1, cols) + count_paths(rows, cols - 1)
```

`count_paths(3, 3)` works because each call splits into two smaller
calls and the leaves are size-1 grids that return 1. CS 106A "exhaustive
search" problems all share this shape.

## Common pitfalls

- **No base case** → `RecursionError`. Add one.
- **Base case never hit** → either your shrinkage is wrong or the
  parameter isn't moving toward the base. Print the call args at the top
  of the function and watch.
- **Ignoring the recursive return value** → `helper(rest)` instead of
  `return ... helper(rest)`. The recursive call has to be *used*.
- **Mutating the input** → leads to confusing behavior because every
  call sees the same underlying list. Pass slices or indices instead.

## Practice

Take {{QUIZ:cs106a-recursion-quiz}}. The hardest question is the
"identify the bug" one — make sure you can articulate why each base case
is or isn't correct. For exhaustive-search problems specifically, jump to
{{GUIDE:cs106a-midterm-review}} for additional patterns.
