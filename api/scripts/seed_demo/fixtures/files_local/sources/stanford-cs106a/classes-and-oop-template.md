---
slug: stanford-cs106a-classes-oop-template
title: "Classes and OOP Template"
mime: application/vnd.openxmlformats-officedocument.wordprocessingml.document
filename: classes-oop-template.docx
course: stanford/cs106a
description: "Starter template for Python classes: __init__, instance methods, encapsulation, and dunder methods."
author_role: bot
---

# Classes and OOP Template

A class bundles **state** (attributes) with the **behavior** (methods) that operates on that state. Use this template whenever you start a new class on an assignment.

## Minimal Template

```python
class BankAccount:
    """Represents a single customer bank account."""

    def __init__(self, owner: str, balance: float = 0.0) -> None:
        self.owner = owner
        self._balance = balance      # leading underscore => "treat as private"

    def deposit(self, amount: float) -> None:
        if amount <= 0:
            raise ValueError("Deposit must be positive")
        self._balance += amount

    def withdraw(self, amount: float) -> None:
        if amount > self._balance:
            raise ValueError("Insufficient funds")
        self._balance -= amount

    def balance(self) -> float:
        return self._balance

    def __repr__(self) -> str:
        return f"BankAccount(owner={self.owner!r}, balance={self._balance:.2f})"
```

## The Four Moves

1. **`__init__`** runs once, when the object is created. Set every attribute here.
2. **Instance methods** take `self` as the first argument and act on one object.
3. **Encapsulation** — prefix private fields with `_`. Expose behavior, not raw data.
4. **Dunder methods** (`__repr__`, `__eq__`, `__len__`) plug your object into Python's syntax.

## Encapsulation in Practice

```python
acct = BankAccount("Ada", 100)
acct.deposit(50)         # good: goes through validated method
acct._balance = 999_999  # bad: bypasses validation
```

Python won't stop you, but the underscore is a social contract. Reviewers will flag direct access.

## Comparing Styles

| Style            | When to use                                  | Example                                    |
|------------------|----------------------------------------------|--------------------------------------------|
| Plain class      | State + behavior, custom methods             | `BankAccount` above                        |
| `@dataclass`     | Dumb bag of fields, auto `__init__`/`__eq__` | `@dataclass\nclass Point: x: int; y: int`  |
| Named tuple      | Immutable record                             | `Point = namedtuple("Point", "x y")`       |
| Dict             | Ad-hoc, short-lived                          | `{"x": 1, "y": 2}`                         |

## Dunder Methods You'll Reach For

```python
class Vec2:
    def __init__(self, x: float, y: float) -> None:
        self.x, self.y = x, y

    def __add__(self, other: "Vec2") -> "Vec2":
        return Vec2(self.x + other.x, self.y + other.y)

    def __eq__(self, other: object) -> bool:
        return isinstance(other, Vec2) and (self.x, self.y) == (other.x, other.y)

    def __repr__(self) -> str:
        return f"Vec2({self.x}, {self.y})"
```

Now `Vec2(1, 2) + Vec2(3, 4) == Vec2(4, 6)` just works.

## Review Checklist

- [ ] Every attribute is created in `__init__`.
- [ ] Methods either *query* or *mutate* — not both silently.
- [ ] Invalid state is rejected with a clear exception.
- [ ] `__repr__` returns something you'd be happy to see in a debugger.
- [ ] Tests construct the object, call a method, and assert the observable behavior.

Copy this file into your assignment directory, rename the class, and fill in the blanks.
