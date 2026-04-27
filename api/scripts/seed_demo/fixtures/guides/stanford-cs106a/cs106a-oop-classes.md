---
slug: cs106a-oop-classes
course:
  ipeds_id: "243744"
  department: "CS"
  number: "106A"
title: "Classes and Object-Oriented Programming — CS 106A"
description: "Defining classes, instance state, methods, and the dunder protocol Python uses to make objects feel native."
tags: ["python", "oop", "classes", "objects", "final", "demo-hero"]
author_role: bot
recommended: true
quiz_slug: cs106a-oop-classes-quiz
attached_files:
  - stanford-cs106a-classes-oop-template
  - stanford-cs106a-debugging-techniques
  - stanford-cs106a-list-comprehensions-cheatsheet
  - stanford-cs106a-recursion-primer
attached_resources: []
---

# Classes and OOP

Halfway through CS 106A you stop writing top-level scripts and start
modeling the problem with **objects**. An object is a bundle of state
(attributes) plus the operations that act on that state (methods). A
class is the recipe for making instances of that bundle.

The starter template at
`{{FILE:stanford-cs106a-classes-oop-template}}` is the boilerplate the
graders expect.

## A minimal class

```python
class Point:
    def __init__(self, x: float, y: float) -> None:
        self.x = x
        self.y = y

    def distance_to(self, other: "Point") -> float:
        dx = self.x - other.x
        dy = self.y - other.y
        return (dx * dx + dy * dy) ** 0.5

p = Point(3.0, 4.0)
q = Point(0.0, 0.0)
print(p.distance_to(q))   # 5.0
```

The pieces:

- `class Point:` — defines a new type named `Point`.
- `def __init__(self, ...)` — the **constructor**. Called automatically
  when you write `Point(3.0, 4.0)`.
- `self` — the conventional name for "the instance being acted on". Not a
  keyword, but **always** call it `self`.
- `self.x = x` — creates an instance attribute. Each `Point` gets its
  own `x` and `y`.
- `def distance_to(self, other)` — a method. Called as `p.distance_to(q)`,
  which Python rewrites as `Point.distance_to(p, q)`.

## Why `self`?

When you call `p.distance_to(q)`, Python:

1. Looks up `distance_to` on `Point` (the class).
2. Calls it with `p` as the first argument.

So `self` inside the method *is* the instance. Forgetting `self` is the
single most common Day-1 OOP error in CS 106A:

```python
class Counter:
    def __init__(self):
        count = 0       # BUG: local variable, not attribute
    def tick(self):
        count += 1      # NameError; never created
```

Fix: `self.count = 0` and `self.count += 1`.

## Instance vs class attributes

```python
class Dog:
    species = "Canis familiaris"   # class attribute (shared)

    def __init__(self, name: str):
        self.name = name           # instance attribute (per-dog)
```

Reading `d.species` first looks on `d`, then on `Dog`. Writing
`d.species = "X"` creates a new instance attribute that **shadows** the
class attribute for that instance. This trips people up; assign to class
attributes through the class itself: `Dog.species = "X"`.

Use class attributes for constants and shared defaults; use instance
attributes for per-object state.

## Dunder methods

"Dunder" = "double underscore". These are how Python hooks into your class
to make it feel like a built-in type. The most common ones in CS 106A:

```python
class Fraction:
    def __init__(self, num: int, den: int):
        self.num = num
        self.den = den

    def __repr__(self) -> str:
        return f"Fraction({self.num}, {self.den})"

    def __str__(self) -> str:
        return f"{self.num}/{self.den}"

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, Fraction):
            return NotImplemented
        return self.num * other.den == other.num * self.den

    def __add__(self, other: "Fraction") -> "Fraction":
        return Fraction(
            self.num * other.den + other.num * self.den,
            self.den * other.den,
        )
```

Now `Fraction(1, 2) + Fraction(1, 3)` works, `print(f)` shows `1/2`, and
`Fraction(1, 2) == Fraction(2, 4)` is `True`.

| Dunder | Triggered by |
|---|---|
| `__init__` | `Cls(...)` |
| `__repr__` | `repr(x)`, REPL display |
| `__str__` | `str(x)`, `print(x)`, f-strings |
| `__eq__` | `x == y` |
| `__lt__`, `__le__`, ... | `x < y`, `x <= y`, ... |
| `__add__`, `__sub__`, `__mul__` | `+`, `-`, `*` |
| `__len__` | `len(x)` |
| `__getitem__` | `x[i]` |
| `__iter__` | `for el in x` |
| `__contains__` | `el in x` |

CS 106A typically asks for `__init__`, `__repr__`, `__str__`, and
`__eq__`. The rest are "nice to know".

## Inheritance

A subclass extends or overrides a parent class:

```python
class Animal:
    def __init__(self, name: str) -> None:
        self.name = name

    def speak(self) -> str:
        return "..."

class Dog(Animal):
    def speak(self) -> str:
        return "Woof"

class Puppy(Dog):
    def __init__(self, name: str, weeks_old: int) -> None:
        super().__init__(name)         # call parent constructor
        self.weeks_old = weeks_old
```

`super()` returns a proxy that calls methods on the parent class. Use it
inside `__init__` to ensure parent state is set up. CS 106A keeps
inheritance simple — single-parent and one or two levels deep.

## Composition vs inheritance

When in doubt, **compose** rather than **inherit**. A `Car` is not a
kind of `Engine`; it *has* an engine. Inheritance is for "is-a"
relationships only.

```python
class Engine:
    def start(self) -> None:
        pass

class Car:
    def __init__(self) -> None:
        self.engine = Engine()      # composition

    def go(self) -> None:
        self.engine.start()
```

## Encapsulation conventions

Python has no `private` keyword. Instead it uses naming conventions:

| Prefix | Meaning |
|---|---|
| `name` | public — fair game |
| `_name` | "internal" — touch at your own risk |
| `__name` | name-mangled — rewritten to `_ClassName__name` |

CS 106A grading expects single-underscore for "internal" attributes. Don't
use the double-underscore form unless you have a specific reason — it's
mostly there to avoid collisions in deep inheritance hierarchies.

## Common pitfalls

- **Forgetting `self`** in method definitions or accesses.
- **Mutable class attribute** (`items = []`) acting as a shared list
  across all instances. Use `__init__` to create per-instance lists.
- **`__repr__` returning non-string** raises `TypeError` deep inside
  the REPL — usually because you returned a tuple by accident.
- **Comparing instances with `==`** without defining `__eq__` falls back
  to identity (`is`), which is almost never what you want.

## Practice

Once you've defined a `Fraction` end-to-end without looking at the
example, take {{QUIZ:cs106a-oop-classes-quiz}}. The most-missed question
is the `class attribute vs instance attribute` one — re-read that section
if you flub it.

For background on the function-call mechanics that underpin methods, see
{{GUIDE:cs106a-functions-and-scope}}.
