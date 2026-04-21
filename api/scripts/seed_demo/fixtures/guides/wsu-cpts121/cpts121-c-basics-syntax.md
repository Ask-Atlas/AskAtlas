---
slug: cpts121-c-basics-syntax
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "C Basics and Syntax — CPTS 121 Week 1"
description: "Translation units, types, the preprocessor, and the small set of syntax rules CPTS 121 will hold you to from day one."
tags: ["c", "basics", "syntax", "week-1", "intro"]
author_role: bot
attached_files:
  - wsu-cpts121-lab-readme
  - wikimedia-modern-c-gustedt-pdf
attached_resources: []
---

# C Basics and Syntax

CPTS 121 is most students' first exposure to a language that *won't* hold
your hand. C does what you tell it — including the wrong thing — and the
compiler is willing to let you walk off a cliff with only a warning.
This guide walks the syntax you'll be expected to read fluently by the
end of week 2.

## The shape of a C program

Every C program has exactly one `main`. The minimal compilable file looks
like this:

```c
#include <stdio.h>

int main(void) {
    printf("Hello, CPTS 121!\n");
    return 0;
}
```

A few things to notice that the lectures will hammer on:

- `#include <stdio.h>` is a **preprocessor** directive, not a statement.
  No semicolon. It runs before the compiler ever sees your code.
- `int main(void)` declares `main` as returning `int` and taking no
  arguments. Use `void` in the parameter list — leaving it empty is a
  pre-C99 wart that the course style guide explicitly bans.
- `return 0;` from `main` signals success. Non-zero is the convention for
  failure, and the shell uses that to know whether your program worked.

## Types you'll actually use

CPTS 121 sticks to a small subset of the type system at first:

| Type | Typical size | Use for |
|---|---|---|
| `int` | 4 bytes | counters, indices, most integer math |
| `double` | 8 bytes | scientific values, anything with decimals |
| `char` | 1 byte | a single byte (often an ASCII character) |
| `size_t` | platform | array lengths, anything from `sizeof` |

Stay away from `float` unless an assignment specifically asks for it.
`double` is faster on modern hardware and avoids a class of precision
bugs that the autograder *will* catch.

## Declarations vs definitions

```c
int x;          // definition: allocates storage for x
int x = 10;     // definition + initialization
extern int x;   // declaration only: "x exists somewhere"
```

The course's lab style guide requires that **every variable be initialised
at the point of declaration**. Uninitialised stack variables hold garbage
in C, and reading garbage is undefined behavior. The cost of typing `= 0`
is much smaller than the cost of debugging the alternative.

## Statements and expressions

Every statement ends in a semicolon. Forgetting one is the #1 cause of
the cryptic `expected ';' before ...` error you'll see in office hours.

```c
int total = 0;             // statement
total = total + value;     // statement (the assignment is an expression)
if (total > 100) total = 100;  // also a statement
```

Compound statements use braces:

```c
if (x > 0) {
    printf("positive\n");
    x -= 1;
}
```

The course style guide requires braces *even on single-line bodies*.
This rule is not negotiable on lab submissions — Apple's "goto fail"
bug is the canonical example of why.

## The preprocessor in 90 seconds

Three things you'll see constantly:

```c
#include <stdio.h>          // pulls in a system header
#include "my_header.h"      // pulls in a header from this project
#define MAX_NAME 64         // text substitution before compilation
```

Macros like `MAX_NAME` are not variables — they are textual replacements
done by the preprocessor. You can't take their address, you can't change
them at runtime, and they don't respect scope. Use them for compile-time
constants and conditional compilation, nothing else.

For named constants inside a function, prefer `const`:

```c
const int max_attempts = 3;
```

## Compiling on the lab machines

The CS department's Linux boxes have `gcc` installed. The flags you
should use for *every* lab are:

```sh
gcc -std=c11 -Wall -Wextra -Werror -g -O0 -o lab1 lab1.c
```

Breaking that down:

- `-std=c11` picks the language version the course uses.
- `-Wall -Wextra` turns on the warnings worth reading.
- `-Werror` promotes warnings to errors. Submissions that warn under
  `-Wall -Wextra` lose style points; making them errors keeps you honest.
- `-g` keeps debug symbols so `gdb` can show you source lines.
- `-O0` disables optimization so the variables you set actually exist
  when you stop in the debugger.

The lab README — see {{FILE:wsu-cpts121-lab-readme}} — has the canonical
`Makefile` you should copy into every assignment.

## Comments

```c
// single line — fine
/* multi-line; the older C89 style */
```

The lab style guide expects a header block at the top of every `.c` file
with your name, WSU ID, the lab number, and a brief description. Missing
that header is a deduction.

## Common day-1 errors

- `printf("%d", x)` where `x` is a `double` prints garbage. Mismatched
  format specifiers are undefined behavior. Use `%f` or `%g` for `double`.
- `if (x = 5)` is an *assignment*, not a comparison. Always `==` for
  equality. Modern compilers warn on this if the parens look wrong.
- `int x = 5/2;` is `2`, not `2.5`. Integer division throws away the
  remainder. Cast one operand to `double` if you want a real answer.
- A semicolon after `if` is a no-op body: `if (x > 0); { ... }` runs
  the brace block unconditionally.

## Where to next

Once the syntax above feels mechanical, move on to {{GUIDE:cpts121-control-flow}}
to see how `if`, `while`, and `for` compose, then to
{{GUIDE:cpts121-functions-and-scope}} for how to break programs into
pieces.

For the canonical book treatment of everything in this guide, see
chapters 1–3 of Gustedt's *Modern C* in this guide's attachments.
