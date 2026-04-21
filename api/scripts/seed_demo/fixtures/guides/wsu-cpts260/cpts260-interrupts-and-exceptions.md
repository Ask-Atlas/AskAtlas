---
slug: cpts260-interrupts-and-exceptions
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "260"
title: "Interrupts and Exceptions in MIPS"
description: "How MIPS handles synchronous exceptions and asynchronous interrupts via Coprocessor 0."
tags: ["interrupts", "exceptions", "cp0", "kernel", "mips"]
author_role: bot
attached_files: []
attached_resources: []
---

# Interrupts and Exceptions

The pipeline we drew in {{GUIDE:cpts260-pipelining-hazards}} assumed a
clean, well-behaved instruction stream. Real systems are full of
discontinuities: a page fault while loading, a divide-by-zero, a
keyboard interrupt mid-loop, a `syscall` deliberately diving into the
kernel. MIPS handles all of these through a single uniform mechanism
called the **exception path**, and most of the bookkeeping lives in
**Coprocessor 0 (CP0)**.

## Vocabulary

The terminology trips students up; nail it once:

- **Exception** — synchronous, caused by the *current* instruction.
  Examples: arithmetic overflow, unaligned load, undefined opcode,
  `syscall`, `break`, page fault.
- **Interrupt** — asynchronous, caused by *external* hardware (timer,
  disk, NIC). The current instruction is innocent.
- **Trap** — used informally for both. Don't fight it on the midterm.

The *handling* is identical: switch to kernel mode, save the PC, jump
to a fixed handler address. The *cause* tells the handler what to do.

## Coprocessor 0

CP0 is a separate register file inside the CPU dedicated to system
state. The relevant ones for this guide:

| Register | Number | Purpose |
|---|---|---|
| `BadVAddr` | 8 | virtual address that caused a memory exception |
| `Status` | 12 | mode bits, interrupt enable, masks |
| `Cause` | 13 | why the exception fired (ExcCode field) |
| `EPC` | 14 | Exception Program Counter — where to resume |

You read/write CP0 with `mfc0` and `mtc0`:

```mips
mfc0 $t0, $13      # $t0 = Cause
mtc0 $t0, $12      # Status = $t0
```

`Status` packs a lot:

```text
| ... | IM[7:0] | KSU | EXL | IE |
       interrupt mask kernel  exception   global
                       /user  level       interrupt
                       mode   (in handler) enable
```

When an exception fires the hardware:

1. Saves the current PC into `EPC`.
2. Encodes the cause into `Cause.ExcCode`.
3. Sets `Status.EXL = 1` (in exception mode, interrupts masked).
4. Sets `Status.KSU` to kernel mode.
5. Jumps to the exception vector — `0x80000180` in MIPS32.

## Cause codes (subset)

The 5-bit `ExcCode` field tells the handler what happened:

| Code | Meaning |
|---|---|
| 0 | External interrupt |
| 4 | Address error on load / fetch |
| 5 | Address error on store |
| 6 | Bus error on instruction fetch |
| 7 | Bus error on data load/store |
| 8 | `syscall` |
| 9 | `break` |
| 10 | Reserved instruction |
| 12 | Arithmetic overflow |

A unified handler is therefore a switch on `ExcCode`.

## A minimal handler

```mips
        .ktext 0x80000180    # MIPS32 exception vector
handler:
        mfc0  $k0, $13        # $k0 = Cause
        andi  $k1, $k0, 0x7C  # ExcCode << 2 (5 bits, shifted)
        beq   $k1, 0x20, syscall_handler   # ExcCode 8 << 2
        beq   $k1, 0x30, overflow_handler  # ExcCode 12 << 2
        # ... etc
        # default: kill the process
        j     panic

syscall_handler:
        # ... do the syscall ...
        mfc0  $k0, $14        # $k0 = EPC
        addiu $k0, $k0, 4     # advance past the syscall instruction
        mtc0  $k0, $14
        eret                  # return: restores PC = EPC, EXL = 0
```

Notice the use of `$k0`/`$k1`: those registers are **reserved for the
kernel** (see {{GUIDE:cpts260-mips-cheatsheet}}) precisely so the
handler can scribble on them without saving anything first.

## Restarting vs advancing

For most exceptions you want to **re-execute** the faulting instruction
after the handler fixes the cause (e.g. demand-paging it in). For that
you set `EPC` back to the original PC and `eret`.

For `syscall` and `break` you want to **skip** the trap instruction —
otherwise you would loop forever. Hence the `addiu $k0, $k0, 4` above.

For exceptions inside a branch delay slot, MIPS sets `Cause.BD = 1` and
puts the *branch* PC into `EPC`. The handler must re-execute the entire
branch, otherwise it will not know whether to continue at PC + 4 or at
the branch target.

## Pipeline implications

An exception during EX, MEM, or WB requires squashing all younger
instructions in earlier stages. This is identical to a branch
misprediction — the pipeline is built to handle the squash + restart
sequence in the same machinery (see {{GUIDE:cpts260-pipelining-hazards}}
for the underlying squash signals).

The harder problem is **precise exceptions**: the architectural state
must look as if every instruction *before* the faulting one completed
and no instruction *after* it had any effect. In-order pipelines get
this almost for free; out-of-order pipelines need a re-order buffer
(ROB) to maintain the illusion.

## Interrupt prioritisation

Multiple interrupts can be pending at once. `Status.IM[7:0]` masks them
individually, and the hardware picks the highest-priority unmasked one.
Once the handler enters, `Status.EXL = 1` blocks all further interrupts
until the handler explicitly re-enables them — important to prevent
re-entrancy bugs.

A polite handler:

1. Saves enough state to be re-entrant.
2. Lowers `Status.EXL` and `Status.IE` to allow nested interrupts.
3. Does its work.
4. Restores state and `eret`s.

OS courses (CPTS 360 at WSU) go deep on this; CPTS 260 just needs you
to recognise the pattern.

## Pitfalls

- **Touching `$k0`/`$k1` in user code.** They will be clobbered the
  next time anything traps. Convention says hands off.
- **Forgetting the BD bit.** Exception in a branch delay slot needs
  special handling.
- **Re-entrancy.** Don't enable interrupts in your handler before
  saving the state you care about.

## Going deeper

The page-fault path is one of the most common synchronous exceptions —
its handler runs against the page table, sets the valid bit, and
restarts the faulting instruction. For where to actually *run* a tiny
exception-aware program, see {{GUIDE:cpts260-spim-lab-walkthrough}}.
