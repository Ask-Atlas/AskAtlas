---
slug: wsu-cpts121-final-review-slides
title: "Final Exam Review Slides (CPTS 121)"
mime: application/vnd.openxmlformats-officedocument.presentationml.presentation
filename: final-review-slides.pptx
course: wsu/cpts121
description: "Slide-friendly final review: one concept per slide, bullet-heavy, optimized for pandoc-to-pptx."
author_role: bot
---

## Why C Is Different

- Manual memory management — you allocate, you free
- No bounds checking — your code must be correct
- Compiled to machine code — no runtime safety net
- Fast, small, predictable

## The Compilation Pipeline

- Preprocessor: handles `#include`, `#define`
- Compiler: turns `.c` into `.o` object files
- Linker: combines objects into an executable
- Errors live in different stages — read the message carefully

## Variables and Scope

- Declared with a type: `int x;`
- Block scope — visible inside `{ }`
- Global variables exist for the whole program
- Prefer local; minimize globals

## Integer vs Float Division

- `7 / 2` is `3` (integer)
- `7.0 / 2` is `3.5` (one operand is double)
- Cast when you need float precision: `(double)a / b`

## Common Operators

- Arithmetic: `+ - * / %`
- Comparison: `== != < > <= >=`
- Logical: `&& || !`
- Assignment shortcuts: `+= -= *= /=`

## If, Else, Else-If

- One condition at a time, top to bottom
- Use braces even for one-line branches
- Test edge cases: zero, negative, boundary

## Loops

- `for` when count is known
- `while` when condition-driven
- `do-while` runs body at least once
- `break` exits, `continue` skips to next iteration

## Functions: Why

- Break big problems into small ones
- Reuse code across the program
- Name things so readers understand intent

## Functions: How

- Declare prototype above `main`
- Define with body below or in another file
- Parameters are local copies (pass by value)

## Pass By Value

- The function gets a copy
- Changing the parameter does not change the caller
- Default behavior in C

## Pass By Pointer

- Pass the address: `f(&x)`
- Function mutates via `*p = ...`
- The only way to change a caller's variable

## Arrays: The Basics

- Fixed size, zero-indexed
- Contiguous memory
- `int a[5];` reserves 5 ints
- `a[0]` through `a[4]` are valid

## Arrays In Functions

- Pass the pointer and the length
- `sizeof(a)` inside the function lies
- Never return a stack array

## C-Strings

- Array of chars ending in `'\0'`
- Length = chars before `'\0'`
- Always budget room for the terminator

## String Functions

- `strlen` counts chars (not `\0`)
- `strcmp` returns 0 on equal
- `strcpy` has no bounds check
- Prefer `snprintf` for building strings

## Pointers

- Variable that holds an address
- `&x` = address of x
- `*p` = value at address p
- `NULL` = "points to nothing"

## Memory Layout

- Stack: local variables, function calls
- Heap: `malloc` / `free` region
- Globals: fixed, program lifetime
- Stack is fast, heap is flexible

## Dynamic Memory

- `malloc(n)` asks for n bytes
- Returns pointer, or `NULL` on failure
- Every `malloc` needs exactly one `free`
- Set pointer to `NULL` after freeing

## Structs

- Group related fields into one type
- Access with dot: `p.x`
- Access through pointer with arrow: `pp->x`
- Great for records, less typing

## File I/O

- `fopen(path, "r")` or `"w"`
- Check for `NULL` return
- `fscanf`, `fprintf`, `fgets`, `fputs`
- Always `fclose` when done

## Debugging Tips

- Read the compiler error carefully
- Print variable values at checkpoints
- Rubber-duck the logic line by line
- Reduce to the smallest failing example

## Exam-Day Strategy

- Skim all questions first
- Answer easy ones to build momentum
- Trace small code examples by hand
- Manage time, leave five minutes to review

## One Last Thing

- You learned a language computers actually run
- Every modern language borrows ideas from C
- Struggling with pointers is normal
- You got this — good luck
