---
slug: cs106a-lists
course:
  ipeds_id: "243744"
  department: "CS"
  number: "106A"
title: "Lists in Python — CS 106A"
description: "Indexing, slicing, mutation, and the list operations CS 106A assignments rely on."
tags: ["python", "lists", "sequences", "indexing"]
author_role: bot
attached_files:
  - stanford-cs106a-pset-setup-readme
attached_resources: []
---

# Lists

The `list` is the workhorse data structure of CS 106A. If a problem asks
for "a sequence of things" and you don't have a strong reason to pick
something else, use a list.

## Construction

```python
empty: list[int] = []
nums = [1, 2, 3]
mixed = ["x", 4, True, None]    # heterogeneous; legal but rarely a good idea
zeros = [0] * 5                 # [0, 0, 0, 0, 0]
chars = list("hello")           # ['h', 'e', 'l', 'l', 'o']
```

`list` is mutable. The same object can shrink, grow, and have its
elements rebound in place. That's a feature and a footgun — see the
"aliasing" section below.

## Indexing

Indices start at 0. Negative indices count from the end:

```python
nums = [10, 20, 30, 40, 50]
nums[0]    # 10
nums[4]    # 50
nums[-1]   # 50
nums[-2]   # 40
nums[5]    # IndexError: list index out of range
```

`len(nums)` returns the count of elements. `len(nums) - 1` is the last
valid positive index. CS 106A psets are full of off-by-one bugs that
trace back to forgetting this.

## Slicing

A slice produces a *new* list:

```python
nums[1:4]    # [20, 30, 40]   - start inclusive, stop exclusive
nums[:3]     # [10, 20, 30]
nums[2:]     # [30, 40, 50]
nums[:]      # full copy
nums[::2]    # [10, 30, 50]   - every other element
nums[::-1]   # [50, 40, 30, 20, 10] - reversed
```

`nums[:]` is the canonical "shallow copy" idiom. Use it when you need to
hand a function a list it can mutate without disturbing the original.

## Mutation

```python
nums.append(60)         # adds to end
nums.insert(0, 5)       # insert at index 0
nums.extend([70, 80])   # append every element of another iterable
nums.pop()              # remove + return last element
nums.pop(0)             # remove + return element at index 0
nums.remove(30)         # remove first occurrence of value 30 (ValueError if absent)
nums[1] = 99            # in-place reassignment
del nums[2]             # delete by index
nums.clear()            # remove all elements
```

Two pairs that confuse beginners:

| Operation | Effect | Returns |
|---|---|---|
| `nums.sort()` | mutates in place | `None` |
| `sorted(nums)` | leaves `nums` alone | new sorted list |
| `nums.reverse()` | mutates in place | `None` |
| `reversed(nums)` | leaves `nums` alone | iterator (wrap in `list(...)` to materialise) |

A common bug: `result = nums.sort()` — `result` is `None`, not the sorted
list. Use `sorted(nums)` or call `.sort()` and use `nums` afterwards.

## Membership and search

```python
3 in [1, 2, 3]           # True
4 not in [1, 2, 3]       # True
[1, 2, 3].index(2)       # 1     - first index of 2
[1, 2, 3, 2].count(2)    # 2
```

`in` is O(n) for lists. If you find yourself doing many `in` checks,
that's a signal to switch to a `set` or `dict` — see
{{GUIDE:cs106a-dictionaries}}.

## Iterating

```python
for n in nums:
    print(n)

for i, n in enumerate(nums):
    print(i, n)

for a, b in zip([1, 2, 3], ["a", "b", "c"]):
    print(a, b)
```

`zip` stops at the shorter input. `enumerate` is preferred over
`range(len(nums))` — see {{GUIDE:cs106a-control-flow}}.

## Aliasing — the "two names, one list" trap

Lists are passed and assigned **by reference**:

```python
a = [1, 2, 3]
b = a              # b is the SAME list, not a copy
b.append(4)
print(a)           # [1, 2, 3, 4]   - surprise!
```

If you want a copy:

```python
b = a[:]           # idiomatic shallow copy
b = list(a)        # equivalent
import copy
b = copy.deepcopy(a)   # if list contains lists you also want copied
```

The same trap appears with `=` defaults and with passing lists into
functions. Whenever you hear "but I never touched the original", ask
yourself if there's an alias.

## Lists of lists (2D)

CS 106A's grid problems use nested lists:

```python
grid = [[0] * 3 for _ in range(3)]
grid[1][2] = 9
# [[0, 0, 0], [0, 0, 9], [0, 0, 0]]
```

**Critical**: do *not* write `grid = [[0] * 3] * 3`. That creates one
inner list and three references to it — mutating one row mutates all
three. The list comprehension form creates a fresh inner list per row.

## Sorting with a key

```python
words = ["banana", "fig", "apple"]
sorted(words)                  # ['apple', 'banana', 'fig']  (alphabetical)
sorted(words, key=len)         # ['fig', 'apple', 'banana'] (by length)
sorted(words, reverse=True)    # ['fig', 'banana', 'apple']
```

The `key` argument takes a function applied to each element to produce
the value to sort on. Don't sort then post-process — pass `key` instead.

## Common patterns

**Filter into a new list:**

```python
positives = [n for n in nums if n > 0]
```

That's a list comprehension — see {{GUIDE:cs106a-list-comprehensions}}.

**Find the maximum:**

```python
biggest = max(nums)
biggest_word = max(words, key=len)
```

**Sum or average:**

```python
total = sum(nums)
avg = total / len(nums) if nums else 0   # guard against empty list
```

That `if nums else 0` guard is the clean way to avoid `ZeroDivisionError`.

For boilerplate when starting an assignment that uses lists, see
`{{FILE:stanford-cs106a-pset-setup-readme}}`.

## Practice

Lists don't have a dedicated quiz in this set; their idioms are tested
implicitly across the comprehensions, recursion, and OOP quizzes. Make
sure you can recite — without thinking — the difference between
`nums.sort()` and `sorted(nums)`. Then move on to
{{GUIDE:cs106a-dictionaries}}.
