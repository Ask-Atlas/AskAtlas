---
slug: cs161-greedy-algorithms
course:
  ipeds_id: "243744"
  department: "CS"
  number: "161"
title: "Greedy Algorithms — CS 161"
description: "Activity selection, Huffman coding, Kruskal/Prim, and the exchange-argument proof technique."
tags: ["algorithms", "greedy", "huffman", "mst", "proof-technique", "final"]
author_role: bot
attached_files:
  - stanford-cs161-dp-patterns
  - stanford-cs161-graph-algorithms-worksheet
attached_resources: []
---

# Greedy Algorithms

A **greedy algorithm** makes the locally optimal choice at every step and never reconsiders. When greedy works, it produces astonishingly short, fast algorithms. When greedy fails — and it often does — it fails subtly and silently, returning suboptimal answers without any sign of error.

The hard part is not writing greedy algorithms. The hard part is **proving they are correct**.

## When greedy works

A problem admits a greedy algorithm when it has both:

1. **Greedy choice property.** Some locally optimal choice is part of *some* globally optimal solution.
2. **Optimal substructure.** After making that greedy choice, what remains is a smaller instance of the same problem whose optimal solution combines with the greedy choice to yield a global optimum.

Most exam problems hand you a candidate greedy rule and ask whether it works. Your job: prove correctness or construct a counterexample.

## Activity selection

Given $n$ activities each with a start and finish time, select the maximum number of mutually non-overlapping activities.

**Greedy rule.** Repeatedly pick the activity with the **earliest finish time** that does not conflict with previously chosen activities.

```python
def select(activities):  # activities: list of (start, finish)
    activities.sort(key=lambda a: a[1])
    chosen = []
    last_finish = float("-inf")
    for s, f in activities:
        if s >= last_finish:
            chosen.append((s, f))
            last_finish = f
    return chosen
```

Time $\Theta(n \log n)$ for the sort; $\Theta(n)$ for the scan.

**Proof (exchange argument).** Let $G$ be the activity chosen first by the greedy algorithm and let $O$ be any optimal solution. If $G \in O$, recurse. If $G \notin O$, let $A$ be the activity in $O$ with the earliest finish time. Since $G$ has the earliest finish time of all activities, $G$'s finish time $\leq$ $A$'s finish time. Replace $A$ in $O$ with $G$: every activity that didn't conflict with $A$ also doesn't conflict with $G$ (because $G$ finishes at least as early). The size of $O$ is unchanged, so the modified solution is still optimal — and now contains $G$. Recurse on the rest.

The exchange argument is the **dominant** proof technique for greedy correctness. Memorize its shape: assume an optimal solution differs from greedy, swap, show the swap doesn't reduce optimality.

## Huffman coding

Given character frequencies $f_1, \ldots, f_n$, find a prefix-free binary code that minimizes the total encoded length $\sum f_i \cdot |\text{code}_i|$.

**Greedy rule.** Repeatedly take the two least frequent characters, merge them into a single virtual character with frequency equal to their sum, and recurse. This builds the code tree bottom-up.

```python
import heapq

def huffman(freqs):  # freqs: list of (char, frequency)
    heap = [[f, [c, ""]] for c, f in freqs]
    heapq.heapify(heap)
    while len(heap) > 1:
        lo = heapq.heappop(heap)
        hi = heapq.heappop(heap)
        for pair in lo[1:]:
            pair[1] = "0" + pair[1]
        for pair in hi[1:]:
            pair[1] = "1" + pair[1]
        heapq.heappush(heap, [lo[0] + hi[0]] + lo[1:] + hi[1:])
    return sorted(heap[0][1:], key=lambda p: (len(p[-1]), p))
```

Time $\Theta(n \log n)$ using a min-heap.

The correctness proof uses the same exchange-argument template: in any optimal tree, the two deepest leaves can be swapped to be the two least-frequent characters without increasing the cost.

## Kruskal's algorithm (MST)

Given a connected, weighted, undirected graph $G = (V, E)$, find a minimum spanning tree.

**Greedy rule.** Sort edges by weight ascending. Repeatedly add the next edge if it does not form a cycle with already-chosen edges. Stop after $|V| - 1$ edges.

```
sort E by weight
DSU = DisjointSet(V)
T = []
for (u, v) in E:
    if DSU.find(u) != DSU.find(v):
        DSU.union(u, v)
        T.append((u, v))
return T
```

Time $\Theta(E \log V)$ using a Disjoint Set Union with path compression and union by rank. The cycle check via DSU is amortized $\alpha(V)$ per operation — effectively constant for any input size you'll see in practice.

## Prim's algorithm (MST)

Same problem, different greedy rule:

**Greedy rule.** Start at any vertex. Repeatedly add the lightest edge crossing the cut between vertices already in the tree and vertices outside.

Time $\Theta(E \log V)$ with a binary heap, or $\Theta(E + V \log V)$ with a Fibonacci heap. Prim wins on dense graphs (closer to $\Theta(V^2)$ in the dense limit when implemented with adjacency matrices).

Both Kruskal and Prim are correct because of the **cut property**: for any cut $(S, V \setminus S)$, the minimum-weight edge crossing the cut is in some MST.

For a full graph-algorithms refresher, see {{GUIDE:cs161-graph-algorithms}} and the worksheet in {{FILE:stanford-cs161-graph-algorithms-worksheet}}.

## When greedy fails

The classic trap: **the coin change problem**.

Given denominations $\{1, 5, 10, 25\}$ (US coins), the greedy rule "take as many of the largest coin as possible" produces optimal change for any amount. Problem: this is a **special property of US coinage**.

Counterexample: with denominations $\{1, 5, 12\}$ and target $15$, greedy gives $12 + 1 + 1 + 1 = 4$ coins, but $5 + 5 + 5 = 3$ coins is better. The general coin change problem requires DP — see {{GUIDE:cs161-dynamic-programming}}.

This is why CS 161 emphasizes **proofs**, not intuition. Many greedy strategies look reasonable until you spend ten minutes hunting for a counterexample.

## Decision template

When you see an optimization problem on the exam, try in this order:

1. Is there an obvious greedy rule? Try to prove it via exchange argument. If you can prove it, you're done.
2. If you can't prove the greedy rule and can't find a counterexample within 5 minutes, try to find a counterexample more aggressively. Often the smallest counterexample has $n = 3$ or $n = 4$.
3. If greedy provably fails, fall back to DP. Identify the subproblem and write the recurrence.
4. If DP state space is too large, look for a polynomial-time reduction to a known graph problem (matching, max-flow, MST).

For more pattern templates, see {{FILE:stanford-cs161-dp-patterns}} (it covers when DP outperforms greedy).

For related drilling on optimization techniques, take the {{QUIZ:cs161-dynamic-programming-quiz}} — it includes "is greedy enough or do you need DP?" judgment questions that pair naturally with this material.
