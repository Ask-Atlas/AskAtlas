---
slug: cpts121-file-io
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "121"
title: "File I/O in C — CPTS 121"
description: "fopen, fclose, fgets, fprintf, and the patterns that keep your file handles from leaking."
tags: ["c", "file-io", "fopen", "fgets", "stdio", "week-9"]
author_role: bot
attached_files:
  - wsu-cpts121-lab-readme
  - wikimedia-modern-c-gustedt-pdf
attached_resources: []
---

# File I/O in C

By project 2 you'll be reading data from a file the user names on the
command line. C's `<stdio.h>` exposes a small set of functions for
that. This guide walks the ones you'll actually use and the patterns
that keep handles from leaking.

## Opening and closing

```c
FILE *f = fopen("input.txt", "r");
if (f == NULL) {
    fprintf(stderr, "could not open input.txt: %s\n", strerror(errno));
    return 1;
}

// ... use f ...

fclose(f);
f = NULL;
```

`fopen` returns `NULL` on failure (file missing, permission denied,
etc.). Always check. The errno + `strerror` pattern is what the lab
style guide expects for diagnostic messages.

## Modes

| Mode | Meaning |
|---|---|
| `"r"` | Read text. Fails if the file doesn't exist. |
| `"w"` | Write text. **Truncates** the file if it exists. |
| `"a"` | Append text. Creates the file if missing. |
| `"r+"` | Read + write. File must exist. |
| `"rb"` / `"wb"` | Binary variants — required on Windows for non-text data. |

Lab convention: open text files with `"r"` / `"w"`. Use `"rb"` / `"wb"`
only when you're reading raw bytes (an image, a serialised struct).

## Reading line by line

The right way to read text files in CPTS 121:

```c
char buf[256];
while (fgets(buf, sizeof(buf), f) != NULL) {
    // buf includes the trailing \n if it fit
    process(buf);
}
```

`fgets` reads at most `sizeof(buf) - 1` bytes (leaving room for the
NUL), stops at end-of-line or EOF, and returns `NULL` when there's
nothing left. It is the safe equivalent of the deprecated `gets()` —
**never use `gets()`**, which has no length cap and is gone from C11.

To strip the trailing newline:

```c
size_t len = strlen(buf);
if (len > 0 && buf[len - 1] == '\n') {
    buf[len - 1] = '\0';
}
```

## Reading structured data

```c
int  id;
char name[64];
double gpa;

while (fscanf(f, "%d %63s %lf", &id, name, &gpa) == 3) {
    printf("%d %s %.2f\n", id, name, gpa);
}
```

Two patterns the lectures will hammer:

- **Always check the return value.** `fscanf` returns the number of
  fields successfully read. Use it as the loop condition rather than
  `feof(f)` — see "Don't loop on feof" below.
- **Always cap `%s` and `%[`.** `%s` with no width specifier is the
  same overflow bug as `gets`. Use `%63s` to read at most 63 chars
  into a 64-byte buffer.

## Writing

```c
fprintf(f, "%d,%s,%.2f\n", id, name, gpa);
```

Same format-string rules as `printf`. The CPTS 121 grading scripts
typically diff your output against an expected file, so trailing
whitespace, newline conventions, and decimal precision all matter.

## Closing on every exit path

Every successful `fopen` must have a matching `fclose`. The
single-exit-point pattern makes this manageable:

```c
int run(const char *path) {
    int rc = 0;
    FILE *f = fopen(path, "r");
    if (f == NULL) return 1;

    if (read_header(f) != 0) {
        rc = 2;
        goto cleanup;
    }
    if (read_body(f) != 0) {
        rc = 3;
        goto cleanup;
    }

cleanup:
    fclose(f);
    return rc;
}
```

This is the *one* place CPTS 121 will tolerate `goto` — cleanup at
function exit. For shorter functions, manual `fclose` on each return is
fine.

## Don't loop on `feof`

```c
// WRONG
while (!feof(f)) {
    fscanf(f, "%d", &x);
    process(x);
}
```

`feof` becomes true *only after* a read attempt has failed. The loop
above runs one extra iteration with stale `x`. The right idiom is to
loop on the read function's return value:

```c
while (fscanf(f, "%d", &x) == 1) {
    process(x);
}
```

## Binary I/O

For raw bytes (e.g. dumping a struct to disk):

```c
FILE *f = fopen("save.bin", "wb");
fwrite(&record, sizeof(record), 1, f);
fclose(f);
```

`fwrite` returns the number of *items* written (not bytes). Same for
`fread`. Binary files are not portable across machines with different
endianness or struct padding — CPTS 121 uses them only for short-lived
data on the same lab box.

## Command-line file paths

```c
int main(int argc, char *argv[]) {
    if (argc < 2) {
        fprintf(stderr, "usage: %s <input-file>\n", argv[0]);
        return 1;
    }
    FILE *f = fopen(argv[1], "r");
    // ...
}
```

`argv[0]` is the program name; `argv[1]` is the first user-provided
argument. The grader often checks that you print a usage message when
called with no arguments — read the lab spec ({{FILE:wsu-cpts121-lab-readme}})
for the exact wording it expects.

## Common bugs

- **Mode mismatch.** Opening with `"r"` then trying to write returns
  garbage and may not even error. Check the mode against what the
  function does.
- **Forgetting to `fclose`.** The OS will reclaim the descriptor at
  program exit, but lab graders run leak checks and notice. Each
  `fopen` needs a matching `fclose`.
- **Buffered output.** `printf` and `fprintf` are line-buffered to a
  terminal but **block-buffered** to a file. Output may not appear until
  `fclose` or `fflush`. Call `fflush(f)` if you need the bytes on disk
  immediately.
- **Reading past EOF.** Always check the return value of `fread` /
  `fscanf` / `fgets` before using whatever was supposed to be read.

## Worked example: count lines in a file

```c
int main(int argc, char *argv[]) {
    if (argc < 2) {
        fprintf(stderr, "usage: %s <file>\n", argv[0]);
        return 1;
    }

    FILE *f = fopen(argv[1], "r");
    if (f == NULL) {
        fprintf(stderr, "open %s: %s\n", argv[1], strerror(errno));
        return 1;
    }

    char buf[256];
    long lines = 0;
    while (fgets(buf, sizeof(buf), f) != NULL) {
        lines += 1;
    }

    fclose(f);
    printf("%ld\n", lines);
    return 0;
}
```

Note: this counts *physical lines that ended with a newline OR ended
the file*. Lines longer than 255 chars get split across multiple `fgets`
calls — for the "exact wc -l" answer, check whether the buffer ended in
`'\n'` before incrementing. CPTS 121 typically calls that out as a
bonus question.

## Where to next

For the data-structures side, see
{{GUIDE:cpts121-structs-and-typedefs}} — most file-I/O labs end with
a roster of structs read from disk. For debugging the inevitable
"why is my file empty?" issue, jump to
{{GUIDE:cpts121-debugging-with-gdb}}.
