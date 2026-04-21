---
slug: cpts121-structs-and-typedefs
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "Structs and Typedefs in C — CPTS 121"
description: "Compose related fields into a single named type, pass them around, and stop juggling parallel arrays."
tags: ["c", "structs", "typedef", "data-types", "week-7"]
author_role: bot
quiz_slug: cpts121-structs-and-typedefs-quiz
attached_files:
  - wsu-cpts121-arrays-and-strings
  - wikimedia-modern-c-gustedt-pdf
attached_resources: []
---

# Structs and Typedefs in C

By the time the course gets to the second project, the temptation to
keep three parallel arrays — `names[]`, `ages[]`, `gpas[]` — is going to
get you in trouble. Structs are the fix. They bundle related fields
into a single value you can name, pass around, and reason about as one
thing.

## Declaring a struct

```c
struct student {
    char name[64];
    int  age;
    double gpa;
};
```

This declares a *type* called `struct student`. No memory is allocated
yet — you have to instantiate it:

```c
struct student s = {"Cougar Coug", 20, 3.7};
printf("%s, %d, %.2f\n", s.name, s.age, s.gpa);
```

Field access uses the `.` operator. The order of the initialiser fields
must match the order of the declaration. To be explicit (and resilient
to reordering), use designated initialisers:

```c
struct student s = {
    .name = "Cougar Coug",
    .age  = 20,
    .gpa  = 3.7
};
```

Designated initialisers also default any field you skip to zero, which
is the safe choice.

## `typedef` removes the `struct ` prefix

Writing `struct student` everywhere is noisy. `typedef` gives the type
a new alias:

```c
typedef struct student {
    char name[64];
    int  age;
    double gpa;
} student_t;

student_t s = {.name = "Coug", .age = 20, .gpa = 3.7};
```

The `_t` suffix is a CPTS 121 convention. It signals "this is a typedef
name" and avoids collisions with C's reserved-name rules.

You can also typedef anonymously:

```c
typedef struct {
    int x, y;
} point_t;
```

The struct itself has no tag — only the `typedef` name is usable. Fine
for self-contained types; bad if the struct ever needs to refer to
itself (linked-list nodes, for example).

## Passing structs to functions

A struct is a value, just like an `int`. It is **passed by copy**:

```c
void print_student(student_t s) {
    printf("%s\n", s.name);
}
```

For small structs (a few `int`s) that's fine. For larger ones it's
wasteful to copy — pass a pointer instead:

```c
void print_student(const student_t *s) {
    printf("%s\n", s->name);  // -> dereferences before .field
}
```

The `->` operator is shorthand for `(*p).field` and is the only
practical way to write that pattern. Mark the pointer `const` if you
won't modify the struct — same idea as for any other input parameter.

## Returning structs from functions

C lets you return a struct by value. The compiler manages the copy.

```c
student_t make_default(void) {
    student_t s = {.name = "TBD", .age = 0, .gpa = 0.0};
    return s;
}

student_t s = make_default();
```

This is much safer than the alternative — returning a pointer to a
local struct, which dangles the moment the function returns (see
{{GUIDE:cpts121-functions-and-scope}}).

## Arrays of structs

The natural replacement for parallel arrays:

```c
student_t roster[30];
roster[0] = (student_t){.name = "Coug", .age = 20, .gpa = 3.7};
roster[1] = (student_t){.name = "Butch", .age = 22, .gpa = 3.4};

for (int i = 0; i < 2; i += 1) {
    printf("%s, %.2f\n", roster[i].name, roster[i].gpa);
}
```

The compound literal `(student_t){...}` is a struct value created on
the spot, useful for assigning into an existing slot.

## Nested structs

Structs can contain other structs:

```c
typedef struct {
    int year;
    int month;
    int day;
} date_t;

typedef struct {
    char name[64];
    date_t birthday;
} person_t;

person_t p = {.name = "Coug", .birthday = {.year = 2003, .month = 9, .day = 1}};
printf("%d\n", p.birthday.year);
```

The `.field.field` chain reads naturally. With pointers, the chain
becomes `p->birthday.year` (dereference then access), which is also
fine.

## Bitwise size and padding

Don't assume `sizeof(struct foo)` equals the sum of its fields' sizes.
The compiler inserts **padding** so each field lands on a properly
aligned address.

```c
struct s {
    char  c;       // 1 byte
    int   i;       // 4 bytes, aligned to 4
    char  d;       // 1 byte
};
// sizeof(struct s) is typically 12 on x86-64, not 6
```

CPTS 121 doesn't go deep on alignment, but the lectures will mention
it as a "this is why your struct is bigger than you expected" footnote.

## Common gotchas

- **`p.field` vs `p->field`.** If `p` is a struct, use `.`. If `p` is a
  pointer to a struct, use `->`. Mixing them gives a compile error
  that's easy to fix once you read it carefully.
- **Forgetting the `typedef` name.** Without `typedef`, you must write
  `struct student s;` everywhere — the bare `student s;` won't compile.
- **Comparing structs with `==`.** Not allowed. Compare field-by-field
  or write a helper.
- **`scanf` into a `char[64]` field.** Use `%63s` (note the size cap)
  to avoid overflowing the buffer. Even better, use `fgets` and parse.

## Worked example

A function that finds the highest-GPA student in a roster:

```c
const student_t *top_student(const student_t *roster, size_t n) {
    if (n == 0) return NULL;
    const student_t *best = &roster[0];
    for (size_t i = 1; i < n; i += 1) {
        if (roster[i].gpa > best->gpa) {
            best = &roster[i];
        }
    }
    return best;
}
```

Returns a `const` pointer into the caller's array — no copy, and the
caller can't accidentally modify the result through it. Returns `NULL`
on the empty case so callers can guard.

## Practice

Once you can declare, initialise, and pass structs without thinking,
take {{QUIZ:cpts121-structs-and-typedefs-quiz}}. The arrays-and-strings
cheatsheet ({{FILE:wsu-cpts121-arrays-and-strings}}) is a useful side
reference for the array-of-structs section.

## Where to next

The natural next topic is {{GUIDE:cpts121-dynamic-memory-allocation}} —
allocating an array of structs whose size is decided at runtime. After
that, {{GUIDE:cpts121-file-io}} shows how to serialise structs to and
from disk.
