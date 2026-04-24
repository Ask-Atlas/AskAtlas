---
slug: cs106a-midterm-review
course:
  ipeds_id: "243744"
  department: "CS"
  number: "106A"
title: "Midterm Review — CS 106A"
description: "Consolidated review for the CS 106A midterm: types, control flow, functions, lists, dicts, and recursion in one place."
tags: ["python", "midterm", "review", "exam-prep", "consolidated"]
author_role: bot
attached_files:
  - stanford-cs106a-midterm1-review-slides
attached_resources: []
---

# Midterm Review

The CS 106A midterm covers everything from the first half of the quarter:
types and expressions, control flow, functions and scope, lists and
dictionaries, and (depending on the year) early recursion. Each topic has
a dedicated guide; this is the consolidated cram sheet that ties them
together with the highest-yield practice.

The official review slides live at
`{{FILE:stanford-cs106a-midterm1-review-slides}}` — pair them with this
guide.

## How the exam is graded

Past midterms historically allocate roughly:

- 25% short-answer / "what does this print" trace problems
- 35% small functions to write end-to-end
- 30% one larger problem that combines two or three topics
- 10% one recursion or string-manipulation question

Partial credit is generous when your code is structured. Sloppy code that
happens to work scores worse than clean code that's slightly buggy.

## Topic-by-topic

### Types and expressions

Covered in {{GUIDE:cs106a-python-intro}}.

The two trace-problem traps:

- `int / int` is **always** float in Python 3. `7 / 2 == 3.5`, not 3.
- `int * str` repeats: `3 * "ab" == "ababab"`.
- `bool` IS an `int` subtype. `True + 1 == 2`. Almost never useful, but
  occasionally appears on trace problems.

### Control flow

Covered in {{GUIDE:cs106a-control-flow}}.

The patterns the exam tests:

- `range(n)` produces 0..n-1, **not** 1..n.
- `for x in xs: x = ...` does not modify `xs`. To mutate, index by
  position.
- A `while True: ... break` loop is fine, idiomatic, and often clearer
  than priming reads.

### Functions and scope

Covered in {{GUIDE:cs106a-functions-and-scope}}.

Likely trace question: nested function with the same name as a global.

```python
x = 1
def outer():
    x = 2
    def inner():
        return x        # which x?
    return inner()
print(outer())   # prints 2 (LEGB: enclosing wins)
```

Likely write-it question: a helper that takes a list, returns a tuple of
two things (e.g., `min` and `max`, or `count` and `sum`).

### Lists

Covered in {{GUIDE:cs106a-lists}}.

The exam tests:

- The difference between `nums.sort()` and `sorted(nums)`.
- Slicing producing a copy: `b = a[:]` does not alias.
- The 2D-grid trap: `[[0]*3]*3` is broken; use a comprehension.

### Dictionaries

Covered in {{GUIDE:cs106a-dictionaries}}.

The two patterns to memorise:

```python
# Counting
counts: dict[str, int] = {}
for word in words:
    counts[word] = counts.get(word, 0) + 1

# Grouping
groups: dict[str, list[str]] = {}
for word in words:
    key = word[0]
    groups.setdefault(key, []).append(word)
```

### List comprehensions

Covered in {{GUIDE:cs106a-list-comprehensions}}.

Mnemonic: **`if` after `for` filters; `if/else` before `for` transforms.**

```python
[n for n in nums if n > 0]      # filter
[n if n > 0 else 0 for n in nums] # transform
```

### Recursion (if covered before midterm)

Covered in {{GUIDE:cs106a-recursion}}.

Always answer the two-piece template:

1. What's the **base case**? When can I answer without recursing?
2. What's the **recursive step**? How do I make the problem smaller and
   combine the result?

If the exam asks "what's wrong with this recursive function", the
answer is almost always either a missing base case or a recursive call
that doesn't shrink the input.

## Trace problem strategy

When asked "what does this print?":

1. Draw a table of variable names down the left and time across the top.
2. Update one row per executed statement.
3. Mark the lines `print` runs against.

Trying to do it in your head causes more wrong answers on this section
than any other mistake.

## Code-writing strategy

When asked to write a function:

1. **Write the docstring first** — forces you to articulate inputs and
   output before you commit to syntax.
2. **Handle the easy cases first** — empty list, n == 0, etc. Often
   these are most of the test cases.
3. **Make a small example, work it by hand** — turn that work into the
   loop body.
4. **Ask: does this work for size 0, size 1, size n?** Off-by-one bugs
   live at the boundaries.

## Common-mistake catalog

- `lst.append(x)` returns `None`. Don't write `lst = lst.append(x)`.
- `for i in range(len(lst)):` should usually be `for x in lst:` or
  `for i, x in enumerate(lst):`.
- `if x = 5:` is a `SyntaxError`. Use `==` for comparison.
- `dict[key]` raises `KeyError` when key is missing; use `.get()` or
  check `in`.
- Mutable default args. See {{GUIDE:cs106a-functions-and-scope}}.

## Day-of advice

- Sleep beats one more cram hour.
- Bring water, scratch paper, and a pencil that won't smudge.
- If you blank, skip it and come back. Coming back to a problem with a
  fresh head almost always finds something you missed.
- Triple-check your code reads variables you actually defined. Half of
  the partial-credit losses are typo'd variable names.

For a final, debug-oriented sweep, also re-read
{{GUIDE:cs106a-debugging-techniques}}.
