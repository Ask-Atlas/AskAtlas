---
slug: cs106a-dictionaries
course:
  ipeds_id: "243744"
  department: "CS"
  number: "106A"
title: "Dictionaries — CS 106A"
description: "Key-value mappings in Python: lookup, update, iteration, and the patterns that make assignments tractable."
tags: ["python", "dictionaries", "data-structures", "midterm"]
author_role: bot
attached_files:
  - stanford-cs106a-dict-patterns-primer
attached_resources: []
---

# Dictionaries

A `dict` maps keys to values. Where a `list` answers "what's at position
3?", a dict answers "what's the value for the key 'fig'?". Once you start
seeing problems through that lens, half the assignments in CS 106A
collapse from twenty lines of nested loops to four lines of dict updates.

The patterns primer at `{{FILE:stanford-cs106a-dict-patterns-primer}}`
catalogs the half-dozen recipes you reach for again and again.

## Construction

```python
empty: dict[str, int] = {}
ages = {"Ada": 36, "Grace": 85, "Margaret": 80}
pairs = dict([("a", 1), ("b", 2)])     # from a list of tuples
zeros = dict.fromkeys(["a", "b", "c"], 0)   # {'a': 0, 'b': 0, 'c': 0}
```

Trailing commas inside `{}` are fine and recommended for multi-line
literals — they make diffs cleaner.

## Lookup

```python
ages["Ada"]              # 36
ages["nobody"]           # KeyError: 'nobody'
ages.get("nobody")       # None
ages.get("nobody", 0)    # 0   - default if missing
"Ada" in ages            # True
"Ada" not in ages        # False
```

The `.get(key, default)` form is one of the most useful methods in
Python. Use it whenever "key might be missing, in which case I want a
sensible fallback" describes your situation.

## Insert / update / delete

```python
ages["Linus"] = 54       # insert (or update if present)
ages["Ada"] = 37         # update
del ages["Grace"]        # delete (KeyError if missing)
ages.pop("Ada")          # delete + return value
ages.pop("nobody", None) # delete if present, default if not
```

A dict literal with the same key written twice keeps the **last** value.
This is occasionally a source of silent bugs in hand-typed lookup tables.

## Iteration

```python
for key in ages:               # iterates KEYS
    print(key, ages[key])

for key in ages.keys():        # explicit; equivalent to the above
    handle(key)

for value in ages.values():    # iterates VALUES
    handle(value)

for key, value in ages.items(): # iterates pairs
    print(f"{key}: {value}")
```

`for k in d` iterates keys, not pairs. The `.items()` form is the one you
want most often. CS 106A graders flag `for k in d.keys(): v = d[k]` as a
code smell — use `.items()`.

### Iteration order

Since Python 3.7, dicts preserve insertion order. CS 106A is taught on a
modern Python so you can rely on that. **Do not** mutate the dict (add
or delete keys) while iterating; mutating *values* is fine.

## Counting pattern

The single most common dict pattern in CS 106A:

```python
counts: dict[str, int] = {}
for word in text.split():
    if word in counts:
        counts[word] += 1
    else:
        counts[word] = 1
```

The clean idiom uses `.get`:

```python
counts: dict[str, int] = {}
for word in text.split():
    counts[word] = counts.get(word, 0) + 1
```

The cleanest uses `collections.Counter`:

```python
from collections import Counter
counts = Counter(text.split())
counts.most_common(3)   # 3 most common words
```

`Counter` is a `dict` subclass; everything you know about `dict` works
on it. It's what graders expect for problems explicitly about counting.

## Grouping pattern

When you want a dict from key to *list* of values:

```python
buckets: dict[str, list[int]] = {}
for n in nums:
    bucket = "even" if n % 2 == 0 else "odd"
    if bucket not in buckets:
        buckets[bucket] = []
    buckets[bucket].append(n)
```

The cleaner version uses `setdefault`:

```python
buckets: dict[str, list[int]] = {}
for n in nums:
    bucket = "even" if n % 2 == 0 else "odd"
    buckets.setdefault(bucket, []).append(n)
```

Or `collections.defaultdict`:

```python
from collections import defaultdict
buckets: dict[str, list[int]] = defaultdict(list)
for n in nums:
    bucket = "even" if n % 2 == 0 else "odd"
    buckets[bucket].append(n)
```

## Inverting a dict

When you want value-to-key:

```python
ages = {"Ada": 36, "Grace": 85}
by_age = {age: name for name, age in ages.items()}
# {36: 'Ada', 85: 'Grace'}
```

That's a *dict comprehension*. If multiple keys share a value, the later
one wins. If you need to keep all keys per value, use the grouping pattern
above instead.

## Keys must be hashable

A key has to be **hashable** — its hash value must not change for the
life of the dict. Practically:

| Hashable | Not hashable |
|---|---|
| `int`, `float`, `str`, `bool` | `list`, `dict`, `set` |
| `tuple` of hashables | `tuple` containing a list |
| `frozenset` | `set` |

Trying `d[[1, 2]] = 3` raises `TypeError: unhashable type: 'list'`. If you
need a "list-like" key, convert to a tuple: `d[(1, 2)] = 3`.

## Nested dicts

A dict whose values are themselves dicts handles many "table of records"
problems:

```python
students = {
    "ada": {"grade": "A", "units": 5},
    "grace": {"grade": "B", "units": 4},
}
students["ada"]["grade"]    # 'A'
```

For deeper nesting, consider a class — see {{GUIDE:cs106a-oop-classes}}.

## Common pitfalls

- **`KeyError` on read**: use `.get()` or check `in` first.
- **Mutating during iteration**: `RuntimeError: dictionary changed
  size during iteration`. Iterate over `list(d.keys())` if you must
  delete inside the loop.
- **List as key**: `TypeError: unhashable type`. Convert to tuple.
- **Counting with `+=` on missing key**: `KeyError`. Initialise to 0
  first or use `Counter`.

## Practice

When the grouping and counting patterns feel automatic, jump back to
{{GUIDE:cs106a-list-comprehensions}} for the comprehension form, then
{{GUIDE:cs106a-debugging-techniques}} for how to introspect a malformed
dict in pdb.
