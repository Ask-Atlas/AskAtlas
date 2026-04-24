---
slug: cs106a-list-comprehensions
course:
  ipeds_id: "243744"
  department: "CS"
  number: "106A"
title: "List Comprehensions — CS 106A"
description: "Pythonic transforms: replacing build-up loops with concise, expressive comprehensions."
tags: ["python", "list-comprehensions", "idioms", "midterm"]
author_role: bot
quiz_slug: cs106a-list-comprehensions-quiz
attached_files:
  - stanford-cs106a-list-comprehensions-cheatsheet
attached_resources: []
---

# List Comprehensions

A list comprehension is a single-expression form of the build-up loop:
"give me a new list, where each element is f(x) for x in some
iterable." Once you read them fluently they're shorter, faster, and
clearer than the equivalent `for` loop.

The cheatsheet at
`{{FILE:stanford-cs106a-list-comprehensions-cheatsheet}}` has the
syntax-card version of this guide.

## The basic shape

```python
squares = [x * x for x in range(5)]
# [0, 1, 4, 9, 16]
```

This is exactly equivalent to:

```python
squares = []
for x in range(5):
    squares.append(x * x)
```

Three pieces:

1. The **expression** computed for each element (`x * x`).
2. A **`for`** clause naming the iteration variable (`x`) and the
   iterable (`range(5)`).
3. Optional **`if`** filter (covered next).

## With a filter

Add `if condition` to keep only elements that match:

```python
positives = [n for n in nums if n > 0]
even_squares = [x * x for x in range(10) if x % 2 == 0]
```

The filter runs *after* the iteration variable is bound and *before* the
expression is evaluated. So you can filter on derived values too:

```python
short_words = [w for w in words if len(w) < 5]
```

## With a transform AND a filter

```python
upper_short = [w.upper() for w in words if len(w) < 5]
```

Read it left-to-right: "uppercase each w, for each w in words, where
len(w) < 5".

## Conditional expression

Filtering REMOVES elements; conditional EXPRESSION transforms them
differently based on a condition. The two are different and the placement
of `if` tells you which:

```python
# filter: keep only positives
[n for n in nums if n > 0]

# conditional expression: keep all, transform negatives to 0
[n if n > 0 else 0 for n in nums]
```

The mnemonic: **`if` after `for` filters, `if/else` before `for`
transforms.**

## Multiple `for` clauses

The clauses nest left-to-right, outer first:

```python
pairs = [(x, y) for x in range(3) for y in range(3)]
# (0,0), (0,1), (0,2), (1,0), (1,1), (1,2), (2,0), (2,1), (2,2)
```

This is identical to:

```python
pairs = []
for x in range(3):
    for y in range(3):
        pairs.append((x, y))
```

Real example — flatten a list of lists:

```python
flat = [item for row in grid for item in row]
```

The leftmost `for` is the outer loop. Beginners often write the clauses
in the wrong order; reach for the explicit-loop form when in doubt.

## Dict and set comprehensions

The same syntax with different brackets:

```python
{w: len(w) for w in words}              # dict
{w.lower() for w in words}              # set (no duplicates)
{n * 2 for n in nums}                   # set
```

Use a set comp when "you want each distinct value once" — that's O(n)
where the loop+`if not in` form is O(n^2).

## When NOT to use a comprehension

Comprehensions optimise for one thing: **building a new collection**. If
you're building anything else, use a loop:

```python
# BAD: comprehension for side effects
[print(x) for x in nums]   # builds and discards a list of Nones

# GOOD:
for x in nums:
    print(x)
```

Skip comprehensions when:

- you're not collecting results (use a `for` loop),
- the expression is several lines or has a try/except (use a `for` loop),
- multiple `if`/`else` clauses make it unreadable (use a `for` loop).

CS 106A graders mark "clever" comprehensions down when readability suffers.

## Generator expressions

Same syntax with parentheses produces a *generator* — lazy, doesn't build
the list in memory:

```python
total = sum(x * x for x in range(10**6))   # streams; no million-element list
```

Use this for `sum`, `min`, `max`, `any`, `all`, `next`. The parens around
the genexp can be omitted when it's the only argument to a function:

```python
any(n < 0 for n in nums)
all(s.startswith("p") for s in words)
```

## The `walrus` operator (Python 3.8+)

Occasionally useful to avoid recomputing in a comprehension:

```python
# Without walrus: f(x) computed twice
out = [f(x) for x in xs if f(x) > 0]

# With walrus: computed once, captured into y
out = [y for x in xs if (y := f(x)) > 0]
```

CS 106A doesn't require the walrus — it's a 'nice to know'.

## Performance and clarity

Comprehensions are typically faster than the equivalent `for`/`append`
loop because the bytecode skips the method-lookup overhead on each
iteration. That said, in CS 106A the win is **clarity**: a one-liner that
says what it does beats five lines that say how.

## Reach-for examples

```python
# Strip and lowercase every line
clean = [line.strip().lower() for line in raw_lines if line.strip()]

# Squared distance to origin for each point
dists = [x*x + y*y for x, y in points]

# Indices where flag is True
on = [i for i, flag in enumerate(flags) if flag]

# All vowels in a string
vowels = [c for c in text if c in "aeiou"]

# Word frequencies (dict comp + Counter)
from collections import Counter
freq = {w: c for w, c in Counter(words).most_common(10)}
```

For the underlying list operations, see {{GUIDE:cs106a-lists}}; for the
filter expression patterns under `if`, see {{GUIDE:cs106a-control-flow}}.

## Practice

Take {{QUIZ:cs106a-list-comprehensions-quiz}} once you can read the
"flatten" and "conditional expression" examples without slowing down.
The most-missed question is the `if` placement one — re-read the
**conditional expression** section if you flub it.
