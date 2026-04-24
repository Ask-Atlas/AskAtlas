---
slug: wsu-cpts121-midterm-prep-template
title: "Midterm Prep Template (CPTS 121)"
mime: application/vnd.openxmlformats-officedocument.wordprocessingml.document
filename: midterm-prep-template.docx
course: wsu/cpts121
description: "Fill-in-the-blank study template covering every midterm topic area — print and complete by hand."
author_role: bot
---

## How to Use This Template

Print this out or duplicate it digitally. Fill in each section **from memory first**, then verify against your notes and textbook. Topics you cannot recall are your study priorities for the next session. Budget one hour per section.

---

## Section 1: Data Types and Operators

Primitive types and their typical sizes on a 64-bit system:

| Type | Size (bytes) | Format specifier | Example literal |
|---|---|---|---|
| `char` | _____ | _____ | _____ |
| `int` | _____ | _____ | _____ |
| `float` | _____ | _____ | _____ |
| `double` | _____ | _____ | _____ |

What is the value of `7 / 2` in C? __________

What is the value of `7.0 / 2`? __________

Explain integer overflow in one sentence: ______________________________

---

## Section 2: Control Flow

Write a `for` loop that prints even numbers from 2 to 20 inclusive:

```c

```

Write the equivalent `while` loop:

```c

```

When would you use `do-while` instead of `while`? ______________________________

What does `break` do inside a `switch`? ______________________________

What does `continue` do inside a loop? ______________________________

---

## Section 3: Functions

Complete the signature of a function that takes two integers and returns their greatest common divisor:

```c
int _____ ( _____ , _____ ) {
    // your implementation
}
```

What is the difference between a function **declaration** (prototype) and a function **definition**? ______________________________

What happens if you call a function before its prototype is visible? ______________________________

---

## Section 4: Arrays

Declare a 10-element array of doubles initialized to zero:

```c

```

Write a loop that finds the maximum value in `int a[n]`:

```c

```

Why can you not write `a = b;` to copy one array to another? ______________________________

---

## Section 5: Strings

How do you detect the end of a C-string? ______________________________

Why is `if (s1 == s2)` wrong for comparing strings? ______________________________

Write the correct comparison: ______________________________

Name one safer alternative to `gets` and explain why: ______________________________

---

## Section 6: Pointers

If `int x = 5; int *p = &x;`, what is the value of `*p`? _____

What does `p` hold? ______________________________

After `*p = 10;`, what is `x`? _____

Write a `swap` function that takes two `int*` and exchanges their values:

```c

```

---

## Section 7: File I/O

How do you open a file for reading? ______________________________

How do you check if the open succeeded? ______________________________

Why is it important to `fclose` every file you open? ______________________________

---

## Section 8: Debugging Reflection

Pick one bug you encountered in lab this semester. Describe:

- The symptom: ______________________________
- The root cause: ______________________________
- How you found it: ______________________________
- What you will check first next time: ______________________________

---

## Final Self-Check

- [ ] Every section above is filled in from memory
- [ ] Weak areas are flagged for re-study
- [ ] At least two practice problems attempted per section
- [ ] Sleep scheduled before the exam
