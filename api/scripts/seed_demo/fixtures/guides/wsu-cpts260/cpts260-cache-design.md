---
slug: cpts260-cache-design
course:
  ipeds_id: "236939"
  department: "CPTS"
  number: "260"
title: "Cache Design — Mapping, Associativity, and Replacement"
description: "How a cache is laid out, how addresses get split into tag/index/offset, and what associativity buys you."
tags: ["cache", "memory", "associativity", "replacement", "midterm"]
author_role: bot
quiz_slug: cpts260-cache-design-quiz
attached_files:
  - wsu-cpts260-cache-hierarchy-notes
attached_resources: []
---

# Cache Design

The previous guide ({{GUIDE:cpts260-memory-hierarchy}}) made the case for
caches at all. This one gets specific about how a cache is built — how
addresses are sliced, what "set-associative" actually means in hardware,
and how to compute hit rates by hand. The instructor's notes in
`{{FILE:wsu-cpts260-cache-hierarchy-notes}}` walk through the same
diagrams used in lecture.

## Address breakdown

Every byte address splits into three fields that the cache uses directly:

```text
| tag                         | index    | offset |
|<--- (32 - i - o) bits ----->|<-i bits->|<-o b->|
```

- **offset** — selects a byte within a cache line. For a 64-byte line
  that is the bottom 6 bits.
- **index** — selects a *set* in the cache. Width = log₂(number of sets).
- **tag** — the rest. Used to confirm we found the right line in the
  set.

### Worked example: 32 KB direct-mapped, 64-byte lines

- Cache size = 32 768 B
- Line size = 64 B
- Number of lines = 32 768 / 64 = 512
- Direct-mapped → 1 line per set → 512 sets
- offset = log₂(64) = 6 bits
- index = log₂(512) = 9 bits
- tag = 32 − 9 − 6 = 17 bits

So address `0xDEADBEEF` decomposes as:

```text
0xDEADBEEF = 1101 1110 1010 1101 1011 1110 1110 1111
                tag (17)        | index (9) | offset (6)
              1101 1110 1010 1101 1 | 011 1110 11 | 10 1111
```

## Direct-mapped vs set-associative vs fully-associative

| Style | Lines per set | Tag compares | Hit rate | Hardware cost |
|---|---|---|---|---|
| Direct-mapped | 1 | 1 | worst | cheapest |
| N-way set-associative | N | N | better | N comparators |
| Fully associative | all | all | best | impractical past ~32 entries |

The trade-off is simple: more associativity reduces **conflict misses**
but costs more comparators and a slower hit time (the data path now has
to MUX between N candidate lines). In practice L1 is 4-way or 8-way, L2
is 8-way to 16-way, and L3 is 16-way+.

### Why associativity matters

Consider a direct-mapped cache with 4 sets, addresses 0, 16, 0, 16
(assuming each address maps to set 0). Every access conflicts with the
previous, so the hit rate is 0%. With 2-way associativity both lines
fit in set 0 and the hit rate goes to 50%. Same total capacity, very
different behaviour.

## The three Cs of misses

Every miss falls into one of three categories — the "3C model":

1. **Compulsory** (cold) — first time you ever touched this line. You
   cannot avoid these without prefetching.
2. **Capacity** — the working set exceeds the cache size. Even a fully
   associative cache misses these. Cure: bigger cache, or smaller
   working set (blocking, tiling).
3. **Conflict** — the working set fits, but two hot lines map to the
   same set. Cure: more associativity.

Some textbooks add a fourth C, **Coherence**, for misses caused by
another core invalidating your line.

## Replacement policies

Once a set is full and a new line arrives, you must evict something.
Common policies:

- **LRU (Least Recently Used)** — evict the line untouched longest.
  Optimal in many real workloads, expensive past 4-way (needs N! orderings).
- **Pseudo-LRU** — tree of bits approximating LRU. What real chips use.
- **FIFO** — evict the oldest insertion. Cheap, occasionally pathological.
- **Random** — flip a coin. Surprisingly competitive at high associativity.
- **MRU** — evict the most recently used. Useful for streaming workloads
  where you know you will not revisit.

For a 2-way cache, "true LRU" is one bit per set: 0 means way 0 was
older, 1 means way 1 was older. Trivial.

## Hand-tracing a cache

Given a 4-line, 2-way set-associative cache (so 2 sets), 16-byte lines,
LRU replacement, and the byte-address access stream:

`0x00, 0x10, 0x20, 0x30, 0x00, 0x40, 0x10`

Address layout: offset = 4 bits, index = 1 bit, tag = the rest.

| # | Addr | Tag | Set | Action | Result |
|---|---|---|---|---|---|
| 1 | 0x00 | 0 | 0 | load into way 0 | miss (compulsory) |
| 2 | 0x10 | 0 | 1 | load into way 0 | miss (compulsory) |
| 3 | 0x20 | 1 | 0 | load into way 1 | miss (compulsory) |
| 4 | 0x30 | 1 | 1 | load into way 1 | miss (compulsory) |
| 5 | 0x00 | 0 | 0 | tag 0 hits in way 0 | **hit** |
| 6 | 0x40 | 2 | 0 | set 0 full → evict LRU (way 1, tag 1) | miss (conflict) |
| 7 | 0x10 | 0 | 1 | tag 0 hits in way 0 | **hit** |

Hit rate: 2 / 7 ≈ 28.6%. Most of the misses here are compulsory; conflict
misses become significant as the trace grows.

## Designing for the cache

A few rules of thumb that fall out of all this:

- **Block your loops.** Tile matrix multiplication so each tile fits in
  L1 — 10–30× speedup is routine.
- **Pad to avoid conflicts.** Power-of-two strides land everything in
  the same set. A 1024-element array of 4-byte ints, accessed by stride
  1024, will trash a 4 KB direct-mapped cache.
- **Pack hot data.** Splitting a struct into a hot half and a cold half
  doubles your effective L1 capacity if the cold fields are rarely
  touched.
- **Mind alignment.** A 64-byte cache line that straddles two pages
  takes two TLB lookups instead of one — small effect per access, but
  it adds up on misaligned bulk copies.

## Practice

When you can compute tag/index/offset from a cache geometry without the
formula in front of you, take {{QUIZ:cpts260-cache-design-quiz}}.
After that, the natural next stop is {{GUIDE:cpts260-pipelining-hazards}}
since cache misses are the largest single source of pipeline stalls.

For the bigger map of the course, see {{COURSE:wsu/cpts260}}.
