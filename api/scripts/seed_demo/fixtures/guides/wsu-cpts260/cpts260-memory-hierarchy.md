---
slug: cpts260-memory-hierarchy
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "260"
title: "The Memory Hierarchy — Why Caches Exist"
description: "Locality, the memory wall, and the layered storage model that makes modern CPUs tolerable."
tags: ["memory", "cache", "hierarchy", "locality", "performance"]
author_role: bot
attached_files:
  - wsu-cpts260-memory-hierarchy-excerpt
attached_resources: []
---

# The Memory Hierarchy

Modern CPUs would be useless without caches. A 4 GHz core executes one
instruction every 0.25 ns; a DRAM access takes around 70 ns. That is a
**280-cycle stall** for every uncached load. The memory hierarchy is the
engineering workaround for that gap, and CPTS 260 spends a lot of time on
it because every higher-level performance topic — out-of-order execution,
prefetching, NUMA, GPUs — falls out of the same locality argument.

The full chapter excerpt is in
`{{FILE:wsu-cpts260-memory-hierarchy-excerpt}}`. This guide is the bird's-
eye view that ties it together.

## The hierarchy itself

Typical 2024-era desktop CPU:

| Level | Size | Latency (cycles) | Bandwidth |
|---|---|---|---|
| Registers | ~1 KB total | 0 | trivially infinite |
| L1 cache (per core) | 32–48 KB | 4 | ~1 TB/s |
| L2 cache (per core) | 256 KB–1 MB | 12 | ~500 GB/s |
| L3 cache (shared) | 8–32 MB | 40 | ~200 GB/s |
| DRAM | 16–128 GB | 200–300 | ~50 GB/s |
| NVMe SSD | 1–8 TB | 100,000+ | ~7 GB/s |
| Spinning disk | 4–20 TB | 10,000,000+ | ~200 MB/s |

Each level is roughly **10× larger and 10× slower** than the one above
it. That ratio is not a coincidence — it falls out of the cost-per-bit
curve for SRAM vs DRAM vs flash vs magnetic media.

## The two locality principles

Caches work because real programs do not access memory uniformly. They
exhibit two kinds of locality:

1. **Temporal locality** — if you accessed an address recently, you are
   likely to access it again soon. Loop induction variables, function
   parameters, and the top of the stack live here.
2. **Spatial locality** — if you accessed an address, you are likely to
   access nearby addresses soon. Sequential array scans, struct field
   accesses, and code execution all show this.

Caches exploit temporal locality by *keeping* recently-touched data, and
exploit spatial locality by fetching whole **cache lines** (typically
64 bytes) rather than individual bytes.

## Average memory access time (AMAT)

The single equation you must memorise:

```text
AMAT = HitTime + MissRate × MissPenalty
```

For a multi-level hierarchy this is recursive:

```text
AMAT_L1 = HitTime_L1 + MissRate_L1 × AMAT_L2
AMAT_L2 = HitTime_L2 + MissRate_L2 × AMAT_L3
```

A 95% L1 hit rate looks great until you do the arithmetic with a 200-
cycle DRAM penalty: `4 + 0.05 × 200 = 14` cycles per access. That is
why L2 and L3 exist — to keep the miss penalty small even when L1
misses.

### Worked example

Given `HitTime_L1 = 2`, `MissRate_L1 = 8%`, `HitTime_L2 = 12`,
`MissRate_L2 = 30%`, `MissPenalty_L2 = 200`:

```text
AMAT_L2 = 12 + 0.30 × 200 = 72 cycles
AMAT_L1 = 2 + 0.08 × 72 = 7.76 cycles
```

Cut the L1 miss rate in half (to 4%) and you get 4.88 cycles —
essentially a 2× speedup on memory-bound code, just by tuning the cache.

## Inclusion, exclusion, and write policies

Two design axes that come up on the midterm:

### Inclusive vs exclusive

- **Inclusive** caches keep everything in L1 also present in L2/L3.
  Simpler coherence; wastes some capacity.
- **Exclusive** caches guarantee no overlap between levels. More usable
  capacity; harder coherence (Intel uses inclusive, AMD has historically
  used exclusive at L3).

### Write-through vs write-back

- **Write-through** sends every store to the next level immediately.
  Simpler, deterministic; saturates L2 bandwidth.
- **Write-back** marks the line dirty and flushes only on eviction.
  Almost always the right choice for L1; almost universally used today.

### Write-allocate vs no-write-allocate

On a *store miss*, do you bring the line into the cache?

- **Write-allocate** fetches the line so subsequent stores hit. Pairs
  naturally with write-back.
- **No-write-allocate** sends the store straight through. Pairs with
  write-through.

## Why this is more than trivia

Once you understand the hierarchy you can read code and predict its
behaviour. Compare:

```c
// Row-major scan of a 1024x1024 int matrix
for (int i = 0; i < 1024; i++)
    for (int j = 0; j < 1024; j++)
        sum += A[i][j];
```

vs

```c
// Column-major scan of the same matrix
for (int j = 0; j < 1024; j++)
    for (int i = 0; i < 1024; i++)
        sum += A[i][j];
```

The first version touches 16 ints (64 bytes) per cache line and gets a
fresh line every 16 accesses — a hit rate around 94%. The second version
touches one int per line and evicts it before the next column iteration
needs it — a hit rate near 0% on caches smaller than the working set.
Same algorithmic complexity, **10–20× wall-clock difference**.

The next guide, {{GUIDE:cpts260-cache-design}}, gets specific about how
the cache is *organised* — direct-mapped vs set-associative vs fully
associative, and what changing the associativity buys you.

## Pitfalls

- **Confusing latency and bandwidth.** L1 is faster *per access*; DRAM
  is not slower *per gigabyte streamed*. Both metrics matter.
- **Counting cycles, not nanoseconds.** Hierarchies move with clock
  speeds; relative ratios are stable.
- **Forgetting the TLB.** Address translation has its own cache (the
  Translation Lookaside Buffer) — every memory access pays the TLB
  cost on top of the data-cache cost.

## Going deeper

The Hennessy & Patterson chapter excerpt
`{{FILE:wsu-cpts260-memory-hierarchy-excerpt}}` includes the full
write-policy decision matrix. Skim it; you will see this material on
the final.
