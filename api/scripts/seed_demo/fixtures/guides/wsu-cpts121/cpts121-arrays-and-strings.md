---
slug: cpts121-arrays-and-strings
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "Arrays and Strings in C — CPTS 121"
description: "How C arrays decay to pointers, why C strings are NUL-terminated char arrays, and the bugs that combination produces."
tags: ["c", "arrays", "strings", "midterm", "week-4"]
author_role: bot
quiz_slug: cpts121-arrays-and-strings-quiz
attached_files:
  - wsu-cpts121-arrays-and-strings
  - unsplash-abstract-computer-code-1
attached_resources: []
---

# Arrays and Strings in C

If pointers are the most-tested topic on the midterm, arrays and strings
are a close second — partly because they *are* pointers in disguise.
This guide unpacks the relationship and the bugs that follow.

## Declaring arrays

```c
int scores[10];                     // 10 ints, uninitialised garbage
int zeros[10] = {0};                // all 10 set to 0
int firsts[5] = {1, 2, 3};          // first three set; rest are 0
double weights[] = {0.5, 0.3, 0.2}; // length 3, inferred from initialiser
```

Two rules from week 4 lecture:

- **The size must be known at declaration time** for arrays at file
  scope or `static` arrays. C99 added VLAs (variable-length arrays) for
  block scope, but the course style guide forbids them — the size of
  every lab array must be a compile-time constant.
- **Always initialise.** `int scores[10];` contains garbage values until
  you write to each slot. `int scores[10] = {0};` is two extra characters
  and saves an entire category of bugs.

## Indexing

C arrays are **zero-indexed**. An array of length `n` has valid indices
`0` through `n - 1`. Indexing out of range is undefined behavior — there
is no bounds check. The most common shape:

```c
const int n = 10;
int sum = 0;
for (int i = 0; i < n; i += 1) {
    sum += arr[i];
}
```

Off-by-one errors here are responsible for half of all "I don't know
why my program crashes" office-hours visits. See
{{GUIDE:cpts121-control-flow}} for the canonical loop shape.

## Arrays decay to pointers

This is the single most important fact about C arrays. When you pass an
array to a function, what the function actually receives is a pointer
to the first element. The length is **not** carried along.

```c
void print_first(int arr[]) {     // really `int *arr`
    printf("%d\n", arr[0]);
    printf("%zu\n", sizeof(arr)); // prints sizeof(int *), NOT array length
}
```

The implication: every function that takes an array also needs the
length passed explicitly:

```c
int sum(const int *arr, size_t n) {
    int total = 0;
    for (size_t i = 0; i < n; i += 1) {
        total += arr[i];
    }
    return total;
}
```

Both `int arr[]` and `int *arr` are equivalent in a parameter list. The
second form is more honest about what's actually happening.

## C strings: char arrays with a sentinel

C does not have a real string type. A "string" is a `char` array whose
last meaningful byte is a NUL (`'\0'`) terminator. Every `<string.h>`
function reads until it hits one of those NULs.

```c
char greeting[] = "hi";   // 3 bytes: 'h', 'i', '\0'
char buffer[16] = {0};    // 16 zeros — empty string + room to grow
```

Counting the NUL is the difference between writing a working program
and corrupting the heap. If you allocate space for "an 8-character
name," you actually need a buffer of size `9`.

## String functions you'll use

| Function | What it does | Watch out for |
|---|---|---|
| `strlen(s)` | Length not counting the NUL | UB if `s` is not NUL-terminated |
| `strcpy(d, s)` | Copy until NUL | No bound on `d` — overflow risk |
| `strncpy(d, s, n)` | Copy at most `n` bytes | May NOT NUL-terminate `d` |
| `strcmp(a, b)` | 0 if equal, neg/pos otherwise | Returns int, **not** boolean |
| `snprintf(d, n, fmt, ...)` | Safe formatted copy | Always NUL-terminates if `n > 0` |

The course strongly prefers `snprintf` over `strcpy` / `sprintf` for
exactly this reason — it takes a destination size.

## The `strcpy` overflow

The classic bug:

```c
char dest[8];
strcpy(dest, "this string is much longer than 8 bytes");  // UB
```

`strcpy` walks `src` byte-by-byte writing into `dest` until it hits a
NUL. It has *no idea* how big `dest` is, so it happily writes past the
end. On the lab machines that's typically a stack-smashing canary trip
and a `*** stack smashing detected ***` at exit. Use `snprintf`:

```c
char dest[8];
snprintf(dest, sizeof(dest), "%s", "long string");
// dest now holds "long st" + '\0' — truncated but safe
```

## Iterating over a string

Two idiomatic forms:

```c
// 1. Index until NUL
for (int i = 0; s[i] != '\0'; i += 1) {
    putchar(s[i]);
}

// 2. Pointer until NUL
for (const char *p = s; *p != '\0'; p += 1) {
    putchar(*p);
}
```

Both are O(length). Don't call `strlen(s)` inside a loop test —
that turns the loop into O(n²).

## Multi-dimensional arrays

```c
int board[3][3] = {
    {1, 2, 3},
    {4, 5, 6},
    {7, 8, 9}
};

printf("%d\n", board[1][2]);   // 6
```

Stored row-major in memory. Pass them to functions with the inner
dimension fixed:

```c
void print_3x3(int b[][3]) { ... }    // inner dim required
```

CPTS 121 typically only goes as deep as 2D for the matrix lab.

## Common gotchas

- **`sizeof(arr)` inside a function** gives `sizeof(int *)`, not the
  original array's byte size. The decay strikes again.
- **Forgetting the NUL** when building a string by hand. After
  `buf[i] = c; i += 1;`, you must do `buf[i] = '\0';` before any
  `<string.h>` call reads `buf`.
- **`char *s = "literal"; s[0] = 'X';`** — string literals are stored in
  read-only memory. Modifying them is undefined behavior. Use
  `char s[] = "literal";` if you need to modify it.
- **Using `==` to compare strings.** `==` on `char *` compares pointers,
  not contents. Use `strcmp(a, b) == 0` for equality.

## Worked example: count digits

```c
#include <ctype.h>
#include <string.h>

int count_digits(const char *s) {
    int total = 0;
    for (size_t i = 0; s[i] != '\0'; i += 1) {
        if (isdigit((unsigned char)s[i])) {
            total += 1;
        }
    }
    return total;
}
```

Note the `(unsigned char)` cast — the `<ctype.h>` macros are undefined
for negative values, and `char` may be signed on some systems. The
course style guide requires the cast on every `isdigit`/`isalpha`/etc.
call.

## Practice

Once the relationship between arrays and pointers makes sense, take
{{QUIZ:cpts121-arrays-and-strings-quiz}}. The cheatsheet PDF
({{FILE:wsu-cpts121-arrays-and-strings}}) has memory-layout diagrams
that pair well with the questions.

## Where to next

For the underlying memory model, see
{{GUIDE:cpts121-pointers-cheatsheet}}. To learn how to *grow* an array
at runtime instead of fixing its size at compile time, jump to
{{GUIDE:cpts121-dynamic-memory-allocation}}.
