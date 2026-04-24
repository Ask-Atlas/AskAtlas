---
slug: wsu-cpts260-mips-cheatsheet
title: "MIPS Assembly Cheat Sheet"
mime: application/pdf
filename: mips-cheatsheet.pdf
course: wsu/cpts260
description: "Quick reference for MIPS32 register conventions, addressing modes, and the most common R/I/J instructions."
author_role: bot
---

## Register Conventions

MIPS32 defines 32 general-purpose registers, each 32 bits wide. The calling convention is not enforced by hardware — discipline matters.

| Number | Name    | Purpose                          | Preserved across call? |
|--------|---------|----------------------------------|------------------------|
| $0     | $zero   | Hard-wired zero                  | n/a                    |
| $1     | $at     | Assembler temporary              | no                     |
| $2–$3  | $v0–$v1 | Return values                    | no                     |
| $4–$7  | $a0–$a3 | Argument registers               | no                     |
| $8–$15 | $t0–$t7 | Temporaries                      | no                     |
| $16–$23| $s0–$s7 | Saved registers                  | yes (callee-saved)     |
| $24–$25| $t8–$t9 | More temporaries                 | no                     |
| $28    | $gp     | Global pointer                   | yes                    |
| $29    | $sp     | Stack pointer                    | yes                    |
| $30    | $fp     | Frame pointer                    | yes                    |
| $31    | $ra     | Return address                   | no (set by `jal`)      |

Rule of thumb: if you need a value to survive a `jal`, put it in `$s0`–`$s7` (and save/restore those on the stack in your prologue).

## Addressing Modes

MIPS is a load/store architecture, so only `lw`/`sw` touch memory. The five classical modes:

1. **Register** — operand is a register: `add $t0, $t1, $t2`
2. **Immediate** — 16-bit signed constant in the instruction: `addi $t0, $t1, 42`
3. **Base + displacement** — the only memory mode: `lw $t0, 4($sp)`
4. **PC-relative** — conditional branches: `beq $t0, $t1, label`
5. **Pseudo-direct** — `j`/`jal` concatenate the top 4 PC bits with a 26-bit index

## Common Instructions

```asm
# Arithmetic
add  $t0, $t1, $t2      # $t0 = $t1 + $t2
addi $t0, $t1, -8       # $t0 = $t1 + (-8)
sub  $t0, $t1, $t2
mul  $t0, $t1, $t2      # pseudo-instruction on MIPS32

# Memory
lw   $t0, 0($sp)        # load word
sw   $t0, 4($sp)        # store word
lb   $t0, 0($a0)        # load byte, sign-extended
lbu  $t0, 0($a0)        # load byte, zero-extended

# Control flow
beq  $t0, $t1, equal
bne  $t0, $zero, loop
j    target
jal  printf             # $ra = PC + 4; jump
jr   $ra                # return
```

## Function Prologue / Epilogue

```asm
myfunc:
    addi $sp, $sp, -8
    sw   $ra, 4($sp)
    sw   $s0, 0($sp)
    # ... body uses $s0 freely ...
    lw   $s0, 0($sp)
    lw   $ra, 4($sp)
    addi $sp, $sp, 8
    jr   $ra
```

Keep the stack 8-byte aligned on entry. Always save `$ra` before any nested `jal`.
