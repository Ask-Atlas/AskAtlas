---
slug: cs106a-debugging-techniques
course:
  ipeds_id: "243744"
  department: "CS"
  number: "106A"
title: "Debugging Techniques — CS 106A"
description: "Print debugging, the pdb interactive debugger, reading tracebacks, and minimal-reproducible-example discipline."
tags: ["python", "debugging", "pdb", "tracebacks", "methodology"]
author_role: bot
quiz_slug: cs106a-debugging-techniques-quiz
attached_files:
  - stanford-cs106a-debugging-techniques
attached_resources: []
---

# Debugging

Most of the time you spend on CS 106A psets is debugging, not writing
new code. The students who get 100% are not the ones who write the most
correct code on the first try — they're the ones who diagnose mistakes
fastest. This guide is the toolbox.

The full deep-dive lives in
`{{FILE:stanford-cs106a-debugging-techniques}}`; this is the working
checklist.

## Read the traceback bottom-up

A Python traceback looks like this:

```
Traceback (most recent call last):
  File "game.py", line 42, in <module>
    play(board)
  File "game.py", line 17, in play
    move = best_move(board)
  File "game.py", line 8, in best_move
    return scores[max(scores)]
KeyError: 9
```

Read it **bottom-up**:

1. The error type and message tell you what went wrong: `KeyError: 9`
   means "tried to look up the key 9 in a dict, and it wasn't there".
2. The line just above tells you exactly where: `game.py:8`, inside
   `best_move`.
3. The lines further up tell you how you got there. Useful when the bug
   is data-dependent.

90% of the time, the bottom three lines are all you need.

## Common error types

| Error | Usual cause |
|---|---|
| `SyntaxError` | typo before runtime — Python won't even start |
| `IndentationError` | mixed tabs/spaces or wrong nesting |
| `NameError` | using a variable that wasn't defined / typo'd name |
| `TypeError` | wrong operation for the type (`"a" + 1`) |
| `ValueError` | right type, wrong content (`int("five")`) |
| `IndexError` | `lst[i]` where `i` is out of range |
| `KeyError` | `d[k]` where `k` not in dict |
| `AttributeError` | `obj.foo` where `obj` has no `foo` |
| `ZeroDivisionError` | `x / 0` |
| `RecursionError` | infinite recursion — usually a missing base case |

Most of these are listed because you'll cause every one of them at least
once.

## Print debugging

The fastest debugger is still `print`:

```python
def best_move(board):
    scores = score_each_move(board)
    print("DEBUG scores =", scores)
    print("DEBUG max(scores) =", max(scores))
    return scores[max(scores)]
```

Three habits that make print debugging effective:

1. **Print the variable name AND the value**: `print("scores =", scores)`,
   not just `print(scores)`. Five lines deep in output you'll thank
   yourself.
2. **Tag debug prints**: `print("DEBUG ...")` makes them grep-removable.
3. **Print at boundaries**: function entry and exit, before and after
   loops, around the suspected bug.

Remove them when you're done — leaving stray `print` statements in
submitted code is a CS 106A style deduction.

## The `pdb` interactive debugger

For bugs you can't reason out from prints, drop `breakpoint()` into your
code:

```python
def best_move(board):
    scores = score_each_move(board)
    breakpoint()           # interpreter stops here
    return scores[max(scores)]
```

Run the script normally. When execution hits `breakpoint()`, you get a
`(Pdb)` prompt. The commands you'll use 90% of the time:

| Command | Effect |
|---|---|
| `n` | next line (step over) |
| `s` | step into a function call |
| `c` | continue until next breakpoint |
| `l` | list source around current line |
| `p expr` | print the value of an expression |
| `pp expr` | pretty-print (multi-line dict / list) |
| `w` | where — print the call stack |
| `u` / `d` | move up / down the stack |
| `q` | quit the program |

Type the variable name itself (no `p`) and pdb prints it. `Enter` repeats
the previous command — useful with `n` to walk forward.

## Minimal reproducible example

When you can't figure it out alone, the act of stripping the bug down
into the smallest reproducer often *finds* the bug. The discipline:

1. Copy the failing function into a fresh file.
2. Hard-code the input that triggers the bug.
3. Remove every line that's not necessary to reproduce the failure.
4. If the bug now goes away, the line you just removed was relevant.

By the time you're at three lines of code and one input value, you'll
either see the bug or have a tight question to ask office hours.

## Hypothesis-driven debugging

When the print/pdb output doesn't immediately reveal the bug, fall back
to the scientific method:

1. **Form a hypothesis**: "I think the off-by-one happens because
   `range(len(words))` includes one extra element."
2. **Pick a check**: "If true, then `len(words)` should equal the index
   that fails."
3. **Run the check**: print or pdb the value.
4. **Update the hypothesis** based on what you observe.

Two iterations of this beats two hours of staring at the file.

## Type-introspection in the REPL

When you're not sure what something is, ask:

```python
type(x)              # <class 'list'>
isinstance(x, list)  # True
dir(x)               # all attributes of x - useful for "what method does this object have?"
help(str.split)      # docstring of any method
```

CS 106A graders use `help` and `dir` constantly. Internalising them is
much faster than searching online for the same information.

## Assertions

`assert` checks an invariant and raises `AssertionError` if it fails:

```python
def average(nums: list[float]) -> float:
    assert nums, "average() requires a non-empty list"
    return sum(nums) / len(nums)
```

CS 106A allows assertions for "this should never happen" checks. Don't
use them for user-input validation — assertions can be disabled at
runtime.

## When to ask for help

After:

- you've read the bottom three lines of the traceback and looked at the
  named line,
- you've added two or three `print`s to inspect intermediate state,
- you've stripped to a minimal repro,

…then post in the course Ed thread or visit office hours. Including the
minimal repro and your current hypothesis turns a 30-minute debugging
session into a 30-second answer.

## Practice

Take {{QUIZ:cs106a-debugging-techniques-quiz}}. The most-missed question
is the one about reading tracebacks; if you flub it, re-read the
**Read the traceback bottom-up** section.

For style and structure conventions that prevent bugs in the first
place, see {{GUIDE:cs106a-functions-and-scope}}.
