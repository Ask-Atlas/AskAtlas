---
slug: cpts121-pointers-cheatsheet
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "Pointers, Arrays, and Memory in C — CPTS 121 Cheatsheet"
description: "Common pointer patterns from CPTS 121 lectures + lab notes, with worked memory diagrams."
tags: ["c", "pointers", "memory", "midterm"]
author_role: bot
quiz_slug: cpts121-pointers-quiz
attached_files:
  - wsu-cpts121-pointers-cheatsheet
  - wikimedia-modern-c-gustedt-pdf
---

# Pointers, Arrays, and Memory in C

Pointers are the single most-tested topic on the CPTS 121 midterm. Get them
right and the rest of the course is mostly bookkeeping. Get them wrong and
you spend the next three labs hunting `Segmentation fault (core dumped)`.

## Quick reference

| Symbol | Meaning |
|---|---|
| `&x` | address-of `x` |
| `*p` | dereference `p` (read or write through it) |
| `int *p` | declares `p` as a pointer to `int` |
| `int *p = NULL;` | initialised to "no address" — safe to test against `NULL` |

## The three rules

1. **Always initialise.** An uninitialised pointer holds garbage; dereferencing
   it is undefined behavior. Set it to `NULL` if you don't have a real
   address yet.
2. **Never dereference `NULL`.** Test before you read. Most "random crashes"
   in CPTS 121 trace back to this.
3. **`free()` exactly once.** Calling `free(p)` twice is a double-free, and
   the heap allocator may corrupt itself in subtle ways. Set `p = NULL`
   immediately after `free(p)` so a second `free()` is a no-op.

## Allocation pattern

```c
int *p = malloc(sizeof(int));
if (p == NULL) {
    fprintf(stderr, "out of memory\n");
    exit(1);
}
*p = 42;
printf("%d\n", *p);
free(p);
p = NULL;  // defensive
```

## Arrays decay to pointers

When you pass an array to a function, the function sees a pointer to its
first element — the array's length is **not** carried along.

```c
void print_first(int arr[]) {  // really `int *arr`
    printf("%d\n", arr[0]);
}
```

This is why CPTS 121 functions that accept arrays almost always take a
length parameter too. The diagrams in `{{FILE:wsu-cpts121-pointers-cheatsheet}}`
walk through what's actually on the stack.

## When to use what

- **Local variable on the stack** — declared inside a function, freed
  automatically when the function returns. Don't return its address.
- **`malloc`'d memory on the heap** — survives until you `free` it.
  You're responsible for the cleanup.
- **String literal** — read-only memory. Modifying `*"hello"` is undefined.

For the canonical reference on all of this, the relevant chapters of
Gustedt's *Modern C* are linked in this guide's attachments.

## Common gotchas

- `sizeof(arr)` inside a function gives `sizeof(int *)`, not the array
  length — because `arr` is a pointer there.
- `int *p, q;` declares `p` as a pointer and `q` as a plain `int`. The
  `*` binds to the variable, not the type.
- `scanf("%d", n)` segfaults if `n` is an `int` (you forgot the `&`).
  `scanf` always wants pointers.

## Practice

Once you can recite the three rules from memory and walk through the
allocation pattern without looking, take {{QUIZ:cpts121-pointers-quiz}}.
If you miss any of the questions on `free()` or NULL-checking, re-read
the "three rules" section before moving on.

For the next topic in the sequence, jump back to the {{COURSE:wsu/cpts121}}
catalog page.
