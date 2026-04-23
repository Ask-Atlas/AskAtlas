---
slug: cpts121-dynamic-memory-allocation
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "Dynamic Memory Allocation in C — CPTS 121"
description: "malloc, calloc, realloc, and free — what they do, what they don't do, and how to keep the heap clean."
tags: ["c", "memory", "malloc", "free", "heap", "midterm-2"]
author_role: bot
quiz_slug: cpts121-dynamic-memory-allocation-quiz
attached_files:
  - wsu-cpts121-pointers-cheatsheet
  - wikimedia-modern-c-gustedt-pdf
attached_resources: []
---

# Dynamic Memory Allocation in C

Static arrays only get you so far. Once a lab needs "however many
records the user wants to enter," you need the heap. The C standard
library exposes four functions for managing it. This guide walks each
one and the rules that keep them from blowing up.

## The four functions

```c
#include <stdlib.h>

void *malloc(size_t size);
void *calloc(size_t count, size_t size);
void *realloc(void *ptr, size_t new_size);
void  free(void *ptr);
```

All allocators return a `void *` (untyped pointer) on success and
`NULL` on failure. Always check.

## `malloc` — the workhorse

```c
int *p = malloc(sizeof(int));
if (p == NULL) {
    fprintf(stderr, "out of memory\n");
    return 1;
}
*p = 42;
free(p);
p = NULL;
```

Two patterns the lectures will repeat:

1. **Use `sizeof(*p)`, not `sizeof(int)`.** If you change `p`'s type
   later, the size stays correct automatically:

   ```c
   int *p = malloc(sizeof(*p));     // good
   int *q = malloc(sizeof(int));    // brittle
   ```

2. **Don't cast the return value.** In C (unlike C++), the implicit
   `void *` to `int *` conversion is fine. A cast can mask the missing
   `<stdlib.h>` include.

## `calloc` — for arrays, zero-initialised

```c
int *arr = calloc(n, sizeof(*arr));
```

Two parameters: count and element size. Returned memory is **zeroed
out**, unlike `malloc`. Use this when you need a block of N items that
should start at zero (most of the time).

`calloc` also checks for multiplication overflow internally — `calloc(n, sizeof(*arr))`
is safer than `malloc(n * sizeof(*arr))` if `n` could be huge.

## `realloc` — grow or shrink

```c
int *bigger = realloc(arr, new_n * sizeof(*arr));
if (bigger == NULL) {
    free(arr);   // arr is still valid; clean it up
    return 1;
}
arr = bigger;
```

Three subtleties:

- `realloc` may return a **different** address. The old pointer is
  invalid after a successful realloc.
- If `realloc` fails, it returns `NULL` and **leaves the original
  pointer valid**. The pattern above (`bigger = realloc(...); if (bigger
  != NULL) arr = bigger;`) avoids the textbook leak `arr = realloc(arr,
  ...)`, which loses `arr` on failure.
- `realloc(NULL, n)` is equivalent to `malloc(n)`. Useful for "first
  allocation looks the same as a grow."

## `free` — exactly once

```c
free(p);
p = NULL;     // defensive: a second free becomes a no-op
```

Three rules:

1. **Free exactly once.** Double-free is undefined behavior — the heap
   allocator may corrupt itself silently.
2. **Free only what you allocated.** `free`-ing a stack pointer or an
   address that was never returned by `malloc` is also UB.
3. **`free(NULL)` is fine.** It's a guaranteed no-op. That's why setting
   `p = NULL` after `free(p)` makes a second `free` safe.

For the deeper memory-model picture, see
{{GUIDE:cpts121-pointers-cheatsheet}} and the cheatsheet PDF in this
guide's attachments ({{FILE:wsu-cpts121-pointers-cheatsheet}}).

## A growable array

The realistic shape of "I don't know how many records I'll need":

```c
typedef struct {
    int    *data;
    size_t  len;
    size_t  cap;
} int_vec_t;

int int_vec_push(int_vec_t *v, int value) {
    if (v->len == v->cap) {
        size_t new_cap = (v->cap == 0) ? 8 : v->cap * 2;
        int *bigger = realloc(v->data, new_cap * sizeof(*v->data));
        if (bigger == NULL) return -1;
        v->data = bigger;
        v->cap  = new_cap;
    }
    v->data[v->len] = value;
    v->len += 1;
    return 0;
}

void int_vec_free(int_vec_t *v) {
    free(v->data);
    v->data = NULL;
    v->len = v->cap = 0;
}
```

Doubling the capacity makes `push` amortised O(1). Halving on shrink
is wasteful and rarely worth doing in CPTS 121 labs.

## Strdup, by hand

There is no `strdup` in C99. CPTS 121 expects you to write one:

```c
char *my_strdup(const char *s) {
    size_t n = strlen(s) + 1;       // +1 for the NUL
    char  *copy = malloc(n);
    if (copy == NULL) return NULL;
    memcpy(copy, s, n);
    return copy;
}
```

The `+1` for the NUL terminator is the bug everyone hits the first
time. See {{GUIDE:cpts121-arrays-and-strings}} for why.

## Common bugs

- **Forgetting the size of one element.** `malloc(n)` allocates `n`
  bytes, not `n` ints. Always multiply by `sizeof(*p)`.
- **Returning the address of a local then trying to `free` it.** Stack
  memory is not heap memory; you can't `free` it.
- **Memory leaks.** Every successful `malloc`/`calloc`/`realloc` must
  eventually have a matching `free`. Run `valgrind` (lab machines
  preinstalled) on every lab — it will tell you exactly where you
  leaked.
- **Use-after-free.** Reading `*p` after `free(p)` is UB. The memory
  may have been recycled into something else by then.
- **`realloc` losing the original.** `arr = realloc(arr, n)` leaks `arr`
  on failure. Use a temporary variable.

## Valgrind in 90 seconds

```sh
valgrind --leak-check=full --show-leak-kinds=all ./lab5
```

A clean run shows `All heap blocks were freed -- no leaks are possible`.
The CPTS 121 grading rubric checks for this on every project — leaks
cost points.

## Worked example: read N integers from the user

```c
int main(void) {
    size_t n = 0;
    if (scanf("%zu", &n) != 1) return 1;

    int *arr = calloc(n, sizeof(*arr));
    if (arr == NULL) return 1;

    for (size_t i = 0; i < n; i += 1) {
        if (scanf("%d", &arr[i]) != 1) {
            free(arr);
            return 1;
        }
    }

    int sum = 0;
    for (size_t i = 0; i < n; i += 1) sum += arr[i];
    printf("%d\n", sum);

    free(arr);
    return 0;
}
```

Note the `free(arr)` on **every** exit path, including the error case
inside the loop. Single-exit functions make this easier; otherwise be
disciplined about it.

## Practice

Once you can write the growable-array pattern from memory, take
{{QUIZ:cpts121-dynamic-memory-allocation-quiz}}.

## Where to next

The natural next topic is {{GUIDE:cpts121-file-io}} — allocating
buffers to read records from disk. The
{{GUIDE:cpts121-debugging-with-gdb}} guide pairs well too: most
heap bugs are caught fastest by stepping through the allocator calls
in the debugger.
