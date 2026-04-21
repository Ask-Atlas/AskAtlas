---
slug: wsu-cpts121-arrays-and-strings
title: "Arrays and C-Strings (CPTS 121)"
mime: application/pdf
filename: arrays-and-strings.pdf
course: wsu/cpts121
description: "Arrays, C-strings, indexing rules, bounds bugs, and safe input handling in intro C."
author_role: bot
---

## Arrays: Fixed-Size, Zero-Indexed, Contiguous

```c
int scores[5];           // uninitialized — garbage values
int primes[] = {2,3,5,7}; // size inferred: 4
double grid[3][4] = {0}; // all zeros
```

Three rules that trip up beginners:

1. Valid indices are `0` through `n-1`. Never write to `scores[5]` in a size-5 array.
2. C **does not** check array bounds at runtime. Out-of-bounds access compiles and "runs" — it just corrupts memory.
3. Array size must be known when declared (VLAs exist but avoid them in intro coursework).

## Iterating an Array

```c
int n = sizeof(scores) / sizeof(scores[0]);   // 5
for (int i = 0; i < n; i++) {
    printf("%d ", scores[i]);
}
```

The `sizeof` trick only works in the **scope where the array was declared**. If you pass the array to a function, you must also pass its length.

## Arrays as Function Parameters

```c
// These three are equivalent declarations:
void print(int a[5], int n);
void print(int a[],  int n);
void print(int *a,   int n);
```

Inside `print`, `sizeof(a)` gives the pointer size, not the array size. Always pass length as a separate argument.

## C-Strings: Char Arrays with a `\0` Terminator

A C-string is a `char` array whose last meaningful byte is `'\0'` (the null terminator). Without it, string functions walk off the end.

```c
char a[] = "hi";   // {'h','i','\0'} — size 3
char b[3] = {'h','i','\0'};  // same thing
char c[2] = "hi";  // BUG: no room for '\0'
```

## Essential `<string.h>` Functions

| Function | What it does | Gotcha |
|---|---|---|
| `strlen(s)` | Length, not counting `\0` | Undefined if not terminated |
| `strcpy(dst, src)` | Copy including `\0` | No bounds check — buffer overflow risk |
| `strncpy(dst, src, n)` | Copy up to n bytes | May not null-terminate |
| `strcmp(a, b)` | 0 if equal, neg/pos otherwise | Do not use `==` on strings |
| `strcat(dst, src)` | Append | `dst` must have room |

Prefer `snprintf` for building strings safely:

```c
char buf[64];
snprintf(buf, sizeof(buf), "name=%s, age=%d", name, age);
```

## Reading Input Without Blowing Up

Never use `gets` — it is removed from modern C. Use `fgets` with an explicit size:

```c
char line[100];
if (fgets(line, sizeof(line), stdin) != NULL) {
    line[strcspn(line, "\n")] = '\0';  // strip trailing newline
}
```

With `scanf("%s", buf)`, there is no bounds check. At minimum use a width:

```c
char word[32];
scanf("%31s", word);   // leaves room for '\0'
```

## Classic Bounds Bugs

**Off-by-one loop:**
```c
for (int i = 0; i <= n; i++) a[i] = 0;  // writes a[n] — out of bounds
```

**Forgetting the null terminator budget:**
```c
char name[4];
strcpy(name, "Bob!"); // 5 bytes including '\0' — overflow
```

**Comparing strings with `==`:**
```c
if (word == "yes")     // compares pointers, almost always false
if (strcmp(word,"yes") == 0)  // correct
```

## Checklist

- [ ] Is every index in `[0, n-1]`?
- [ ] Does every C-string have room for `\0`?
- [ ] Are you passing array lengths to functions?
- [ ] Did you bound every input read?
