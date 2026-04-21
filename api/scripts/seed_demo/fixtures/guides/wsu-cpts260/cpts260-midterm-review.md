---
slug: cpts260-midterm-review
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "260"
title: "Midterm Review — CPTS 260"
description: "A consolidated review of every topic that has appeared on past CPTS 260 midterms, organised by question type."
tags: ["midterm", "review", "consolidation", "exam-prep"]
author_role: bot
attached_files:
  - wsu-cpts260-mips-cheatsheet
  - wsu-cpts260-pipelining-hazards-worksheet
attached_resources: []
---

# CPTS 260 Midterm Review

The CPTS 260 midterm reuses the same five-or-six question shapes every
semester. If you can do each shape from a blank page, you have already
covered 90% of the test. This guide walks through them in the order
they tend to appear, with pointers to the deeper guides for each.

The two cheat sheets `{{FILE:wsu-cpts260-mips-cheatsheet}}` and
`{{FILE:wsu-cpts260-pipelining-hazards-worksheet}}` are the printed
materials students bring in (the exam is closed-book, but you can
recreate them mentally).

## Question 1 — assemble or disassemble

You are given a snippet of MIPS and asked to encode it (or vice versa).

**Drill:** encode `addi $s0, $s0, -4`.

| Field | Value | Bits |
|---|---|---|
| op | 8 | `001000` |
| rs | `$s0` (16) | `10000` |
| rt | `$s0` (16) | `10000` |
| imm | -4 | `1111 1111 1111 1100` |

Concatenated: `0010 0010 0001 0000 1111 1111 1111 1100`
= `0x2210FFFC`.

**Drill:** disassemble `0x00432020`.

- op = 0 → R-type
- rs = `$v0` (2)
- rt = `$v1` (3)
- rd = `$a0` (4)
- shamt = 0
- funct = `100000` = 32 → `add`

So: `add $a0, $v0, $v1`.

For the field layouts and full opcode table, see
{{GUIDE:cpts260-instruction-encoding}}.

## Question 2 — translate C to MIPS

Given a C function, write equivalent MIPS, including the prologue and
epilogue.

```c
int sum_to(int n) {
    int s = 0;
    for (int i = 1; i <= n; i++) {
        s += i;
    }
    return s;
}
```

Becomes:

```mips
sum_to:
        # n in $a0; result in $v0
        li      $v0, 0          # s = 0
        li      $t0, 1          # i = 1
loop:
        bgt     $t0, $a0, done  # if i > n, exit
        add     $v0, $v0, $t0   # s += i
        addi    $t0, $t0, 1     # i++
        j       loop
done:
        jr      $ra
```

Notes for the grader:

- This is a leaf function — no `$s` registers used, no `$ra` save.
- `bgt` is a pseudo-instruction; it expands to `slt`/`bne`.
- Calling convention: argument in `$a0`, return in `$v0`.

For the full register convention table see
{{GUIDE:cpts260-mips-cheatsheet}}.

## Question 3 — pipeline diagram

Draw the cycle-by-cycle execution with hazards and forwarding.

```mips
lw   $t0, 0($s0)
add  $t1, $t0, $t2     # load-use, needs a stall
sub  $t3, $t1, $t4     # depends on prev, forwarding works
sw   $t3, 0($s1)
```

```text
              | C1 | C2 | C3 | C4 | C5 | C6 | C7 | C8 |
lw  t0,0(s0)  | IF | ID | EX | MEM| WB |    |    |    |
add t1,t0,t2  |    | IF | ID | ** | EX | MEM| WB |    |   <- stall
sub t3,t1,t4  |    |    | IF | ** | ID | EX | MEM| WB |
sw  t3,0(s1)  |    |    |    |    | IF | ID | EX | MEM|
```

The `**` is the load-use bubble; everything else gets resolved by
EX→EX or MEM→EX forwarding. For more pipeline drills see
{{GUIDE:cpts260-pipelining-hazards}}.

## Question 4 — cache hit/miss trace

Given a cache geometry and an access stream, fill in the hit/miss
column and compute the hit rate.

The two pieces you need are:

1. The address decomposition (offset / index / tag bits).
2. The replacement policy (almost always LRU).

Worked traces, plus the formula for AMAT, live in
{{GUIDE:cpts260-cache-design}} and {{GUIDE:cpts260-memory-hierarchy}}.

A typical exam stream:

`0x00, 0x40, 0x80, 0x00, 0xC0, 0x40` against a direct-mapped 64-byte
cache with 16-byte lines. Compute it once and you have done it forever.

## Question 5 — performance arithmetic

Two flavours:

### CPI / speedup

```text
CPI = sum_i (frequency_i × cycles_i)
Speedup = CPU_time_old / CPU_time_new
       = (IC × CPI_old × clock_old) / (IC × CPI_new × clock_new)
```

If 25% of instructions are loads with CPI 2 and 75% are ALU ops with
CPI 1, average CPI = 0.25 × 2 + 0.75 × 1 = **1.25**.

### Amdahl's Law

```text
Speedup = 1 / ((1 - f) + f / s)
```

where `f` is the fraction of time spent on the part you sped up, and
`s` is the speedup of that part. The textbook example: if 60% of
runtime is parallelisable and you parallelise it 8×, speedup is
`1 / (0.4 + 0.075) = ~2.1×`. Not 8×. Amdahl is brutal.

## Question 6 — short-answer concepts

Examples that have appeared:

- Why does MIPS have a branch delay slot?
- Why is L1 cache 4-way associative instead of 16-way?
- What is the difference between an interrupt and an exception?
- What is the difference between `addi` and `addiu`?
- Why are `$k0` and `$k1` reserved?

For the last three, see {{GUIDE:cpts260-interrupts-and-exceptions}},
{{GUIDE:cpts260-mips-cheatsheet}}, and the kernel-register convention
in the same guide.

## Study plan (one week out)

| Day | Focus |
|---|---|
| Mon | Re-read {{GUIDE:cpts260-mips-cheatsheet}} cover to cover; rewrite the calling convention from memory. |
| Tue | Hand-encode/decode 5 instructions of each format. |
| Wed | Pipeline-diagram drill: 3 sequences, including loads + branches. |
| Thu | Cache trace + AMAT problems from `{{FILE:wsu-cpts260-pipelining-hazards-worksheet}}`. |
| Fri | Translate two C functions into MIPS, including the prologue. |
| Sat | Take every quiz: {{QUIZ:cpts260-mips-cheatsheet-quiz}}, {{QUIZ:cpts260-instruction-encoding-quiz}}, {{QUIZ:cpts260-cache-design-quiz}}, {{QUIZ:cpts260-pipelining-hazards-quiz}}. |
| Sun | Re-read whichever guide you scored worst on. |

## Things students underestimate

- **Hand arithmetic.** The exam is timed and there is no calculator.
  Practice adding and shifting binary by hand.
- **Reading machine code.** "Disassemble this" is the question students
  freeze on. Make it muscle memory.
- **Off-by-one in branch offsets.** Always count from PC + 4, not PC.

For the broader course map and links to every other resource, see
{{COURSE:wsu/cpts260}}.
