---
slug: cs106a-control-flow
course:
  ipeds_id: "243744"
  department: "CS"
  number: "106A"
title: "Control Flow — if, while, for in CS 106A"
description: "Branching and looping patterns in Python with the gotchas CS 106A graders actually test."
tags: ["python", "control-flow", "loops", "conditionals", "midterm"]
author_role: bot
quiz_slug: cs106a-control-flow-quiz
attached_files:
  - stanford-cs106a-midterm1-review-slides
attached_resources: []
---

# Control Flow in Python

After you can evaluate expressions and bind names (see
{{GUIDE:cs106a-python-intro}}), everything interesting happens because the
program picks a *different* next line based on the state of the world.
That's control flow: conditionals (`if`/`elif`/`else`) and loops
(`while`, `for`). CS 106A builds pset after pset on top of these two
primitives, so if your intuition for them is fuzzy, fix it now.

## Indentation is syntax

Python does not have `{` and `}`. The body of every control structure is
defined by indentation. The standard is four spaces. Mixing tabs and
spaces is a fast way to get a `TabError` or — worse — invisible bugs
where the editor *renders* lines as aligned but the interpreter disagrees.

```python
if score >= 90:
    grade = "A"
    print("Great job")
else:
    grade = "B or below"
print("Done")   # NOT part of either branch
```

## if / elif / else

Exactly one branch runs. `elif` is tested only if everything above it was
false.

```python
def letter_grade(score: int) -> str:
    if score >= 90:
        return "A"
    elif score >= 80:
        return "B"
    elif score >= 70:
        return "C"
    else:
        return "F"
```

Pitfall: ordering matters. If you test `score >= 70` first, every passing
student gets a C. Always write the narrowest condition at the top.

### Truthiness

`if` evaluates its condition as a boolean. Python's truthiness rules are
worth memorising because they simplify a lot of code:

| Falsy | Everything else is truthy |
|---|---|
| `False`, `None` | |
| `0`, `0.0` | |
| `""`, `[]`, `{}`, `()`, `set()` | |

```python
if items:           # empty list is falsy; clearer than `len(items) > 0`
    for item in items:
        print(item)
```

CS 106A style expects this idiom over explicit `== 0` or `== None`
comparisons. The one exception: use `is None` / `is not None` to test for
`None` specifically. Don't write `x == None`.

## while loops

A `while` loop repeats as long as its condition is true. The classic use
case is "loop until the user says stop":

```python
SENTINEL = -1
total = 0
value = int(input("Enter a number (-1 to stop): "))
while value != SENTINEL:
    total += value
    value = int(input("Enter a number (-1 to stop): "))
print("Sum:", total)
```

Two rules to internalise:

1. The loop variable must change inside the loop, or you'll spin forever.
2. The sentinel value must be impossible as legitimate input. `-1` is fine
   for "count of cookies"; it's a bug for "temperature in Celsius".

### Infinite loops with break

Sometimes the exit condition is clearest at a specific point inside the
loop body. An infinite loop with `break` is idiomatic and considered
cleaner than priming-read patterns once you're comfortable with it:

```python
while True:
    line = input("> ")
    if line == "quit":
        break
    handle(line)
```

## for loops

`for` iterates over any iterable — a list, string, range, dict, etc. This
is different from C-style `for (int i = 0; i < n; i++)` and much harder
to get wrong.

```python
for letter in "Python":
    print(letter)

for i in range(5):        # 0, 1, 2, 3, 4
    print(i)

for i in range(2, 10, 2): # 2, 4, 6, 8
    print(i)
```

`range` is half-open: `range(n)` stops *before* `n`. Off-by-one bugs
usually come from forgetting this.

### for with index + value

When you need both the position and the value, use `enumerate`:

```python
names = ["Ada", "Grace", "Margaret"]
for i, name in enumerate(names):
    print(f"{i}: {name}")
```

Do **not** write `for i in range(len(names)): name = names[i]`. Graders
will call it out, and `enumerate` is clearer anyway.

### Iterating and modifying

Never add to or delete from a list while iterating over it:

```python
nums = [1, 2, 3, 4]
for n in nums:
    if n % 2 == 0:
        nums.remove(n)   # BUG: skips elements
```

Iterate over a copy, or better, build a new list with a comprehension —
see {{GUIDE:cs106a-list-comprehensions}}.

## break and continue

- `break` exits the enclosing loop immediately.
- `continue` skips to the next iteration.

```python
# Print every non-comment, non-blank line
for line in lines:
    if not line.strip():
        continue
    if line.startswith("#"):
        continue
    print(line)
```

Avoid nesting `continue` three levels deep — at that point the logic is
usually clearer as a helper function.

## Nested loops

The time complexity multiplies, so nested loops are fine for small inputs
but deadly for large ones. The quintessential CS 106A nested pattern is
printing a grid:

```python
def print_grid(rows: int, cols: int) -> None:
    for r in range(rows):
        for c in range(cols):
            print("*", end=" ")
        print()   # newline at end of row
```

The `end=" "` keyword argument is the usual trick — it prevents `print`
from appending its default newline.

## Loop-else

Python has a little-known `for`/`else` and `while`/`else`. The `else`
runs if the loop completed without hitting `break`. It's occasionally
useful for search patterns:

```python
for item in items:
    if item == target:
        print("found")
        break
else:
    print("not found")
```

Not required in CS 106A, but if you see it in staff code, now you know.

## Midterm patterns

The midterm review slides at
`{{FILE:stanford-cs106a-midterm1-review-slides}}` call out three archetypes:

- **Counter**: initialise `count = 0`, increment inside the loop, use it
  outside.
- **Accumulator**: initialise `total = 0` (or `""` or `[]`), append/add
  each iteration.
- **Early exit**: use `break` when the question is satisfied.

Almost every lab problem is one of those three with a slightly different
decision rule inside the loop body.

## Practice

When the idioms feel natural, take {{QUIZ:cs106a-control-flow-quiz}}. Pay
attention to the `range` and `break` questions — those are the two most
frequently missed.
