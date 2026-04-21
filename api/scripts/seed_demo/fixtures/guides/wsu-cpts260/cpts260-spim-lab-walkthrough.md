---
slug: cpts260-spim-lab-walkthrough
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "260"
title: "SPIM Lab Walkthrough — From Source to Single-Step"
description: "Step-by-step setup for SPIM/QtSPIM, loading a program, and stepping through it instruction by instruction."
tags: ["spim", "lab", "tools", "debugging", "mips"]
author_role: bot
attached_files:
  - wsu-cpts260-lab-setup-spim
attached_resources: []
---

# SPIM Lab Walkthrough

This guide walks you through running a real MIPS program on the SPIM
simulator from a clean machine. Once you can do this end-to-end you can
do every CPTS 260 lab. The course's official setup notes are in
`{{FILE:wsu-cpts260-lab-setup-spim}}`; this guide is the friendlier
narrated version.

## What SPIM is

SPIM (the *S*imulator for the *MIPS* ISA) is a software MIPS32
emulator. It does not pretend to be cycle-accurate — there are no
caches, no pipeline, no TLB. What you get is a clean view of:

- The 32 GP registers + HI/LO + PC
- The CP0 system registers
- The text and data segments
- The kernel text segment for exception handlers

There are two interfaces:

- **`spim`** — the command-line REPL. Fast, scriptable.
- **`QtSPIM`** — a graphical front end with register, memory, and code
  panes. What we use in lab.

## Setup

### macOS / Linux

```bash
# macOS
brew install spim

# Debian/Ubuntu
sudo apt install spim qtspim
```

### Windows (WSL or native)

Easiest path: install WSL (Ubuntu) and follow the Linux instructions.
A native Windows build of QtSPIM exists but is finicky to keep current.

### Verify

```bash
spim -version
# SPIM Version 9.x ...
```

## Anatomy of a SPIM program

A SPIM source file mixes assembler directives with MIPS instructions.
The two segments you always see are `.data` (initialised data) and
`.text` (instructions).

```mips
        .data
prompt: .asciiz  "Enter a number: "
result: .asciiz  "You doubled it: "
nl:     .asciiz  "\n"

        .text
        .globl   main
main:
        # print prompt
        li       $v0, 4
        la       $a0, prompt
        syscall

        # read int into $v0
        li       $v0, 5
        syscall
        move     $t0, $v0          # save the input

        # double it
        sll      $t0, $t0, 1

        # print result
        li       $v0, 4
        la       $a0, result
        syscall
        li       $v0, 1
        move     $a0, $t0
        syscall
        li       $v0, 4
        la       $a0, nl
        syscall

        # exit cleanly
        li       $v0, 10
        syscall
```

Save this as `double.s`.

## Running it from the CLI

```bash
spim -file double.s
```

SPIM prints the standard banner, then runs your program. Type a
number when prompted; you should see it doubled. Press `Ctrl-D` to
quit.

## Running it in QtSPIM

1. Launch `qtspim`.
2. **File → Reinitialize Simulator** (always; clears state from any
   previous program).
3. **File → Load File…** and pick `double.s`.
4. The Text Segment pane now shows your assembled instructions
   alongside their addresses and machine-code encoding. The Data
   Segment pane shows your strings.
5. **Simulator → Run** to execute, or **Simulator → Single Step**
   (F10) to advance one instruction.

Watch the Registers pane while you single-step — the changes are
immediate and obvious. This is *the* most efficient way to internalise
how the ABI uses `$v0`/`$a0`.

## Pseudo-instructions in the wild

Single-step through `la $a0, prompt` and watch what actually happens.
You will see two real instructions execute:

```mips
lui  $1, 0x1001       # $1 = upper 16 bits of prompt's address
ori  $a0, $1, 0x0000  # OR in the lower 16 bits
```

That is the assembler expanding `la` into the `lui`/`ori` pair we
discussed in {{GUIDE:cpts260-mips-cheatsheet}}. The use of `$at`
(register 1) for the scratch space is exactly why the convention
reserves it.

## Breakpoints

In QtSPIM right-click any instruction and choose "Set Breakpoint".
Then **Run** — execution stops there and you can inspect state.

In CLI SPIM:

```text
(spim) break 0x00400024
(spim) run
```

To list active breakpoints: `breakpoints`. To clear one: `delete <addr>`.

## Common gotchas

- **Forgetting `.globl main`.** Without it SPIM falls back to the
  built-in startup that calls `main`, but errors with "main not
  defined" if you forgot to mark the symbol global.
- **Writing to `.data` strings.** They are read-only in QtSPIM.
  `sb $t0, 0($a0)` against an `.asciiz` literal raises an exception.
- **Branch delay slots.** Off by default in QtSPIM; turn them on under
  **Simulator → Settings** when you want to match real hardware
  behaviour. The pipeline diagrams in
  {{GUIDE:cpts260-pipelining-hazards}} assume branch delay slots are
  present.
- **Bare-decimal addresses.** SPIM expects hex with `0x`. Typing
  `0x10010000` is what you want; `10010000` is decimal and almost
  certainly outside any real segment.

## A debugging workflow

When a program does not work, the routine is:

1. Single-step until something looks wrong.
2. Check the registers pane — what value did the ALU produce, vs what
   you expected?
3. If a memory access misbehaved, look at the data segment at the
   address the load/store used.
4. If it crashed, check `Cause` in the CP0 registers (see
   {{GUIDE:cpts260-interrupts-and-exceptions}}). The ExcCode field
   tells you whether it was a bad address, an overflow, or something
   else.

This loop will get you through every lab in the course.

## Practice

Once SPIM is working, try writing a tiny program that:

- Reads two integers,
- Computes their GCD with the Euclidean algorithm,
- Prints the result.

The arithmetic is two instructions per loop iteration; the I/O is
mechanical syscall pairs. If you get stuck, revisit
{{GUIDE:cpts260-mips-cheatsheet}} for the syscall codes. For the
big-picture review heading into the midterm, see
{{GUIDE:cpts260-midterm-review}}.
