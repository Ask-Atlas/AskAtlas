---
slug: wsu-cpts121-kr-excerpt-commentary
title: "K&R Exercise Commentary: The getline Problem (CPTS 121)"
mime: application/epub+zip
filename: kr-excerpt-commentary.epub
course: wsu/cpts121
description: "Guided walkthrough of the classic K&R getline exercise — design choices, pitfalls, and modern updates."
author_role: bot
---

## Context: Why Read K&R at All

*The C Programming Language* by Brian Kernighan and Dennis Ritchie (the "K&R book") is the canonical C text. It is short, dense, and every exercise teaches something specific. Chapter 1 introduces a function called `getline` that reads one line from standard input into a caller-provided buffer and returns its length. This function appears again and again in later chapters, so understanding it early pays off.

Note: the K&R `getline` is **not** the same as the POSIX `getline` in `<stdio.h>`. Same name, different function. In modern code prefer `fgets` or the POSIX version. The K&R version is still worth studying as a teaching tool.

## The Function

The textbook version looks roughly like this:

```c
int getline(char s[], int lim) {
    int c, i;
    for (i = 0; i < lim - 1 && (c = getchar()) != EOF && c != '\n'; ++i)
        s[i] = c;
    if (c == '\n') {
        s[i] = c;
        ++i;
    }
    s[i] = '\0';
    return i;
}
```

A lot happens in those few lines. Let us unpack it.

## Design Choice 1: `lim - 1`

The loop stops at `lim - 1`, not `lim`. Why? Because every C-string needs a trailing `'\0'`. If the caller passes a buffer of size 100, the function must leave room for the terminator, which means at most 99 characters plus the null. Writing `s[99] = '\0'` is fine; writing `s[100] = '\0'` would be a buffer overflow. This one detail accounts for a large fraction of buffer-overflow bugs in the wild.

## Design Choice 2: The Three-Way Loop Exit

The `for` loop stops when any of three things happens: the buffer fills, `getchar` returns `EOF`, or a newline arrives. K&R combines them into one condition using short-circuit `&&`. Read it as:

- "Are we out of room?" → exit.
- "Has input ended?" → exit.
- "Did we just see a newline?" → exit.

The order matters. `getchar()` is only called if the first condition passes, which prevents us from reading more than we can store.

## Design Choice 3: Keeping the Newline

After the loop, if we exited because of `'\n'`, we deliberately write the newline into the buffer. This is a style choice: it lets the caller distinguish "I got a full line" from "I hit the buffer limit mid-line." It also matches how `fgets` behaves in the standard library. Some students prefer to strip the newline; do whichever your assignment requires and document it.

## Design Choice 4: Returning the Length

Returning the number of characters stored (not counting `'\0'`) has two nice properties. Zero means "empty line." A return value equal to `lim - 1` tells the caller "we hit the limit — the line may be truncated." Callers can check without scanning the buffer themselves.

## Pitfalls to Watch

- **Using `char` instead of `int` for `c`.** `getchar` returns `int` precisely so it can signal `EOF` (usually `-1`) distinctly from a valid byte. Storing it in a `char` can break the `EOF` comparison on platforms where `char` is unsigned.
- **Forgetting `'\0'`.** If you shorten the function and drop the final assignment, callers will read garbage.
- **Assuming input ends at `\n`.** Piping from a file may hit `EOF` without a trailing newline.

## Exercise

Extend `getline` so that it also reports via an out-parameter whether the line was truncated. Compare your version against the standard `fgets` behavior and explain the differences in comments. Both approaches are valid — the point is to make the trade-off explicit.
