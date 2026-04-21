---
slug: cpts121-midterm-1-review
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "Midterm 1 Review — CPTS 121"
description: "Syntax, control flow, functions, and pointer basics — the topics CPTS 121's first midterm focuses on, with worked review questions."
tags: ["c", "midterm", "review", "exam-prep", "midterm-1"]
author_role: bot
attached_files:
  - wsu-cpts121-midterm-prep-template
  - wsu-cpts121-pointers-cheatsheet
attached_resources: []
---

# Midterm 1 Review

Midterm 1 in CPTS 121 covers weeks 1–6: basic syntax, types, control
flow, functions, and the first half of the pointers material. This
guide hits the topics the exam reliably tests, with the exact pattern
of question you should expect for each.

The midterm prep template ({{FILE:wsu-cpts121-midterm-prep-template}})
is the official scratch sheet. Print it, fill it in, and you'll have
covered the bulk of what the exam tests.

## Topic 1 — types and conversions

Expect 2-3 questions of the form "what does this print?":

```c
int    a = 5;
int    b = 2;
double c = a / b;
printf("%f\n", c);   // prints 2.000000, NOT 2.500000
```

The trap is **integer division**: `a / b` is computed as `int`,
truncating to `2`, *then* converted to `double`. To get `2.5`, force
one operand to `double`:

```c
double c = (double)a / b;
```

Other classic gotchas the lectures will have hammered:

- `printf("%d", 3.14)` prints garbage. Format-specifier mismatches
  are undefined behavior.
- `'A' + 1` is `66` (an `int`), because character arithmetic is just
  ASCII arithmetic.
- `1 / 0` is undefined behavior in standard C, but most compilers
  produce a runtime trap.

## Topic 2 — control flow

The exam loves loops with a sentinel:

```c
int sum = 0, n;
while (scanf("%d", &n) == 1 && n != -1) {
    sum += n;
}
printf("%d\n", sum);
```

Be ready to:

- Hand-trace a loop for a given input
- Identify off-by-one bugs (`<` vs `<=`)
- Spot infinite loops (forgot to update the loop variable)
- Convert a `for` loop into the equivalent `while` loop

The control-flow guide ({{GUIDE:cpts121-control-flow}}) walks the
canonical shapes.

## Topic 3 — functions and scope

Expect a "what does this print?" with parameter passing:

```c
void modify(int x) {
    x = 100;
}

int main(void) {
    int n = 7;
    modify(n);
    printf("%d\n", n);   // prints 7
}
```

C is pass-by-value. To modify a caller's variable, take a pointer:

```c
void modify(int *x) {
    *x = 100;
}
modify(&n);
```

Also expect questions on:

- The difference between **declaration** and **definition**
- Variable scope (block vs file scope)
- Why `return &local_var;` is a bug
- Why prototypes matter (the `implicit int` rule)

Review {{GUIDE:cpts121-functions-and-scope}} for the fully-worked
examples.

## Topic 4 — pointer basics

Pointers usually appear on midterm 1 in their simplest form. Expect:

```c
int x = 10;
int *p = &x;
*p = 20;
printf("%d\n", x);   // 20
```

You should be able to walk through:

- Reading `int *p` as "p is a pointer to int"
- The difference between `p` (the address) and `*p` (the value at
  that address)
- Why `int *p, q;` declares only `p` as a pointer
- Why `scanf("%d", n)` (no `&`) is wrong

The pointers cheatsheet ({{GUIDE:cpts121-pointers-cheatsheet}})
has the canonical memory diagrams; the attached PDF
({{FILE:wsu-cpts121-pointers-cheatsheet}}) has more.

## Topic 5 — basic arrays

Midterm 1 tests arrays only in their simplest form (no pointer
arithmetic yet). Be ready for:

```c
int arr[5] = {1, 2, 3, 4, 5};
int sum = 0;
for (int i = 0; i < 5; i += 1) {
    sum += arr[i];
}
printf("%d\n", sum);   // 15
```

Common exam questions:

- "What's `sizeof(arr) / sizeof(arr[0])`?" — answer: the number of
  elements, here `5`.
- "What does this print after `arr[5] = 99;`?" — trick question:
  index 5 is out of range; UB.
- "What's the last legal index of an array of length N?" — `N - 1`.

## Sample short-answer practice

These are the kinds of free-response questions that show up:

1. **Explain why `if (x = 5)` is almost always a bug.** Single `=` is
   assignment; the condition becomes "is 5 truthy?" which is always
   true. Wanted: `==`.

2. **What's the value of `7 / 2 + 7 % 2`?** `3 + 1 = 4`.

3. **Write a function `int max(int a, int b)` that returns the larger
   value.**
   ```c
   int max(int a, int b) {
       if (a > b) return a;
       return b;
   }
   ```

4. **Why is `scanf("%d %d", &x, y)` (where `y` is a plain int) wrong?**
   `scanf` writes through pointers. Passing `y` instead of `&y`
   interprets `y`'s value as an address — almost always a segfault.

5. **What's the difference between `'5'` and `5`?** The first is the
   character `'5'`, ASCII value `53`. The second is the integer `5`.

## Day-of advice

- The exam is closed-book. Write out the canonical loop shape
  (`for (int i = 0; i < n; i += 1)`) on the scratch paper as soon as
  you sit down so you don't second-guess the bounds.
- Read every "what does this print?" twice. Underline format
  specifiers and operator precedence.
- For free-response code, the rubric scores correct compilation, then
  correct behavior, then style. Get something compiling first.
- Time yourself on practice problems. Average about 2 minutes per
  multiple-choice and 5 minutes per short-answer.

## Where to next

After midterm 1, the course pivots to dynamic memory, structs, and
file I/O. Start with {{GUIDE:cpts121-arrays-and-strings}} for the
bridge between the two halves of the course, then
{{GUIDE:cpts121-dynamic-memory-allocation}} for the second-half
material.
