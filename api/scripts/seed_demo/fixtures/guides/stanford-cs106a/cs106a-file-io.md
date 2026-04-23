---
slug: cs106a-file-io
course:
  ipeds_id: "243744"
  department: "CS"
  number: "106A"
title: "File I/O — CS 106A"
description: "Reading and writing text files in Python with context managers, path conventions, and CS 106A pset patterns."
tags: ["python", "file-io", "io", "context-managers"]
author_role: bot
attached_files:
  - stanford-cs106a-pset-setup-readme
attached_resources: []
---

# File I/O

Most of the back half of CS 106A involves reading text from a file:
word-count assignments, dictionary-based games, simple data analysis. The
mechanics are simple — Python's file API is one of the cleanest you'll
ever meet — but a few conventions matter for grading.

## Opening a file: the `with` statement

The canonical CS 106A pattern:

```python
with open("words.txt") as f:
    for line in f:
        print(line.rstrip())
```

The `with` statement is a **context manager**. When the indented block
exits — normally or via exception — it guarantees `f.close()` runs.
Without `with`, you'd have to remember to close the file yourself, and
forgetting causes resource leaks and (on Windows) "file is locked"
errors.

**Always use `with` for file I/O.** CS 106A grading deducts points for
bare `open(...)` without a corresponding `close()`.

## Read modes

The second argument to `open` controls the mode:

| Mode | Meaning |
|---|---|
| `"r"` (default) | read text |
| `"w"` | write text — **truncates the file first** |
| `"a"` | append text |
| `"r+"` | read + write |
| `"rb"`, `"wb"` | same modes, but bytes (used for images, binaries) |

```python
with open("notes.txt", "w") as f:
    f.write("first line\n")
    f.write("second line\n")
```

`"w"` quietly destroys the existing file's contents. If you want
"create-or-fail-if-exists", use `"x"`.

## Reading: line-by-line is the default

The cleanest read pattern iterates the file object. Each iteration
yields one line **including** its trailing newline:

```python
with open("words.txt") as f:
    for line in f:
        word = line.rstrip("\n")
        process(word)
```

`.rstrip("\n")` strips just newlines. Bare `.strip()` strips all
whitespace — usually fine, but be deliberate.

This streams the file one line at a time and uses constant memory. It
works for files that don't fit in RAM. CS 106A doesn't care about RAM,
but the pattern is universally useful.

## Reading: whole file or list of lines

Sometimes you want the whole thing at once:

```python
with open("words.txt") as f:
    content = f.read()           # one giant string
    print(len(content))

with open("words.txt") as f:
    lines = f.readlines()        # list of strings (each with trailing \n)

with open("words.txt") as f:
    lines = [line.rstrip() for line in f]  # idiomatic
```

The third form is the CS 106A favorite when you want a clean list.

## Writing

```python
with open("out.txt", "w") as f:
    f.write("Hello\n")
    f.write(f"Total: {total}\n")
    print("done", file=f)        # print() can target a file too
```

`f.write` does NOT append a newline; you have to include `"\n"` yourself.
`print` does add one. Mixing the two in the same file is fine but
confusing — pick one style per script.

## Encoding

Always specify `encoding="utf-8"` for text files. Without it Python uses
the platform default, which is `cp1252` on Windows and `utf-8` on
Mac/Linux. That difference causes "works on my machine" bugs:

```python
with open("words.txt", encoding="utf-8") as f:
    handle(f)
```

CS 106A's reference solutions all specify the encoding explicitly. Make
it a habit.

## Path conventions

Use `pathlib.Path` for anything beyond a bare filename:

```python
from pathlib import Path

DATA = Path(__file__).parent / "data"
words_path = DATA / "words.txt"

with words_path.open(encoding="utf-8") as f:
    handle(f)
```

`Path` handles cross-platform path separators, exists checks, and
extension manipulation. The `Path.open(...)` method is identical to
`open(path, ...)`.

## Common patterns

### Word count

```python
from collections import Counter
from pathlib import Path

def word_count(path: Path) -> Counter:
    with path.open(encoding="utf-8") as f:
        return Counter(word for line in f for word in line.split())
```

A genexp inside `Counter` keeps the file streaming. Two-clause
comprehension covered in {{GUIDE:cs106a-list-comprehensions}}.

### CSV (built-in module)

```python
import csv
from pathlib import Path

with Path("grades.csv").open(encoding="utf-8", newline="") as f:
    reader = csv.DictReader(f)
    for row in reader:
        print(row["name"], row["score"])
```

`newline=""` is required by the `csv` module to handle line endings
correctly across platforms. The `DictReader` form returns each row as a
dict keyed by header name — much easier than indexing.

### JSON

```python
import json
from pathlib import Path

with Path("config.json").open(encoding="utf-8") as f:
    config = json.load(f)

with Path("out.json").open("w", encoding="utf-8") as f:
    json.dump(config, f, indent=2)
```

`json.load` reads from a file object; `json.loads` (note the `s`) reads
from a string. Same for `dump` / `dumps`.

## Errors you'll meet

| Error | Cause | Fix |
|---|---|---|
| `FileNotFoundError` | wrong path | print `Path(...).resolve()` to debug |
| `PermissionError` | file is open elsewhere or read-only | close the other handle |
| `UnicodeDecodeError` | wrong encoding | specify `encoding="utf-8"` |
| `IsADirectoryError` | passed a directory to `open` | check `path.is_file()` first |

The pset setup readme `{{FILE:stanford-cs106a-pset-setup-readme}}` lays
out the expected directory layout for every assignment so the relative
paths just work.

## Try / except for missing files

When the user supplies a path you don't control, wrap the open:

```python
try:
    with open(path, encoding="utf-8") as f:
        process(f)
except FileNotFoundError:
    print(f"Could not open {path}")
```

CS 106A's "robust input" rubric rewards this pattern over crashing.

For introspecting what's in a file when something goes wrong, see
{{GUIDE:cs106a-debugging-techniques}}.
