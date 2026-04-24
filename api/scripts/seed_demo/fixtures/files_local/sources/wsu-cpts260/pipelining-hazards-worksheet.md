---
slug: wsu-cpts260-pipelining-hazards-worksheet
title: "Pipelining Hazards Worksheet"
mime: application/vnd.openxmlformats-officedocument.wordprocessingml.document
filename: pipelining-hazards-worksheet.docx
course: wsu/cpts260
description: "Fill-in-the-blank worksheet covering data, control, and structural hazards in the classic 5-stage MIPS pipeline."
author_role: bot
---

## Instructions

Work through each section. Write answers in the provided blanks. Show hazard-detection reasoning explicitly; partial credit is awarded for correct stall counts even when the forwarding path is misidentified.

## Part 1 — The Classic 5-Stage Pipeline

Name the five stages in order:

1. IF — ______________________________
2. ID — ______________________________
3. EX — ______________________________
4. MEM — _____________________________
5. WB — ______________________________

An instruction issued at cycle `t` finishes WB at cycle ______.

## Part 2 — Data Hazards

There are three data-hazard categories. Fill in:

- **RAW** (________ after ________): the true dependency. A consumer reads a register before the producer has written it.
- **WAR** (________ after ________): only a hazard in out-of-order pipelines.
- **WAW** (________ after ________): also only a hazard out-of-order.

### Worked example

```asm
add  $t0, $t1, $t2     # I1
sub  $t3, $t0, $t4     # I2
```

Without forwarding, how many stalls does I2 incur? ______

With full EX-to-EX forwarding, how many stalls does I2 incur? ______

### Load-use delay

```asm
lw   $t0, 0($a0)       # I1
add  $t3, $t0, $t4     # I2
```

Even with forwarding, I2 must stall for ______ cycle(s). Explain why: ________________________________________________________________

## Part 3 — Control Hazards

Branches resolve in stage ______ of the classic pipeline. That means ______ instruction(s) after a taken branch enter the pipeline speculatively.

Name three mitigation techniques:

1. ____________________________________
2. ____________________________________
3. ____________________________________

A **branch delay slot** is the instruction slot immediately after a branch, which always executes. True / False? ______

## Part 4 — Structural Hazards

A structural hazard exists when two instructions need the same ______ in the same cycle. In the classic MIPS design, the memory port conflict is solved by splitting the cache into ________ and ________.

List two other structural hazards and the architectural fix for each:

| Hazard | Fix |
|--------|-----|
| ________________________ | ________________________ |
| ________________________ | ________________________ |

## Part 5 — Diagram

Draw the pipeline diagram for the following sequence, assuming forwarding and 1-cycle branch delay. Mark every stall with `**`.

```asm
lw   $t0, 0($a0)
add  $t1, $t0, $t2
beq  $t1, $zero, done
sub  $t3, $t4, $t5
```

Number of cycles to complete: ______

## Submission

Staple and hand in at the start of lab. Name: ______________________ Section: ______
