---
slug: cs106a-python-intro
course:
  ipeds_id: "243744"
  department: "CS"
  number: "106A"
title: "Programming Methodology in Python — CS 106A Intro"
description: "First contact with Python in CS 106A: values, variables, expressions, and the read-eval-print loop."
tags: ["python", "intro", "methodology", "syntax"]
author_role: bot
quiz_slug: cs106a-python-intro-quiz
attached_files:
  - stanford-cs106a-pset-setup-readme
attached_resources: []
---

# Programming Methodology in Python

CS 106A starts where most CS curricula start: teaching you to think about
problems in a way a computer can actually execute. Python is the tool we
use, but the methodology — decompose, name well, test early — is the
real subject of the course.

This guide covers the bare-bones primitives you need on Day 1: how to
run Python, what counts as a value, how variables work, and the half-dozen
operators you'll touch in every assignment for the rest of the quarter.

## Running Python

There are three ways you'll launch Python in CS 106A:

1. **The REPL** — type `python3` in a terminal. You get a `>>>` prompt
   and the interpreter evaluates one expression at a time. Excellent for
   poking at unfamiliar APIs.
2. **A script** — save code into `hello.py`, then run `python3 hello.py`.
   The interpreter executes top to bottom, then exits.
3. **A starter project** — most assignments ship as a folder with a
   pre-wired `main.py`. The setup readme at
   `{{FILE:stanford-cs106a-pset-setup-readme}}` walks through the
   recommended layout.

A *trivial* first program:

```python
print("Hello, CS 106A!")
```

`print` is a function. The parentheses call it; the string is the single
argument. Forgetting the parentheses gives you `<built-in function print>`
back, which is a common Day 1 confusion — Python is happy to evaluate the
function object itself and just show you what it is.

## Values and types

Every value in Python has a *type*. The five types you meet first:

| Type | Example literal | Notes |
|---|---|---|
| `int` | `42` | Arbitrary precision — no overflow |
| `float` | `3.14` | IEEE 754 double precision |
| `str` | `"hello"` or `'hello'` | Either quote style works |
| `bool` | `True`, `False` | Capitalised; `true` is a `NameError` |
| `NoneType` | `None` | The "no value" sentinel |

`type(x)` returns the type of `x`. Use it when you're debugging a confusing
value:

```python
x = "5"
type(x)        # <class 'str'>
x + 1          # TypeError: can only concatenate str (not "int") to str
```

That `TypeError` is one of the most common errors in CS 106A. Mix `str`
and `int` and Python refuses to guess what you meant. You have to convert
explicitly with `int(x)` or `str(1)`.

## Variables

A variable is a *name bound to a value*. Assignment uses a single `=`:

```python
quarter = "fall"
units = 5
gpa = 3.7
enrolled = True
```

Two important properties:

- **Names are not declared.** You don't say "this is an int". Just assign.
- **Names rebind freely.** `units = "five"` immediately after the snippet
  above is legal — `units` now refers to a string. This power can hide
  bugs; be intentional.

Names follow the convention `snake_case`. Constants are typically
`UPPER_SNAKE_CASE`. PEP 8 is the canonical Python style guide and CS 106A
graders generally expect it.

## Operators

Arithmetic is mostly what you'd expect, with two surprises:

```python
7 / 2     # 3.5    - true division, always returns float
7 // 2    # 3      - floor division
7 % 2     # 1      - modulo
2 ** 10   # 1024   - exponentiation
```

Floor division and modulo are crucial for problems involving wraparound
(clocks, grids, hashing). The `**` operator means "to the power of"; do
not write `2^10`, which is the bitwise XOR and gives `8`. That bug eats
hours every term.

String concatenation uses `+`, repetition uses `*`:

```python
"ab" + "cd"   # "abcd"
"-" * 10      # "----------"
```

Comparisons return booleans:

```python
3 < 5         # True
3 == 3.0      # True   - int and float compare equal when numerically equal
"a" < "b"     # True   - lexicographic
```

Boolean operators are spelled out in English: `and`, `or`, `not`. They
short-circuit, which is occasionally important:

```python
if x != 0 and total / x > 1:   # safe - division never runs when x == 0
    handle()
```

## Input and output

`print` accepts multiple arguments and joins them with a space:

```python
name = "Ada"
age = 36
print("Name:", name, "Age:", age)
# Name: Ada Age: 36
```

f-strings are the modern way to format:

```python
print(f"{name} is {age} years old")
```

`input(prompt)` reads a line from the user and returns a `str`. To get a
number you almost always need to convert:

```python
raw = input("How many cookies? ")
n = int(raw)
```

If the user types `"five"`, `int(raw)` raises `ValueError`. CS 106A
assignments usually wrap this in a `try`/`except` once you reach the
error-handling unit, but on Day 1 it's fine to let the program crash.

## The methodology piece

Karel-the-robot taught you decomposition; Python is where you apply it.
A function should do one thing. A name should describe what its value
represents, not how it's computed. Comments should explain *why*, not
*what* — the *what* is the code itself.

Two heuristics graders apply silently:

- If a function is over ~30 lines, suspect it's doing two things.
- If you copy three lines into a second place, extract a helper.

We'll build out functions and scope properly in
{{GUIDE:cs106a-functions-and-scope}}, and control flow in
{{GUIDE:cs106a-control-flow}}.

## Try it

Once you can run the REPL and a script, take {{QUIZ:cs106a-python-intro-quiz}}.
If the type-conversion question trips you up, re-read the **Values and types**
section before moving on — that confusion will haunt every later assignment.

For the broader course context jump to {{COURSE:stanford/cs106a}}.
