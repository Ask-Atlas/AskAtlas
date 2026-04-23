---
slug: wsu-cpts260-memory-hierarchy-excerpt
title: "Memory Hierarchy: A Primer"
mime: application/epub+zip
filename: memory-hierarchy-excerpt.epub
course: wsu/cpts260
description: "Multi-section primer on the memory hierarchy — registers to DRAM, locality, caching, virtual memory, and TLBs."
author_role: bot
---

## Chapter 1 — The Memory Wall

Processor clocks have outpaced DRAM latency for three decades. A 4 GHz core can issue four instructions per cycle; a DRAM row activation takes roughly 50 nanoseconds, which is 200 of those cycles. If every load went to DRAM, the CPU would spend 99% of its life stalled. The **memory hierarchy** is the architectural answer: several levels of progressively larger, slower storage, with the fast levels acting as automatic caches for the slow ones.

The hierarchy succeeds because real programs are not random-access. Two empirical properties make caching work:

- **Temporal locality** — a location accessed once is likely to be accessed again soon. Loop bodies, stack frames, and hot data structures all exhibit this.
- **Spatial locality** — a location near one just accessed is likely to be accessed soon. Array traversal, sequential instruction fetch, and struct field access are all spatially local.

## Chapter 2 — The Levels

From fastest to slowest:

1. **Registers** — 32 × 32 bits in a classic MIPS integer file. Zero cycle access; compiler-managed.
2. **L1 cache** — 32 KB, 4–5 cycle, split I/D. Line size 64 B.
3. **L2 cache** — 256 KB to 1 MB, 10–15 cycle, unified.
4. **L3 cache** — 8–32 MB, 30–50 cycle, shared across cores.
5. **DRAM** — gigabytes, 150–300 cycle.
6. **SSD** — terabytes, microseconds to milliseconds.
7. **Spinning disk / network storage** — milliseconds to seconds.

Every level is roughly 10× the size of the one above it and 10× slower. This geometric progression is not a coincidence — it is what keeps each level's miss rate low enough that the next level's latency does not dominate.

## Chapter 3 — Caches in Detail

A cache stores fixed-size **lines** (typically 64 B). An address is split into tag, set index, and block offset. On each reference, the cache indexes into a set, compares the tag against each way, and on a match returns the word at the offset. On a miss it fetches the entire line from the next level, possibly evicting another line.

Associativity trades hit rate against latency and energy. Direct-mapped (1-way) caches are fastest but suffer conflict misses; fully associative caches have no conflict misses but require `N` parallel comparators. Set-associative caches (4 to 16 ways at L2/L3) are the practical middle ground.

## Chapter 4 — Virtual Memory

The operating system gives each process its own 32- or 64-bit address space. Translation from virtual to physical pages is done by a **page table** walked by hardware. Pages are typically 4 KB.

The page table itself lives in memory, so a naive translation would double every load. The **TLB** (translation lookaside buffer) caches recent translations — usually 64 to 1024 entries, fully associative or very high associativity. A TLB miss triggers a page-table walk; a page-table miss triggers a **page fault**, handed to the OS.

## Chapter 5 — Exercises

1. Compute AMAT for a three-level hierarchy given per-level hit times and miss rates.
2. Given a 32-bit address, a 64 B line, 4-way associativity, and 16 KB total size, how many bits are tag / index / offset?
3. Explain why `for (i=0; i<N; i++) for (j=0; j<N; j++) A[j][i]` runs slower than the loop-swapped version.
4. When does increasing associativity *not* reduce miss rate? Give a concrete example.
