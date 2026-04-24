---
slug: stanford-cs106a-midterm1-review-slides
title: "Midterm 1 Review Slides"
mime: application/vnd.openxmlformats-officedocument.presentationml.presentation
filename: midterm1-review-slides.pptx
course: stanford/cs106a
description: "Slide deck reviewing expressions, control flow, functions, strings, and lists for CS 106A Midterm 1."
author_role: bot
---

# Midterm 1 Review

A fast pass over the first four weeks. Each section below is one slide.

## What's On the Midterm

- Variables, types, and expressions
- Control flow: `if`, `while`, `for`
- Functions: parameters, return, scope
- Strings and indexing
- Lists: building, iterating, mutating
- Tracing and debugging

## Types and Truthiness

```python
bool(0), bool(""), bool([]), bool(None)   # all False
bool(1), bool("x"), bool([0])             # all True
```

Only those five "falsy" values exist. Everything else is truthy.

## Integer vs. Float Division

```python
7 / 2    # 3.5   (true division, always float)
7 // 2   # 3     (floor division)
7 % 2    # 1     (remainder)
```

## `if` / `elif` / `else`

```python
if score >= 90:
    grade = "A"
elif score >= 80:
    grade = "B"
else:
    grade = "C"
```

Order matters — the first true branch wins.

## `while` Loops

Use when you don't know the count in advance.

```python
guess = ""
while guess != "quit":
    guess = input("> ")
```

## `for` Loops and `range`

```python
for i in range(5):           # 0..4
for i in range(2, 10):       # 2..9
for i in range(10, 0, -1):   # 10..1 (countdown)
```

## Functions: The Shape

```python
def area(width: float, height: float) -> float:
    """Return the area of a rectangle."""
    return width * height
```

- Parameters are local to the function.
- `return` exits immediately.
- A function with no `return` returns `None`.

## Scope in One Picture

Local variables shadow globals. Assignment inside a function creates a new local unless `global` is used (don't use `global`).

```python
x = 10
def f():
    x = 5       # new local x
    print(x)    # 5
f()
print(x)        # 10
```

## Strings Are Immutable

```python
s = "hello"
s[0] = "H"      # TypeError
s = "H" + s[1:] # OK: new string
```

## String Methods to Know

```python
"Hi".lower()       # "hi"
"  x  ".strip()    # "x"
"a,b,c".split(",") # ["a", "b", "c"]
",".join(["a","b"])# "a,b"
"abc".find("b")    # 1
```

## Lists: Mutable Sequences

```python
xs = [3, 1, 4]
xs.append(1)       # [3, 1, 4, 1]
xs.sort()          # [1, 1, 3, 4]
xs[0] = 99         # [99, 1, 3, 4]
len(xs)            # 4
```

## Aliasing Gotcha

```python
a = [1, 2, 3]
b = a              # same list!
b.append(4)
print(a)           # [1, 2, 3, 4]
```

Use `b = a.copy()` or `b = list(a)` to get a fresh copy.

## Tracing Practice

```python
def mystery(n):
    total = 0
    for i in range(1, n + 1):
        if i % 2 == 0:
            total += i
    return total

mystery(6)   # 2 + 4 + 6 = 12
```

Trace on paper. Points are lost when students skip tracing.

## Debugging Strategy

1. Read the error message top to bottom.
2. Add `print()` for every variable you suspect.
3. Test the smallest input that still fails.
4. Re-read the problem statement — most "bugs" are misread specs.

## Night Before Checklist

- [ ] Re-skim lecture slides 1-12
- [ ] Re-do Assignments 1 and 2 from scratch on paper
- [ ] Trace three past-midterm problems
- [ ] Sleep. Caffeinated panic is not a strategy.

Good luck!
