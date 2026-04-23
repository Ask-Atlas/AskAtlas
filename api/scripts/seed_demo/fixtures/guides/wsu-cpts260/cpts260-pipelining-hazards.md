---
slug: cpts260-pipelining-hazards
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "260"
title: "Pipelining and Hazards — The Five-Stage MIPS Pipeline"
description: "How the classic MIPS pipeline overlaps instructions, and the data, control, and structural hazards that get in the way."
tags: ["pipelining", "hazards", "forwarding", "stalls", "midterm"]
author_role: bot
quiz_slug: cpts260-pipelining-hazards-quiz
attached_files:
  - wsu-cpts260-pipelining-hazards-worksheet
attached_resources: []
---

# Pipelining and Hazards

A non-pipelined MIPS implementation executes one instruction at a time
and wastes most of its silicon on every cycle. The classic five-stage
pipeline — IF, ID, EX, MEM, WB — overlaps five instructions in flight
at once, so each cycle a new instruction enters and a finished one
retires. **Throughput is one instruction per cycle in the steady state**,
even though each individual instruction still takes five cycles.

The lecture worksheet `{{FILE:wsu-cpts260-pipelining-hazards-worksheet}}`
has the empty pipeline diagrams; this guide is the explainer that goes
with them.

## The five stages

| Stage | What it does |
|---|---|
| **IF** — Instruction Fetch | Read instruction from I-cache at PC, increment PC |
| **ID** — Instruction Decode | Decode opcode, read register file, sign-extend immediate |
| **EX** — Execute | ALU op, branch comparison, address calculation |
| **MEM** — Memory Access | Load reads D-cache, store writes D-cache |
| **WB** — Write Back | Result written to the register file |

Each stage is separated by a **pipeline register** that latches the
intermediate values for the next cycle. You will see them named like
`IF/ID`, `ID/EX`, `EX/MEM`, `MEM/WB` on diagrams.

## A clean pipeline

Five independent instructions filling the pipeline:

```text
Cycle:   1   2   3   4   5   6   7   8
i1:      IF  ID  EX  MEM WB
i2:          IF  ID  EX  MEM WB
i3:              IF  ID  EX  MEM WB
i4:                  IF  ID  EX  MEM WB
i5:                      IF  ID  EX  MEM WB
```

By cycle 5 all five stages are busy. From there, one instruction
completes per cycle. CPI = 1 in steady state.

## Hazards

A hazard is anything that prevents an instruction from executing in the
next cycle. Three families:

### 1. Structural hazards

Two instructions need the same hardware resource in the same cycle. The
classical MIPS pipeline avoids most of these by having separate I-cache
and D-cache (Harvard split) and two register-file ports for parallel
reads. If you only had one cache, IF and MEM would conflict every cycle.

### 2. Data hazards

The next instruction needs a value the previous one has not yet written
back. Three subtypes (the names are NOT all hazards in MIPS, but you
should know them):

- **RAW (read after write)** — true data dependence. The pipeline must
  wait for the producer.
- **WAR (write after read)** — anti-dependence. Cannot occur in our
  in-order pipeline; appears in out-of-order designs.
- **WAW (write after write)** — output dependence. Same story.

The classic RAW example:

```mips
add  $t0, $t1, $t2     # writes $t0 in WB (cycle 5)
sub  $t3, $t0, $t4     # reads $t0 in ID (cycle 3 of its run)
```

Without help, `sub` reads the stale `$t0`. Two fixes:

#### Forwarding (bypassing)

Add wires from EX/MEM and MEM/WB pipeline registers back into the ALU
inputs. The result of `add` is available at the end of its EX stage —
exactly when `sub` enters EX. No stall needed for ALU-to-ALU.

#### Stalling (bubble insertion)

Forwarding cannot save a load-use hazard:

```mips
lw   $t0, 0($t1)       # value available end of MEM (cycle 4)
add  $t2, $t0, $t3     # needs $t0 in EX (cycle 4)
```

The load only knows the value at the *end* of MEM, but the dependent
ALU op needs it at the *start* of EX in the same cycle. Solution: insert
one bubble. The compiler usually fills the slot with an unrelated
instruction.

### 3. Control hazards

Branches resolve in EX (or earlier with optimisations), but by then
two later instructions have already entered the pipeline. Options:

- **Stall.** Freeze the front end until the branch resolves. Costs ~2
  cycles per branch. Death by a thousand cuts on real code.
- **Predict not-taken.** Speculatively continue down the fall-through
  path. Squash the speculatively-fetched ops on a misprediction.
- **Branch delay slot.** The classic MIPS hack — *always* execute the
  instruction after a branch, and let the compiler fill it usefully.
  Works at low pipeline depths; falls apart on deeper modern pipes.
- **Dynamic prediction.** A branch predictor table indexed by PC. The
  classic 2-bit saturating counter mispredicts at most twice per loop
  exit and was good enough to push average accuracy past 90% on real
  workloads — modern hybrid predictors push that into the high 90s.

## Drawing pipeline diagrams

The CPTS 260 midterm always asks you to draw a pipeline diagram. The
template:

```text
Instruction  | C1 | C2 | C3 | C4 | C5 | C6 | C7 | C8 |
add t0,t1,t2 | IF | ID | EX | MEM| WB |    |    |    |
sub t3,t0,t4 |    | IF | ID | EX | MEM| WB |    |    |
lw  t5, 0(t6)|    |    | IF | ID | EX | MEM| WB |    |
add t7,t5,t8 |    |    |    | IF | ID | ** | EX | MEM|   <- bubble
```

The `**` is a stall (bubble) inserted because of the load-use hazard
between `lw` and the dependent `add`. With proper forwarding from MEM/WB
the cycle count is N + 4 + (number of stalls).

## Speedup math

Ideal speedup of an N-stage pipeline is N (every cycle finishes one
instruction instead of every N). Real speedup is much less, because:

- Pipeline fill / drain costs N − 1 cycles at the start and end.
- Hazard stalls add bubbles.
- Cache misses spill in from {{GUIDE:cpts260-cache-design}}.

For long-running programs the fill/drain term is negligible, so:

```text
Speedup = N / (1 + stall_cycles_per_instruction)
```

A 5-stage pipeline with 0.5 stall cycles per instruction (one stall
every other instruction) gets a 3.3× speedup, not 5×. That gap is what
fancy techniques like out-of-order issue and branch prediction try to
close.

## Practice

After you have done the worksheet, take
{{QUIZ:cpts260-pipelining-hazards-quiz}}. The questions on load-use
stalls and forwarding paths are the most common midterm pattern.

For the related ISA cheat sheet that names every register the
forwarding network has to track, see {{GUIDE:cpts260-mips-cheatsheet}}.
For the broader course, see {{COURSE:wsu/cpts260}}.
