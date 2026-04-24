---
slug: wsu-cpts260-instruction-encoding-slides
title: "MIPS Instruction Encoding"
mime: application/vnd.openxmlformats-officedocument.presentationml.presentation
filename: instruction-encoding-slides.pptx
course: wsu/cpts260
description: "Slide deck walking through R-type, I-type, and J-type encoding, field layouts, and opcode decoding."
author_role: bot
---

## Why Fixed-Width Encoding?

Every MIPS instruction is exactly 32 bits. Decode logic is simpler, fetch is aligned, branch targets are predictable. The trade-off is code density — CISC architectures like x86 pack more work into fewer bytes.

## Three Instruction Formats

MIPS uses exactly three formats, distinguished by the top 6 bits (opcode). Everything else is a field layout negotiation.

## R-type Layout

Register-only instructions: `add`, `sub`, `and`, `or`, `sll`, `jr`, and so on.

```
| opcode(6) | rs(5) | rt(5) | rd(5) | shamt(5) | funct(6) |
```

All R-type instructions share `opcode = 0x00`; the `funct` field selects the actual ALU operation.

## R-type Example: add $t0, $t1, $t2

- `opcode = 000000`
- `rs     = 01001` (t1 = 9)
- `rt     = 01010` (t2 = 10)
- `rd     = 01000` (t0 = 8)
- `shamt  = 00000`
- `funct  = 100000` (ADD)

Full word: `0x012A4020`.

## I-type Layout

Immediate, loads, stores, branches: `addi`, `lw`, `sw`, `beq`, `lui`.

```
| opcode(6) | rs(5) | rt(5) | immediate(16) |
```

The 16-bit immediate is sign-extended for arithmetic and branches, zero-extended for logical ops.

## I-type Example: lw $t0, 4($sp)

- `opcode = 100011` (LW)
- `rs     = 11101` (sp = 29)
- `rt     = 01000` (t0 = 8)
- `imm    = 0x0004`

## Branches Are PC-Relative

```
target = (PC + 4) + (sign_extend(imm) << 2)
```

Shifting by 2 gives a 128 KB branch range. The `+ 4` is because PC has already incremented by the time the branch executes.

## J-type Layout

Unconditional jumps: `j`, `jal`.

```
| opcode(6) | address(26) |
```

## Pseudo-Direct Addressing

```
target = { PC[31:28], address << 2 }
```

The low 2 bits are zero (word-aligned), and the top 4 bits come from the current PC — so `j` cannot cross a 256 MB region.

## Decoding in Hardware

A 6-bit opcode decoder steers the instruction into the right path. `opcode == 0` routes to the funct-based R-type decoder. `opcode == 2 || opcode == 3` selects J-type. Everything else is I-type.

## Sign-Extension Matters

`addiu $t0, $t1, 0xFFFF` does **not** add 65535 — the immediate is sign-extended to `0xFFFFFFFF`, so it subtracts 1. Watch this on exam questions.

## Assembler Pseudo-instructions

The assembler expands friendly syntax into real encodings:

- `li $t0, 0x12345678` → `lui $t0, 0x1234` + `ori $t0, $t0, 0x5678`
- `move $t0, $t1`       → `add $t0, $zero, $t1`
- `blt $t0, $t1, L`     → `slt $at, $t0, $t1` + `bne $at, $zero, L`

## Practice

Hand-encode these into hex. Bring your work to lab.

1. `sub $s0, $s1, $s2`
2. `sw  $t0, -8($fp)`
3. `j   0x00400100`
