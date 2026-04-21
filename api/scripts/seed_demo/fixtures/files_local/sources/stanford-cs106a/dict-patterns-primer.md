---
slug: stanford-cs106a-dict-patterns-primer
title: "Dict Patterns Primer: Counter, Grouping, and Memoization"
mime: application/epub+zip
filename: dict-patterns-primer.epub
course: stanford/cs106a
description: "Primer on three high-leverage dict patterns in Python: counting, grouping, and memoization caches."
author_role: bot
---

# Dict Patterns Primer

Three dictionary patterns show up in nearly every Python program you'll write this quarter: **counting**, **grouping**, and **memoization**. Master these and you'll reach for dicts reflexively.

## Pattern 1: Counter

Counting how often each item appears.

### The Hand-Rolled Version

```python
def count_letters(text: str) -> dict[str, int]:
    counts = {}
    for ch in text.lower():
        if ch.isalpha():
            counts[ch] = counts.get(ch, 0) + 1
    return counts

count_letters("Hello")  # {'h': 1, 'e': 1, 'l': 2, 'o': 1}
```

`dict.get(key, default)` returns `default` when the key is missing — no `KeyError`, no branching.

### The Idiomatic Version

```python
from collections import Counter

counts = Counter(text.lower())
counts.most_common(3)   # [('e', 12), ('t', 9), ('a', 7)]
```

`Counter` is just a dict with extras. Use it unless the assignment forbids it.

## Pattern 2: Grouping

Bucketing items by some shared key.

```python
from collections import defaultdict

def group_by_length(words: list[str]) -> dict[int, list[str]]:
    groups = defaultdict(list)
    for w in words:
        groups[len(w)].append(w)
    return dict(groups)

group_by_length(["hi", "bye", "ok", "yep"])
# {2: ['hi', 'ok'], 3: ['bye', 'yep']}
```

`defaultdict(list)` creates `[]` the first time each key is touched, so `.append` always works.

Without `defaultdict` you'd write:

```python
groups = {}
for w in words:
    groups.setdefault(len(w), []).append(w)
```

Both are fine. Pick one style and stick with it.

## Pattern 3: Memoization

Caching expensive function results keyed by input.

```python
def fib(n: int, cache: dict[int, int] = {}) -> int:
    if n in cache:
        return cache[n]
    if n < 2:
        result = n
    else:
        result = fib(n - 1) + fib(n - 2)
    cache[n] = result
    return result
```

> Heads-up: using a mutable default argument like `cache={}` **works** but is a known footgun. The standard library has a cleaner version:

```python
from functools import lru_cache

@lru_cache(maxsize=None)
def fib(n: int) -> int:
    return n if n < 2 else fib(n - 1) + fib(n - 2)
```

With memoization, `fib(100)` returns instantly. Without it, the call tree explodes.

## Comparing the Three

| Pattern       | Typical key         | Typical value | Stdlib helper          |
|---------------|---------------------|---------------|------------------------|
| Counter       | the item itself     | integer count | `collections.Counter`  |
| Grouping      | a category function | list of items | `collections.defaultdict(list)` |
| Memoization   | function arguments  | computed answer| `functools.lru_cache` |

## A Worked Example: Word Frequencies by Length

Combine counter + grouping in one pass.

```python
from collections import Counter, defaultdict

def freq_by_length(text: str) -> dict[int, Counter]:
    buckets: dict[int, Counter] = defaultdict(Counter)
    for word in text.lower().split():
        buckets[len(word)][word] += 1
    return dict(buckets)
```

For `"the cat sat on the mat"`, you get `{3: Counter({'the': 2, 'cat': 1, 'sat': 1, 'mat': 1}), 2: Counter({'on': 1})}`.

## Checklist

- [ ] Reach for `dict.get()` or `defaultdict` before writing `if key in d: ...`.
- [ ] Prefer `Counter` over hand-rolled counting unless the spec forbids imports.
- [ ] Cache with `@lru_cache`, not mutable default args.
- [ ] Remember: dict keys must be hashable (strings, numbers, tuples — not lists).

These three patterns cover 80% of dict work you'll do for the rest of 106A.
