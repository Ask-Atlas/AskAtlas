---
slug: cpts121-functions-and-scope
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "Functions and Scope in C — CPTS 121"
description: "Prototypes, parameter passing, return values, and the scope rules that explain why your variable is suddenly zero."
tags: ["c", "functions", "scope", "prototypes", "week-3"]
author_role: bot
attached_files:
  - wsu-cpts121-lab-readme
  - wikimedia-modern-c-gustedt-pdf
attached_resources: []
---

# Functions and Scope in C

Functions are how you turn a 200-line `main` into something you can
actually reason about. The CPTS 121 grading rubric explicitly penalises
oversized functions — the 50-line cap exists for a reason. This guide
covers the mechanics and the scope rules.

## Anatomy of a function

```c
double average(int total, int count) {
    if (count == 0) {
        return 0.0;
    }
    return (double)total / count;
}
```

Three pieces:

- **Return type** (`double`) — what the function evaluates to. Use `void`
  if the function returns nothing.
- **Name and parameter list** (`average(int total, int count)`) — the
  arguments are local copies inside the function (more on this below).
- **Body** in braces. The function exits at the first `return` reached,
  or by falling off the end (only allowed for `void` functions).

## Prototypes

C is a *single-pass* language — when the compiler sees a call to
`average(...)` it must already know the function's signature. Two
patterns satisfy this:

**1. Define the function above where it's called:**

```c
double average(int total, int count) { ... }

int main(void) {
    printf("%f\n", average(50, 4));
    return 0;
}
```

**2. Provide a prototype at the top, define later:**

```c
double average(int total, int count);   // prototype, ends in ;

int main(void) { ... }

double average(int total, int count) { ... }
```

Lab style guide: when there are more than two helper functions, prefer
prototypes at the top. It puts the high-level structure (`main`) first
and pushes the details below.

## Pass by value

C **always** passes parameters by value — the function gets its own copy.

```c
void zero_out(int x) {
    x = 0;
}

int main(void) {
    int n = 7;
    zero_out(n);
    printf("%d\n", n);   // prints 7, not 0
}
```

If you want a function to modify a caller's variable, pass a pointer to
it (covered in {{GUIDE:cpts121-pointers-cheatsheet}}). This is why
`scanf("%d", &n)` needs the `&` — `scanf` writes through the pointer.

The same rule applies to arrays *almost* — arrays decay to pointers when
passed, so the function actually receives a pointer to the first element.
That decay means modifications inside the function *do* affect the
caller's array. See {{GUIDE:cpts121-arrays-and-strings}} for the gory
details.

## Return values

A function returns at most one value. If you need more, either:

- Wrap them in a struct (see {{GUIDE:cpts121-structs-and-typedefs}}), or
- Pass output parameters as pointers, e.g.
  `void min_max(int *arr, int n, int *min_out, int *max_out)`.

Returning a pointer to a local variable is a serious bug:

```c
int *bad(void) {
    int x = 42;
    return &x;     // x is destroyed when bad() returns
}
```

The pointer points at memory the runtime is free to reuse. The lab
grader has a test for this and it will fail you.

## Scope rules

Three scopes you'll encounter in CPTS 121:

### Block scope

Variables declared inside `{ ... }` exist only until the closing brace.

```c
if (x > 0) {
    int squared = x * x;
    printf("%d\n", squared);
}
// squared is not visible here
```

Loop variables declared in the `for` header have block scope too:

```c
for (int i = 0; i < n; i += 1) { ... }
// i is not visible here in C99/C11
```

### Function scope

Parameters and variables declared at the top of the function body are
visible throughout the function (subject to inner-block shadowing).

### File scope

Variables declared *outside* any function have file scope and live for
the whole program.

```c
static int call_count = 0;   // file scope, this .c only

void track_call(void) {
    call_count += 1;
}
```

The `static` keyword at file scope means "do not export this name to
other translation units." Use it on every file-scope variable and helper
function unless you specifically need it to be visible elsewhere — this
is a project-wide CPTS 121 convention.

## `static` inside a function

Different meaning: a `static` local variable persists across calls.

```c
int next_id(void) {
    static int counter = 0;
    counter += 1;
    return counter;
}
```

The first call returns 1, the next returns 2, and so on. The variable
is initialised once, the first time the function is reached.

This is occasionally useful, but lectures will warn you that
function-local `static` is a hidden global — easy to abuse, hard to test.

## The `const` modifier on parameters

Mark input pointers `const` to document that the function won't modify
the data:

```c
size_t string_length(const char *s) {
    size_t n = 0;
    while (s[n] != '\0') n += 1;
    return n;
}
```

The compiler will reject `s[n] = ...` inside this function, which is
exactly what you want. Adding `const` is also a hint to the reader
about which parameters are "in" and which are "in/out."

## Common bugs

- **Missing prototype.** Calling an undeclared function gets you an
  implicit `int` return type and silent disaster. Compile with `-Wall`
  and you'll see `implicit declaration of function ...`.
- **Forgetting `return`.** A non-void function that falls off the end is
  undefined behavior. The compiler usually warns; do not ignore it.
- **Shadowing.** Reusing a name in an inner scope is legal and almost
  always a bug. Style points come off for it.

## Where to next

The next conceptual jump is composing functions over arrays — see
{{GUIDE:cpts121-arrays-and-strings}}. After that, lecture moves into
pointers proper ({{GUIDE:cpts121-pointers-cheatsheet}}), at which point
you'll re-read this guide and the "pass by value" section will make
much more sense.

The course's lab README ({{FILE:wsu-cpts121-lab-readme}}) has the
required header block and naming conventions for helper-function files.
For the canonical reference, Gustedt's *Modern C* chapter on functions
is included in this guide's attachments.
