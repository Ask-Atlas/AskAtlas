---
slug: wsu-cpts121-pointers-cheatsheet
title: "Pointers Cheatsheet (CPTS 121)"
mime: application/pdf
filename: pointers-cheatsheet.pdf
course: wsu/cpts121
description: "Quick reference for C pointers: address-of, dereference, NULL, and common bugs like double-free."
author_role: bot
---

## The Two Operators You Must Not Confuse

- `&x` means **"the address of `x`"** — it gives you a pointer.
- `*p` means **"the thing `p` points to"** — it follows the pointer.

Mnemonic: `&` makes a pointer, `*` uses one.

```c
int x = 42;
int *p = &x;     // p holds the address of x
printf("%d\n", *p);  // prints 42
*p = 99;             // x is now 99
```

## Declaring Pointers

The `*` in a declaration is part of the type, not the name:

```c
int *p, q;    // p is int*, q is int (gotcha!)
int *p, *q;   // both pointers
```

## NULL: The "Points to Nothing" Value

Always initialize pointers. An uninitialized pointer holds a garbage address and dereferencing it is undefined behavior.

```c
int *p = NULL;
if (p != NULL) { *p = 5; }  // safe check
```

Return `NULL` from functions that allocate memory to signal failure
(matches the "pointer-returning function returns NULL on failure" convention
used throughout the C standard library):

```c
int *allocate_buffer(size_t n) {
    int *arr = malloc(n * sizeof(int));
    if (arr == NULL) { return NULL; }
    return arr;
}
```

## Pointers and Arrays

An array name decays to a pointer to its first element:

```c
int a[5] = {1,2,3,4,5};
int *p = a;       // same as &a[0]
printf("%d", p[2]);  // 3
printf("%d", *(p+2)); // 3 — identical
```

But `sizeof(a)` inside the declaring scope gives 20 (bytes), while `sizeof(p)` gives 8 (pointer size). Do not pass `a` to a function and expect `sizeof` to work.

## malloc / free Rules

| Rule | Why |
|---|---|
| Every `malloc` needs exactly one `free` | Leak or crash otherwise |
| Set pointer to `NULL` after `free` | Prevents use-after-free |
| Never `free` the same block twice | Double-free = heap corruption |
| Never `free` a non-malloc pointer | Crash, undefined behavior |

```c
int *buf = malloc(100 * sizeof(int));
if (!buf) return -1;
// ... use buf ...
free(buf);
buf = NULL;   // defensive
```

## Common Bugs

**Dangling pointer** — the memory was freed but the pointer still holds the old address:

```c
int *p = malloc(sizeof(int));
free(p);
*p = 5;       // UNDEFINED BEHAVIOR
```

**Off-by-one with pointer arithmetic** — walking past the end:

```c
for (int *q = a; q <= a + 5; q++)  // BUG: should be < a+5
```

**Forgetting `&` in `scanf`** — classic segfault:

```c
int n;
scanf("%d", n);   // WRONG, crashes
scanf("%d", &n);  // correct
```

## Passing by Pointer (Simulated Pass-by-Reference)

C passes everything by value, so to modify a caller's variable, pass its address:

```c
void zero(int *x) { *x = 0; }

int main(void) {
    int n = 7;
    zero(&n);     // n is now 0
}
```

## Quick Self-Check

1. What does `int **pp` mean? *(pointer to a pointer to int)*
2. If `p == NULL`, is `*p` safe? *(no — segfault)*
3. After `free(p)`, is `p == NULL`? *(no — `free` does not clear it)*
