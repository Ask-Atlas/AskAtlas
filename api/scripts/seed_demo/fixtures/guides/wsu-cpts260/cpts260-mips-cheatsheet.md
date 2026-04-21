---
slug: cpts260-mips-cheatsheet
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "260"
title: "MIPS Assembly Cheatsheet — CPTS 260"
description: "Core MIPS instructions, register conventions, and addressing modes you need on the CPTS 260 midterm."
tags: ["mips", "assembly", "registers", "isa", "midterm"]
author_role: bot
quiz_slug: cpts260-mips-cheatsheet-quiz
attached_files:
  - wsu-cpts260-mips-cheatsheet
attached_resources: []
---

# MIPS Assembly Cheatsheet

CPTS 260 is built around MIPS for one reason: it is the smallest realistic
ISA you can fit in your head. Once you can read MIPS fluently, every other
load/store architecture (RISC-V, ARMv8, even chunks of x86) feels familiar.
This cheatsheet collects the parts you will actually be tested on.

## Register file

MIPS has 32 general-purpose integer registers. The hardware does not care
which is which — the calling convention does. Memorise the convention; the
compiler, assembler, and graders all assume it.

| Register | Number | Purpose |
|---|---|---|
| `$zero` | 0 | hard-wired zero, writes are silently dropped |
| `$at` | 1 | reserved for the assembler (pseudo-instruction expansion) |
| `$v0`, `$v1` | 2–3 | function return values, syscall code |
| `$a0`–`$a3` | 4–7 | first four function arguments |
| `$t0`–`$t7` | 8–15 | caller-saved temporaries |
| `$s0`–`$s7` | 16–23 | callee-saved (function must preserve) |
| `$t8`, `$t9` | 24–25 | more caller-saved temporaries |
| `$k0`, `$k1` | 26–27 | reserved for the OS kernel — do not touch |
| `$gp` | 28 | global pointer |
| `$sp` | 29 | stack pointer |
| `$fp` | 30 | frame pointer (often unused) |
| `$ra` | 31 | return address (set by `jal`) |

The single most important rule: `$zero` always reads as 0, and writes to it
are discarded. That is why `add $t0, $zero, $a0` is the canonical "move"
in MIPS — there is no real `mov` instruction.

## Instruction families

MIPS has three encoding formats and only a handful of instructions per
family. The full table is in `{{FILE:wsu-cpts260-mips-cheatsheet}}`; the
common ones are below.

### Arithmetic and logical (R-type)

```mips
add  $t0, $t1, $t2   # $t0 = $t1 + $t2  (traps on overflow)
addu $t0, $t1, $t2   # unsigned, no trap
sub  $t0, $t1, $t2
and  $t0, $t1, $t2
or   $t0, $t1, $t2
xor  $t0, $t1, $t2
sll  $t0, $t1, 4     # logical left shift by 4
srl  $t0, $t1, 4     # logical right shift
sra  $t0, $t1, 4     # arithmetic right shift (sign-extends)
slt  $t0, $t1, $t2   # set $t0 = 1 if $t1 < $t2 else 0
```

### Immediates (I-type)

```mips
addi  $t0, $t1, 7      # signed 16-bit immediate
addiu $t0, $t1, 7      # unsigned (no overflow trap; still sign-extended!)
andi  $t0, $t1, 0xFF   # zero-extends the immediate
ori   $t0, $t1, 0x0F
slti  $t0, $t1, 100
lui   $t0, 0xDEAD      # load upper 16 bits, lower 16 are 0
```

The `addiu`/sign-extension subtlety is a classic exam trick. The "u" stands
for "no overflow trap" — the immediate is *still* sign-extended.

### Memory (load/store)

MIPS is a load/store architecture: arithmetic only operates on registers.
All memory access uses base + 16-bit signed offset:

```mips
lw  $t0, 0($sp)      # load word
lh  $t0, 2($sp)      # load halfword (sign-extended)
lhu $t0, 2($sp)      # load halfword unsigned (zero-extended)
lb  $t0, 1($sp)      # load byte (sign-extended)
lbu $t0, 1($sp)      # load byte unsigned

sw  $t0, 0($sp)
sh  $t0, 2($sp)
sb  $t0, 1($sp)
```

Word loads must be 4-byte aligned; halfwords must be 2-byte aligned. An
unaligned load raises an address-error exception (covered in
{{GUIDE:cpts260-interrupts-and-exceptions}}).

### Branches and jumps

```mips
beq  $t0, $t1, label    # branch if equal
bne  $t0, $t1, label    # branch if not equal
blez $t0, label         # branch if <= 0
bgtz $t0, label
j    label              # unconditional jump
jal  label              # jump and link — saves return addr in $ra
jr   $ra                # jump register (function return)
```

Branches use a PC-relative offset (16 bits, shifted left 2). Jumps use a
26-bit absolute target within the current 256 MB region. Note the **branch
delay slot**: the instruction immediately after a branch *always executes*.
SPIM hides this by default; real hardware does not.

## Function call ABI

A typical prologue/epilogue saves callee-saved registers and `$ra` if the
function calls anyone else (a "leaf" function may skip the save):

```mips
my_func:
    addiu $sp, $sp, -8       # carve 2 words of stack
    sw    $ra, 4($sp)        # save return address
    sw    $s0, 0($sp)        # save callee-saved register

    # ... body uses $s0 freely ...

    lw    $s0, 0($sp)        # restore
    lw    $ra, 4($sp)
    addiu $sp, $sp, 8        # release frame
    jr    $ra                # return
```

Caller passes arguments in `$a0`–`$a3` (extras spill onto the stack), and
expects results in `$v0`/`$v1`.

## Pseudo-instructions you'll actually type

The assembler expands these to one or two real instructions. They are not
ISA but they are convenient.

| Pseudo | Expands to |
|---|---|
| `move $t0, $t1` | `add $t0, $zero, $t1` |
| `li $t0, 0xCAFEBABE` | `lui $at, 0xCAFE; ori $t0, $at, 0xBABE` |
| `la $t0, label` | same `lui`/`ori` pair against the label address |
| `bge $t0, $t1, lbl` | `slt $at, $t0, $t1; beq $at, $zero, lbl` |

Notice why `$at` exists — the assembler needs a scratch register to expand
these.

## SPIM-specific details

The SPIM simulator we use in lab implements the syscalls below via
`syscall` with a code in `$v0`:

| Code | Action |
|---|---|
| 1 | print integer in `$a0` |
| 4 | print null-terminated string at address `$a0` |
| 5 | read integer into `$v0` |
| 8 | read string (`$a0` = buffer, `$a1` = length) |
| 10 | exit |

A "Hello, world" in SPIM is therefore:

```mips
        .data
hello:  .asciiz "Hello, CPTS 260!\n"

        .text
main:   la   $a0, hello
        li   $v0, 4
        syscall
        li   $v0, 10
        syscall
```

## Practice

Once you can write a function call frame from memory and explain the
difference between `addi` and `addiu`, take the
{{QUIZ:cpts260-mips-cheatsheet-quiz}}. After that, jump into the encoding
details in {{GUIDE:cpts260-instruction-encoding}} or skip ahead to the
{{GUIDE:cpts260-spim-lab-walkthrough}} if you want to run something
end-to-end first.

For the rest of the course catalog, see {{COURSE:wsu/cpts260}}.
