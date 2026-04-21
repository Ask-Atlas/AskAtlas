---
slug: cpts121-final-review
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "Final Exam Review — CPTS 121"
description: "End-to-end review of every CPTS 121 topic — pointers, dynamic memory, structs, file I/O — with the question patterns the cumulative final actually uses."
tags: ["c", "final-exam", "review", "exam-prep", "cumulative"]
author_role: bot
quiz_slug: cpts121-final-review-quiz
attached_files:
  - wsu-cpts121-final-review-slides
  - wsu-cpts121-pointers-cheatsheet
  - wsu-cpts121-arrays-and-strings
attached_resources: []
---

# Final Exam Review

The CPTS 121 final is cumulative but weighted toward the second half of
the course: pointers, dynamic memory, structs, and file I/O. This
guide walks each topic with the question shape the exam reliably uses
and points you at the dedicated guide for deeper study.

The official slide deck for the review session is attached as
{{FILE:wsu-cpts121-final-review-slides}}. It mirrors the topic order
below.

## Part 1 — first-half topics (lightly tested)

The first-half material that recurs on the final:

- **Type promotion and integer division.** `5 / 2` is `2`. `(double)5 / 2`
  is `2.5`. The "cast one operand" trick is worth one or two questions.
- **Off-by-one in `for` loops.** Always `i < n`, never `i <= n`.
  See {{GUIDE:cpts121-control-flow}} for the canonical loop shape.
- **Pass-by-value vs pass-by-reference.** A `void f(int x) { x = 0; }`
  cannot modify the caller's variable. To do that, take a pointer.
  See {{GUIDE:cpts121-functions-and-scope}}.

These appear as quick warm-ups; expect 4-6 questions total.

## Part 2 — pointers and arrays

Roughly a third of the final. Be ready for:

```c
int arr[5] = {10, 20, 30, 40, 50};
int *p = arr;
printf("%d\n", *(p + 2));   // 30
printf("%d\n", p[3]);       // 40
```

Pointer arithmetic and array indexing are equivalent: `p[i]` is
`*(p + i)`. The exam will often phrase a question in one form and the
choices in the other.

The arrays-and-strings cheatsheet ({{FILE:wsu-cpts121-arrays-and-strings}})
has the diagrams. The pointers-cheatsheet PDF
({{FILE:wsu-cpts121-pointers-cheatsheet}}) covers the memory model.

Free-response question that recurs verbatim every year: **"Write a
function `int find(const int *arr, size_t n, int target)` that returns
the index of `target` in `arr`, or `-1` if not present."**

```c
int find(const int *arr, size_t n, int target) {
    for (size_t i = 0; i < n; i += 1) {
        if (arr[i] == target) return (int)i;
    }
    return -1;
}
```

Notice the `const`, the `size_t` for length, and the cast on the
return. Style points come off if any of those are wrong.

## Part 3 — dynamic memory

Expect a "trace through this and tell me what's leaked" question:

```c
int *a = malloc(10 * sizeof(*a));
int *b = malloc(20 * sizeof(*b));
a = b;        // <-- leaks the original a
free(b);
free(a);      // double-free of the same allocation
```

Both errors in five lines. Be ready to spot and explain them.

The full guide is {{GUIDE:cpts121-dynamic-memory-allocation}}. The
reference patterns to memorise:

- `malloc` returns `NULL` on failure — always check.
- `realloc` may return a *new* pointer; assign through a temporary so
  you don't lose the original on failure.
- Set `p = NULL` after `free(p)` so a second `free` is a no-op.

## Part 4 — structs and file I/O

Combined into one or two large free-response questions: "given a CSV
of student records, read them into an array of `student_t` and print
the highest GPA."

The skeleton:

```c
typedef struct {
    char name[64];
    double gpa;
} student_t;

int main(int argc, char *argv[]) {
    if (argc < 2) {
        fprintf(stderr, "usage: %s <file>\n", argv[0]);
        return 1;
    }
    FILE *f = fopen(argv[1], "r");
    if (f == NULL) {
        perror(argv[1]);
        return 1;
    }

    student_t roster[100];
    size_t n = 0;
    while (n < 100 && fscanf(f, "%63[^,],%lf\n",
                             roster[n].name, &roster[n].gpa) == 2) {
        n += 1;
    }
    fclose(f);

    if (n == 0) {
        printf("no records\n");
        return 0;
    }

    size_t best = 0;
    for (size_t i = 1; i < n; i += 1) {
        if (roster[i].gpa > roster[best].gpa) best = i;
    }
    printf("%s %.2f\n", roster[best].name, roster[best].gpa);
    return 0;
}
```

The patterns the rubric checks:

- Argument-count check with usage message
- `fopen` null-check
- `%63[^,]` (the `^,` reads everything up to a comma) — the size cap
  is required
- `fscanf` return-value check (loop on `== 2`, not on `!feof`)
- `fclose` on every successful `fopen`
- Empty-input guard before computing the max

Review {{GUIDE:cpts121-structs-and-typedefs}} and
{{GUIDE:cpts121-file-io}} for the full mechanics.

## Part 5 — debugging

Often a "given this Valgrind output, identify the bug" question.
{{GUIDE:cpts121-debugging-with-gdb}} walks the canonical workflow.
The exam usually shows you something like:

```text
==1234== Invalid write of size 4
==1234==    at 0x4006A1: make_array (lab.c:7)
==1234==    by 0x40071F: main (lab.c:13)
==1234==  Address 0x52041d4 is 0 bytes after a block of size 20 alloc'd
```

You should be able to read this as: "an off-by-one wrote one element
past the end of a 20-byte (5 × 4) block allocated in `make_array`."
The fix is `i < n`, not `i <= n`.

## Final-week study plan

A recommended week-long plan:

| Day | Focus |
|---|---|
| Mon | Pointers + arrays. Walk every example in {{GUIDE:cpts121-pointers-cheatsheet}} and {{GUIDE:cpts121-arrays-and-strings}}. |
| Tue | Functions, scope, structs. Re-do the last roster lab from scratch. |
| Wed | Dynamic memory. Write the growable-array pattern from memory; run under Valgrind. |
| Thu | File I/O + debugging. Take the practice quiz {{QUIZ:cpts121-final-review-quiz}}. |
| Fri | Cumulative practice problems on paper, no compiler. |
| Sat | Light review, sleep early. |

Don't pull an all-nighter the night before. The exam tests pattern
recognition, not memorisation, and tired pattern-recognition is bad
pattern-recognition.

## Day-of checklist

- Picture ID + WSU ID
- Pen + pencil
- Layered clothes (the lecture hall is unpredictable)
- Empty mind for the first read-through, full mind for the second

## Practice

The cumulative practice quiz is {{QUIZ:cpts121-final-review-quiz}}.
The pointers warm-up is {{QUIZ:cpts121-pointers-quiz}}, and the
arrays-focused one is {{QUIZ:cpts121-arrays-and-strings-quiz}}. If
you can score 80%+ on all three without notes, you're in good shape.
