---
slug: cpts121-control-flow
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "Control Flow in C — CPTS 121"
description: "if / else / switch and the three loop forms, plus the gotchas the autograder loves to find."
tags: ["c", "control-flow", "loops", "switch", "week-2"]
author_role: bot
attached_files:
  - wsu-cpts121-lab-readme
  - unsplash-png-code-abstract-1
attached_resources: []
---

# Control Flow in C

By the end of week 2 the lectures will have covered every flow-control
construct in the language. There are not many. The hard part is using
them well — short bodies, clear conditions, and no off-by-one errors.

## `if` / `else if` / `else`

```c
if (score >= 90) {
    grade = 'A';
} else if (score >= 80) {
    grade = 'B';
} else if (score >= 70) {
    grade = 'C';
} else {
    grade = 'F';
}
```

Course style requires braces on every branch, even one-liners. The
order of the `else if` rungs matters — once a condition matches, the
rest are skipped.

A common bug from week 2 labs:

```c
if (0 < x < 10) { ... }   // does NOT mean what you think
```

C parses this as `(0 < x) < 10`, which is `(0 or 1) < 10`, which is
*always true*. The right way:

```c
if (x > 0 && x < 10) { ... }
```

## `switch`

```c
switch (menu_choice) {
    case 1:
        printf("Add\n");
        break;
    case 2:
        printf("Remove\n");
        break;
    case 3:
    case 4:
        printf("Edit or rename\n");
        break;
    default:
        printf("Unknown option\n");
        break;
}
```

Two things the lectures will hammer:

1. **Always `break`** unless you specifically want fall-through. Forgetting
   `break` is a textbook bug — control "falls through" to the next case.
2. **Always include `default`**. Even if the cases are exhaustive today,
   they may not be tomorrow. The autograder tests pass values you didn't
   anticipate.

`switch` only works on integer-compatible types (`int`, `char`, `enum`).
You can't switch on a `double` or a string — use `if`/`else if` for those.

## `while`

```c
int n = 10;
while (n > 0) {
    printf("%d\n", n);
    n -= 1;
}
```

Use `while` when you don't know up-front how many iterations you'll need
— for example, reading until end-of-file or until a sentinel value.

## `do { ... } while`

```c
int choice;
do {
    choice = read_menu_choice();
} while (choice < 1 || choice > 4);
```

Identical to `while` except the body runs at least once. Useful for
input-validation loops where you must prompt before you can test the
result. Note the trailing semicolon after `while (...)` — easy to forget.

## `for`

```c
for (int i = 0; i < n; i += 1) {
    arr[i] = 0;
}
```

Read this as: *initialise*, then *test*, then run the body, then *step*.
The variable `i` is scoped to the loop in C99/C11 — a frequent point of
confusion for students who learned C89.

The course style guide prefers `i += 1` over `i++` in `for` headers
because the postfix form returns the old value, which trips up new C
programmers. Either is technically correct.

### Off-by-one

The two canonical mistakes:

```c
for (int i = 0; i <= n; i += 1) { arr[i] = 0; }   // writes one past end
for (int i = 1; i < n; i += 1)  { arr[i] = 0; }   // skips arr[0]
```

The mantra: "**zero-indexed, length-bounded.**" Loop from `0` to `< n`,
not from `1` to `<= n`. The first form maps directly to C array indexing
and to how `sizeof(arr) / sizeof(arr[0])` reads as a length.

## `break` and `continue`

- `break` exits the *innermost* loop or `switch`.
- `continue` skips the rest of the current iteration and goes to the
  step / test of the same loop.

```c
for (int i = 0; i < n; i += 1) {
    if (arr[i] < 0) continue;     // skip negatives
    if (arr[i] > 100) break;      // stop entirely once we see one
    process(arr[i]);
}
```

Neither escapes a function — for that, use `return`.

## `goto`

C has it. CPTS 121 forbids it on lab submissions except for a single
documented pattern: cleanup at the end of a function with multiple
allocation steps. Even then, prefer breaking the function into smaller
ones first. If you find yourself reaching for `goto`, your function is
probably too long.

## Worked example

A loop that reads grades until the user enters `-1`, then prints the
average:

```c
double total = 0.0;
int count = 0;
double grade = 0.0;

while (1) {
    printf("grade (or -1 to quit): ");
    if (scanf("%lf", &grade) != 1) {
        fprintf(stderr, "bad input\n");
        return 1;
    }
    if (grade < 0) break;
    total += grade;
    count += 1;
}

if (count == 0) {
    printf("no grades entered\n");
} else {
    printf("average = %.2f\n", total / count);
}
```

Notice the guard before dividing — `count == 0` would otherwise be
integer-divide-by-zero, which is undefined behavior. Defensive coding
like this is the difference between an A and a C on lab grading.

## Common autograder traps

- **Sentinel inside the average.** Forgetting `if (grade < 0) break;`
  *before* `total += grade` adds `-1` to the running sum.
- **`if (x = 0)` instead of `if (x == 0)`.** The first assigns and
  always evaluates to false; the second compares.
- **Integer division.** `(a + b) / 2` where both are `int` truncates.
  Cast one to `double` if you need a real average.

## Where to next

Once `for` and `while` feel mechanical, see {{GUIDE:cpts121-functions-and-scope}}
for how to break long bodies into named pieces, then
{{GUIDE:cpts121-arrays-and-strings}} which uses every construct here.

The lab README ({{FILE:wsu-cpts121-lab-readme}}) lists the input/output
format the grader expects for the first looping lab.
