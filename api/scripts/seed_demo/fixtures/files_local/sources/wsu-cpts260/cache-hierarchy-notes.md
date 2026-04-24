---
slug: wsu-cpts260-cache-hierarchy-notes
title: "Cache Hierarchy Study Notes"
mime: application/pdf
filename: cache-hierarchy-notes.pdf
course: wsu/cpts260
description: "L1/L2/L3 organization, hit/miss behavior, associativity, and the arithmetic behind average memory access time."
author_role: bot
---

## Why a Hierarchy?

DRAM access costs roughly 200 cycles; a register read is free. Caches sit between the CPU and main memory to exploit **temporal locality** (a line reused soon) and **spatial locality** (nearby words accessed together). A modern desktop part typically looks like this:

| Level | Size        | Latency      | Associativity | Line size |
|-------|-------------|--------------|---------------|-----------|
| L1-I  | 32 KB       | ~4 cycles    | 8-way         | 64 B      |
| L1-D  | 32 KB       | ~4 cycles    | 8-way         | 64 B      |
| L2    | 256 KB–1 MB | ~12 cycles   | 8-way         | 64 B      |
| L3    | 8–32 MB     | ~40 cycles   | 16-way        | 64 B      |
| DRAM  | 16 GB+      | ~200 cycles  | —             | —         |

L1 is split I/D to let fetch and load/store issue in parallel. L2 and L3 are unified.

## Address Decomposition

For a cache with `B` bytes per line, `S` sets, and `W` ways, a 32-bit address is split:

```
| tag  | set index | block offset |
```

- **Block offset**: `log2(B)` bits — which byte inside the line
- **Set index**: `log2(S)` bits — which set to look up
- **Tag**: the rest — compared against all `W` ways in the set

A fully-associative cache has `S = 1` (tag-only); a direct-mapped cache has `W = 1`.

## Hits, Misses, and the Three Cs

Every miss falls into one of three categories:

1. **Compulsory** (cold) — first reference to a line. Cannot be avoided by a larger cache.
2. **Capacity** — working set exceeds cache size. Fixed by making the cache bigger.
3. **Conflict** — too many live lines map to the same set. Fixed by raising associativity.

## Average Memory Access Time

```c
// AMAT with two-level cache
AMAT = hit_time_L1 + miss_rate_L1 * (hit_time_L2 + miss_rate_L2 * miss_penalty_mem);
```

Example: L1 hits in 4 cycles at 95%; L2 hits in 12 cycles at 80%; DRAM is 200 cycles.

```
AMAT = 4 + 0.05 * (12 + 0.20 * 200)
     = 4 + 0.05 * 52
     = 6.6 cycles
```

## Write Policies

- **Write-through** + **write-allocate** — simple, but traffic-heavy
- **Write-back** + **write-allocate** — typical for L1-D; dirty bit tracks modified lines

## Replacement

Pure LRU is expensive past 4 ways. Real hardware uses **pseudo-LRU** trees or **re-reference interval prediction (RRIP)**. Random replacement is surprisingly competitive at high associativity.

## Coherence

Multi-core systems run **MESI** (or MOESI): each line is Modified, Exclusive, Shared, or Invalid. A store on a Shared line issues an **invalidate** on the bus, forcing other cores to re-fetch.
