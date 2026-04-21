---
slug: stanford-cs106a-list-comprehensions-cheatsheet
title: "List, Dict, and Set Comprehensions Cheatsheet"
mime: application/pdf
filename: list-comprehensions-cheatsheet.pdf
course: stanford/cs106a
description: "Quick-reference cheatsheet for Python list, dict, and set comprehensions with worked CS 106A examples."
author_role: bot
---

# List, Dict, and Set Comprehensions Cheatsheet

Comprehensions are Python's compact way to build collections from an iterable. They read left-to-right like English: *"give me this, for each of these, when this is true."*

## The Basic Shape

```python
# list
squares = [n * n for n in range(10)]

# dict
square_map = {n: n * n for n in range(10)}

# set (deduped)
unique_lengths = {len(w) for w in ["hi", "bye", "ok"]}
```

All three share the template: `[ expression  for item in iterable  if condition ]`.

## Filtering

Use an `if` clause to keep only some items.

```python
evens = [n for n in range(20) if n % 2 == 0]
short_names = [name for name in roster if len(name) < 5]
```

## Conditional Expressions (ternary)

If you want to *transform* based on a condition, put the `if/else` in the **expression**, not the filter slot.

```python
labels = ["even" if n % 2 == 0 else "odd" for n in range(6)]
# -> ['even', 'odd', 'even', 'odd', 'even', 'odd']
```

## Nested Loops

Multiple `for` clauses read top-down like nested loops.

```python
pairs = [(x, y) for x in range(3) for y in range(3) if x != y]
```

## Dict Comprehensions

Great for building lookup tables.

```python
scores = [("Ada", 95), ("Bo", 72), ("Cy", 88)]
grade_book = {name: score for name, score in scores}
passed = {name: score for name, score in scores if score >= 75}
```

## When to Use Which

| Goal                       | Pick a…           | Example                                |
|----------------------------|-------------------|----------------------------------------|
| Ordered sequence of items  | list comp         | `[w.upper() for w in words]`           |
| Key to value mapping       | dict comp         | `{w: len(w) for w in words}`           |
| Unique items, order-free   | set comp          | `{w.lower() for w in words}`           |
| Lazy / large / one-shot    | generator expr    | `sum(n*n for n in range(10_000_000))`  |

## Common Pitfalls

1. **Readability first.** If a comprehension needs two filters and a nested loop, a plain `for` loop is clearer.
2. **No side effects.** Don't use a comprehension just to call `print()` — that's a loop in disguise.
3. **Walrus (`:=`) sparingly.** Assignment inside a comprehension works in Python 3.8+ but can confuse readers.

## Practice Problems

```python
# 1. Build a list of all vowels in a sentence.
vowels = [ch for ch in sentence.lower() if ch in "aeiou"]

# 2. Word-length histogram as a dict.
hist = {w: len(w) for w in sentence.split()}

# 3. Unique first letters.
firsts = {w[0] for w in sentence.split() if w}
```

Keep this page beside you during Assignment 3 — most "loop-and-append" problems collapse into a one-liner.
