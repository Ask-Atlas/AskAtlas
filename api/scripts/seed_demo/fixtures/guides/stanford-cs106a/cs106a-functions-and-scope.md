---
slug: cs106a-functions-and-scope
course:
  ipeds_id: "243744"
  department: "CS"
  number: "106A"
title: "Functions and Scope — CS 106A"
description: "Defining functions, parameters, return values, and the LEGB scope rules in Python."
tags: ["python", "functions", "scope", "decomposition"]
author_role: bot
attached_files:
  - stanford-cs106a-pset-setup-readme
attached_resources: []
---

# Functions and Scope

Decomposition is the headline skill of CS 106A. The mechanism for
decomposition in Python is the **function**. A well-named function is a
contract: "give me these inputs, I'll produce that output, and you don't
need to read my body to use me." Get good at writing them and the rest of
the course gets easier.

## Defining a function

```python
def square(n: int) -> int:
    """Return n squared."""
    return n * n
```

The pieces:

- `def` introduces the definition.
- `square` is the name. Use `snake_case` and a verb-noun shape.
- `(n: int)` is the parameter list with a type hint.
- `-> int` is the return-type hint.
- The triple-quoted string on the first line of the body is the
  **docstring**. CS 106A graders read these.
- `return` produces the output. A function without an explicit `return`
  returns `None`.

## Calling a function

```python
y = square(7)        # y is 49
print(square(3))     # prints 9
```

A function call evaluates its arguments left-to-right, then runs the body
with each parameter bound to the corresponding argument value.

## Positional and keyword arguments

You can pass arguments by position or by keyword:

```python
def greet(name: str, greeting: str = "Hello") -> str:
    return f"{greeting}, {name}!"

greet("Ada")                       # "Hello, Ada!"
greet("Ada", "Howdy")              # "Howdy, Ada!"
greet(name="Ada", greeting="Hi")   # "Hi, Ada!"
```

`greeting="Hello"` is a *default value*. Defaults are evaluated **once**,
at function-definition time. This causes the most-classic Python footgun:

```python
def add_to(item, bag=[]):    # BUG: shared mutable default!
    bag.append(item)
    return bag

add_to(1)   # [1]
add_to(2)   # [1, 2]   - same list every call
```

Fix: use `None` as the sentinel and create a new list inside.

```python
def add_to(item, bag=None):
    if bag is None:
        bag = []
    bag.append(item)
    return bag
```

## Multiple return values

Python returns a single object, but that object can be a tuple, which
looks like multiple returns:

```python
def min_max(nums: list[int]) -> tuple[int, int]:
    return min(nums), max(nums)

lo, hi = min_max([3, 1, 4, 1, 5, 9])
```

Tuple unpacking on the call site is the idiom. Don't return a list of
length 2 unless the elements are genuinely homogeneous.

## Scope: LEGB

Names in Python resolve in four nested scopes, searched in order:

1. **L**ocal — the current function's local names.
2. **E**nclosing — outer function in a nested-function setup.
3. **G**lobal — the module's top-level names.
4. **B**uiltin — `print`, `len`, `range`, etc.

```python
x = 10                # global

def outer():
    x = 20            # enclosing (relative to inner)
    def inner():
        x = 30        # local
        print(x)      # 30
    inner()
    print(x)          # 20
outer()
print(x)              # 10
```

A bare assignment inside a function creates a **local** name. To rebind
a global, you must say so:

```python
counter = 0
def tick():
    global counter
    counter += 1
```

CS 106A discourages `global`. Pass values in, return them out. Threading
state through parameters is more boring but easier to debug than spooky
action at a distance.

### Reading vs assigning

You can *read* a global name from inside a function without `global`:

```python
PI = 3.14159
def area(r: float) -> float:
    return PI * r * r    # PI is found via the G in LEGB
```

You only need `global` when you intend to **rebind** it. If you just call
methods on a mutable global (e.g., `lst.append(x)` on a global `lst`),
you're mutating, not rebinding — no `global` needed. CS 106A graders
notice this distinction.

## Pure vs side-effecting

A *pure* function depends only on its inputs and has no side effects.
Pure functions are easier to test, reason about, and reuse:

```python
def double(n: int) -> int:
    return n * 2          # pure
```

Side-effecting functions read or write outside state — printing, file I/O,
mutating a shared list. Both are necessary, but try to keep the I/O at
the edges of your program and the logic in the pure middle.

## Parameter mutability

Python passes references. If you mutate the object the parameter points
at, the caller sees the mutation. If you rebind the name, you don't.

```python
def append_one(xs: list[int]) -> None:
    xs.append(1)         # caller sees this

def reset(xs: list[int]) -> None:
    xs = []              # rebinds local; caller does NOT see this
```

This catches everyone the first time. The standard fix when you want a
"new" version is to *return* it:

```python
def reset(xs: list[int]) -> list[int]:
    return []
```

## Decomposition heuristics

The Stanford grading rubric historically rewards code that:

- has each function under ~30 lines,
- has docstrings on every function,
- avoids `global`,
- uses verb names for actions and noun phrases for predicates
  (`is_valid`, `has_capacity`).

The pset setup readme `{{FILE:stanford-cs106a-pset-setup-readme}}` shows
the file structure graders expect.

## Common pitfalls

- **Forgetting `return`**: function returns `None`, caller gets a
  confusing `NoneType` error two lines later.
- **Returning inside a loop without iterating**: e.g., `for x in xs:
  return x` — only ever sees the first element.
- **Mutable default argument**: covered above; trips everyone.
- **Shadowing builtins**: `def list(...)` or `sum = 0; sum(...)` will
  redefine the builtin and cause a `TypeError` later.

## Practice

Functions don't have a dedicated quiz in this set, but the recursion guide
({{GUIDE:cs106a-recursion}}) and OOP guide ({{GUIDE:cs106a-oop-classes}})
both lean heavily on the concepts in this guide. Re-read the **scope**
section before tackling either.
