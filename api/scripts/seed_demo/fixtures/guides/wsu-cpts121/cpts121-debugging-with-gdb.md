---
slug: cpts121-debugging-with-gdb
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "Debugging C with GDB — CPTS 121"
description: "Stop printf-debugging: a working subset of GDB you can learn in an hour and use for the rest of the course."
tags: ["c", "debugging", "gdb", "valgrind", "tooling"]
author_role: bot
attached_files:
  - wsu-cpts121-lab-readme
  - unsplash-png-abstract-tech-11
attached_resources: []
---

# Debugging C with GDB

Most CPTS 121 students try to debug by sprinkling `printf` calls
through their code. It works, sort of, until the bug is "the program
crashes before reaching my `printf`." Then you need a real debugger.
GDB has a hundred commands; you need about ten.

## Compile with debug symbols

GDB is much more useful when it can show you source lines and variable
names. Compile with `-g -O0`:

```sh
gcc -std=c11 -Wall -Wextra -g -O0 -o lab5 lab5.c
```

`-g` keeps debug symbols. `-O0` disables optimisation, which is
critical — optimised code rearranges and removes variables, making the
debugger lie to you about their values. Always debug a `-O0` build.

## Starting GDB

```sh
gdb ./lab5
```

That gives you a `(gdb)` prompt. To run with command-line arguments:

```text
(gdb) run input.txt
```

Or if you want to pass arguments only once:

```sh
gdb --args ./lab5 input.txt
(gdb) run
```

## The ten commands you actually need

| Command | What it does |
|---|---|
| `break <line>` | Stop when execution hits `<line>` in the current file |
| `break <fn>` | Stop on entry to `<fn>` |
| `run [args...]` | Start the program |
| `continue` (`c`) | Resume until the next breakpoint or exit |
| `next` (`n`) | Step to the next source line, jumping over function calls |
| `step` (`s`) | Step into the next function call |
| `print <expr>` (`p`) | Show the value of `<expr>` |
| `backtrace` (`bt`) | Show the call stack from the current frame down |
| `list` (`l`) | Show source lines around the current one |
| `quit` (`q`) | Exit GDB |

That's enough to debug 90% of CPTS 121 bugs. Memorise these before
adding any others.

## Worked example: a segfault

Suppose this program crashes:

```c
#include <stdio.h>
#include <stdlib.h>

int *make_array(int n) {
    int *arr = malloc(n * sizeof(int));
    for (int i = 0; i <= n; i += 1) {
        arr[i] = i;
    }
    return arr;
}

int main(void) {
    int *a = make_array(5);
    printf("%d\n", a[2]);
    free(a);
    return 0;
}
```

A run inside GDB:

```text
(gdb) run
Program received signal SIGSEGV, Segmentation fault.
0x0000... in make_array (n=5) at lab.c:7
7           arr[i] = i;
(gdb) print i
$1 = 5
(gdb) print n
$2 = 5
```

`i` is `5`, but valid indices are `0..4`. The off-by-one is the bug —
`i <= n` should be `i < n`. The debugger pointed straight at the line
and showed the relevant variables. No `printf`s required.

## Watching variables change

```text
(gdb) watch x
Hardware watchpoint 2: x
(gdb) continue
Hardware watchpoint 2: x
Old value = 0
New value = 42
```

Watchpoints fire whenever the value of an expression changes. Useful
for "this variable is somehow getting set to the wrong thing — when?"

## Inspecting the call stack

When the program crashes deep in a helper function, `backtrace` shows
the chain of calls that got there:

```text
(gdb) bt
#0  process_record (r=...) at lab.c:42
#1  0x0000... in main (argc=2, argv=...) at lab.c:88
```

Use `frame N` to switch into a specific frame and `print` its locals.

## Conditional breakpoints

If a bug only manifests on the 1000th iteration:

```text
(gdb) break lab.c:42 if i == 1000
```

Saves you from `continue`-ing 999 times.

## Pair GDB with Valgrind

GDB shows you crashes in the moment. Valgrind catches the leaks and
use-after-frees that *don't* crash but slowly corrupt your program.
The CPTS 121 grading rubric checks for both.

```sh
valgrind --leak-check=full ./lab5
```

A clean Valgrind run has zero leaks and zero "Invalid read/write of
size N" lines. If Valgrind is unhappy, GDB is your next stop — set a
breakpoint at the line Valgrind flagged and inspect the variables.

For more on the underlying allocator behavior, see
{{GUIDE:cpts121-dynamic-memory-allocation}} and
{{GUIDE:cpts121-pointers-cheatsheet}}.

## Common workflow

1. Crash? Reproduce inside GDB, look at the failing line and locals.
2. Wrong output? Set a breakpoint near the discrepancy, step forward.
3. Hang? Run, then `Ctrl-C` to interrupt, then `bt` to see where it's
   stuck.
4. Leak? Run Valgrind, then re-run the relevant code in GDB to find
   the missing `free`.

The lab README ({{FILE:wsu-cpts121-lab-readme}}) has the canonical
`Makefile` target `make debug` which compiles with `-g -O0`. Use it.

## A few quality-of-life tips

- `tui enable` (or `Ctrl-X A`) opens a split-pane source view.
  GDB feels much more pleasant once you can see the source as you step.
- `set print pretty on` formats structs across multiple lines.
- `display x` is like `print` but re-prints `x` after every step. Useful
  for watching one variable through a loop.
- `Ctrl-D` at the prompt exits cleanly.

## Where to next

If your bug is in dynamically-allocated memory, the
{{GUIDE:cpts121-dynamic-memory-allocation}} guide is the right next
read — most heap bugs are easier to diagnose once you know what
`malloc` and `free` actually do. For midterm-2 review of all this
together, see {{GUIDE:cpts121-final-review}}.
