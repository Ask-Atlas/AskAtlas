---
slug: cpts260-instruction-encoding
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "260"
title: "Instruction Encoding — R, I, and J Formats in MIPS"
description: "How a 32-bit MIPS instruction is laid out, with worked examples for each format."
tags: ["mips", "encoding", "isa", "binary", "machine-code"]
author_role: bot
quiz_slug: cpts260-instruction-encoding-quiz
attached_files:
  - wsu-cpts260-instruction-encoding-slides
attached_resources: []
---

# Instruction Encoding in MIPS

Every MIPS instruction is exactly 32 bits. That is the entire reason MIPS
exists — fixed-width encoding makes the fetch stage trivial and lets the
decode stage extract fields with pure wiring instead of state machines.
There are only three formats: **R**, **I**, and **J**. Once you can hand-
encode one of each, the midterm encoding question writes itself.

The slide deck `{{FILE:wsu-cpts260-instruction-encoding-slides}}` has the
full opcode and `funct` tables; this guide walks through how to use them.

## R-format (register)

Used for arithmetic and logical operations between three registers.

```text
| 31  26 | 25  21 | 20  16 | 15  11 | 10   6 | 5    0 |
|  op    |  rs    |  rt    |  rd    |  shamt |  funct |
|  6 b   |  5 b   |  5 b   |  5 b   |  5 b   |  6 b   |
```

For all R-type instructions `op = 000000`. The actual operation is
selected by `funct`. `shamt` is the shift amount (used only by `sll`,
`srl`, `sra`); for everything else it must be `00000`.

### Worked example: `add $t0, $t1, $t2`

| Field | Value | Bits |
|---|---|---|
| op | 0 | `000000` |
| rs | `$t1` (9) | `01001` |
| rt | `$t2` (10) | `01010` |
| rd | `$t0` (8) | `01000` |
| shamt | 0 | `00000` |
| funct | 32 | `100000` |

Concatenated: `000000 01001 01010 01000 00000 100000`
= `0x012A4020`.

The order in the encoding (`rs`, `rt`, `rd`) is **not** the source-order
of the assembly (`rd`, `rs`, `rt`). This catches everyone the first time.

## I-format (immediate)

Used for ALU-with-immediate, loads, stores, and conditional branches.

```text
| 31  26 | 25  21 | 20  16 | 15           0 |
|  op    |  rs    |  rt    |  immediate     |
|  6 b   |  5 b   |  5 b   |  16 b (signed) |
```

The immediate is sign-extended to 32 bits before use, except for
`andi`/`ori`/`xori`, which zero-extend.

### Worked example: `addi $t0, $t1, -1`

| Field | Value | Bits |
|---|---|---|
| op | 8 | `001000` |
| rs | `$t1` (9) | `01001` |
| rt | `$t0` (8) | `01000` |
| immediate | -1 | `1111 1111 1111 1111` |

= `0x2128FFFF`.

### Loads and stores

For `lw $t0, 12($sp)`:

| Field | Value | Bits |
|---|---|---|
| op | 35 | `100011` |
| rs | `$sp` (29) | `11101` |
| rt | `$t0` (8) | `01000` |
| immediate | 12 | `0000 0000 0000 1100` |

The "destination" register goes in `rt`, and the base register goes in
`rs`. Stores use the same layout but `rt` is the *source*.

### Branches

For `beq $t0, $t1, target` the immediate field holds the *PC-relative
word offset* — that is, `(target - PC_after_branch) >> 2`. The hardware
multiplies it by 4 (shifts left 2) and adds it to PC + 4.

If `beq` is at address `0x00400020` and `target` is `0x00400030`, the
offset is `(0x00400030 - 0x00400024) / 4 = 3`.

## J-format (jump)

Used by `j` and `jal` only.

```text
| 31  26 | 25                            0 |
|  op    |  target (26 bits)              |
```

The 26-bit field is shifted left 2 to produce a 28-bit byte offset, and
the top 4 bits of PC (the upper region of the address space) are
concatenated on the front. So a single `j` cannot leave the current
256 MB region; that is what `jr` is for.

### Worked example: `j 0x0040000C`

The target field holds bits 27..2 of the destination:

`0x0040000C >> 2 = 0x00100003`

| Field | Value | Bits |
|---|---|---|
| op | 2 | `000010` |
| target | `0x00100003` | `00 0001 0000 0000 0000 0000 0000 11` |

= `0x08100003`.

## Why this layout?

The MIPS designers cared deeply about the decode stage. Notice that:

1. `op` is always in the same bits regardless of format.
2. `rs` and `rt` are always in the same bits for R and I formats — so
   the register file can begin reading them in parallel with decode.
3. R-type `funct` is the *last* field, decoded only after we know it is
   an R-type — which is fine, because the ALU does not need it until
   the EX stage.

This matters when we look at the pipeline in
{{GUIDE:cpts260-pipelining-hazards}}. Fixed-position fields are why MIPS
can fetch + decode in a single cycle each.

## Disassembly drill

Given the word `0x8DAC0014`, what instruction is it?

1. Top 6 bits: `100011` → opcode 35 → `lw` (I-type).
2. Next 5: `01101` → `rs = 13` → `$t5`.
3. Next 5: `01100` → `rt = 12` → `$t4`.
4. Bottom 16: `0000 0000 0001 0100` → 20.

So: `lw $t4, 20($t5)`.

Practice this on the worksheets in
`{{FILE:wsu-cpts260-instruction-encoding-slides}}` until you can do it
without the opcode table. Then try the
{{QUIZ:cpts260-instruction-encoding-quiz}}.

## Pitfalls

- **Branch offset units.** Always think of the immediate as words, not
  bytes. The hardware shifts left 2 for you.
- **Sign vs zero extension.** `addi` sign-extends, `andi` zero-extends.
  Confusing the two on signed/unsigned arithmetic problems is the most
  common encoding-question mistake.
- **R-type field order.** The encoding is `rs`, `rt`, `rd`; the assembly
  is `rd`, `rs`, `rt`. Translate slowly.

For the broader ISA cheat sheet, see {{GUIDE:cpts260-mips-cheatsheet}}.
