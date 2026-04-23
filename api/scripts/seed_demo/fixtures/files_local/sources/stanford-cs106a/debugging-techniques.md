---
slug: stanford-cs106a-debugging-techniques
title: "Debugging Techniques for CS 106A"
mime: application/pdf
filename: debugging-techniques.pdf
course: stanford/cs106a
description: "Practical debugging techniques for CS 106A: print tracing, tracebacks, interactive Python, and asking for help."
author_role: bot
---

# Debugging Techniques for CS 106A

Debugging is a skill, not a talent. Everybody — professors included — writes broken code on the first try. The question is how fast you can notice, localize, and fix the problem. Here are the moves that work, in the order to try them.

## Read the Traceback Bottom-Up

When Python throws, it prints a stack trace. The **bottom line** is the actual error; the lines above it are the call chain that led there.

```
Traceback (most recent call last):
  File "hangman.py", line 42, in play_round
    reveal(secret, guesses)
  File "hangman.py", line 17, in reveal
    return "".join(shown)
TypeError: sequence item 2: expected str instance, NoneType found
```

The useful pair: the error (`TypeError` at `reveal`, line 17) and the top user-code frame that called it (`play_round`, line 42). Open both and look at what crosses between them.

## Print Tracing, Done Correctly

A `print` is a cheap probe. The trick is printing the *right* thing.

```python
def reveal(secret, guesses):
    shown = [c if c in guesses else "_" for c in secret]
    print(f"[reveal] secret={secret!r} guesses={guesses!r} shown={shown!r}")
    return " ".join(shown)
```

Rules:

- Tag prints with the function name so grep works.
- Use `!r` (repr) so you see quotes, `None`, and empty strings.
- Print **inputs, branch choices, and return values** — not random midpoint state.

## The Scientific Method

1. State what you *expect* to see.
2. Run the code.
3. Compare expected to actual. The bug lives between the last point where they agree and the first point where they diverge.
4. Form one hypothesis, test it, repeat.

If you cannot state an expectation, you cannot debug — you are guessing.

## Binary Search in a File

If the program crashed at line 400, do not read 400 lines. Insert a `print("A")` at line 200 and a `print("B")` at line 300. Whichever prints *last* tells you which half contains the bug. Halve again. Five prints locates a bug in ~400 lines.

## The Interactive Python Sandbox

Open a Python REPL and paste the minimal broken snippet:

```
>>> reveal("cat", {"c"})
"c _ _"
>>> reveal("", {"c"})
""
```

Now you can vary inputs without re-running the whole program. This is faster than re-running `python hangman.py` forty times.

## Rubber-Duck Debugging

Explain your code aloud, line by line, to a non-talking object. You will catch your own bug ~60% of the time before finishing the explanation. The reason: writing and reading use different parts of your brain, and explanation forces you to justify assumptions.

## Common CS 106A Bugs

- **Off-by-one**: `range(1, n)` skips `0`; `range(1, n+1)` covers `1..n`.
- **Mutating a list while iterating**: use a copy (`for x in list(items):`).
- **`is` vs `==`**: `is` tests identity, `==` tests equality — you almost always want `==`.
- **Integer division**: `3 / 2` is `1.5`; `3 // 2` is `1`.
- **Forgetting `return`**: the function still runs, but returns `None`.

## When to Ask for Help

Ask after you have:

1. A **minimal example** (under 20 lines) that reproduces the bug.
2. A **traceback** or exact wrong output.
3. Your **hypothesis** for what might be wrong.

LaIR and Ed Discussion both prioritize students who show this work. Good questions get fast answers; vague ones get "what did you try?"
