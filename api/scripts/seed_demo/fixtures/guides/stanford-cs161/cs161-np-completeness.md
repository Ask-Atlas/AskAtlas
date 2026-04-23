---
slug: cs161-np-completeness
course:
  ipeds_id: "243744"
  department: "CS"
  number: "161"
title: "NP-Completeness — CS 161"
description: "P, NP, polynomial-time reductions, Cook-Levin, and the standard catalog of NP-complete problems."
tags: ["algorithms", "np-completeness", "complexity", "reductions", "final"]
author_role: bot
quiz_slug: cs161-np-completeness-quiz
attached_files:
  - stanford-cs161-big-o-cheatsheet
attached_resources: []
---

# NP-Completeness

The last unit of CS 161 is not about designing algorithms. It is about proving that, for certain problems, **no efficient algorithm exists** — at least not unless one of the most consequential open questions in mathematics resolves the wrong way.

## Decision problems and the classes P and NP

Complexity classes are defined for **decision problems** (yes/no answers). Optimization problems get rephrased: instead of "find the minimum vertex cover," we ask "is there a vertex cover of size $\leq k$?" The decision version is at most as hard as the optimization version, and usually equivalent up to polynomial factors.

- **P** = the class of decision problems solvable in **polynomial time** by a deterministic algorithm.
- **NP** = the class of decision problems whose YES instances admit a **polynomial-time-checkable certificate**.

Notice the asymmetry: NP requires only that we can verify a YES answer quickly, given a hint. We don't need to find the hint quickly. The class **co-NP** is the analogue for NO answers.

Examples:

- **SAT** $\in$ NP: certificate is a satisfying assignment.
- **Hamiltonian Cycle** $\in$ NP: certificate is the cycle itself.
- **Subset Sum** $\in$ NP: certificate is the chosen subset.
- **Sorting** $\in$ P: trivially, we don't even need a certificate.

Crucially, every problem in P is also in NP — just ignore the certificate. The famous open question is whether $P = NP$ (most experts believe NO).

## Polynomial-time reductions

Problem $A$ **polynomial-time reduces** to problem $B$, written $A \leq_P B$, if there is a polynomial-time computable function $f$ that maps instances of $A$ to instances of $B$ such that $x$ is a YES instance of $A$ iff $f(x)$ is a YES instance of $B$.

If $A \leq_P B$ and $B \in P$, then $A \in P$. Contrapositive: if $A \leq_P B$ and $A$ is hard, then $B$ is hard.

This contrapositive is the **proof technique** for NP-hardness. To show that some new problem $B$ is NP-hard, you exhibit a polynomial-time reduction from a known NP-hard problem $A$ to $B$.

## NP-completeness

A problem $B$ is **NP-complete** if:

1. $B \in NP$, AND
2. Every problem in NP polynomial-time reduces to $B$ (i.e., $B$ is **NP-hard**).

The Cook–Levin theorem (1971) proved that **SAT** — Boolean satisfiability — is NP-complete. This was the seed: once you know SAT is NP-complete, you can prove other problems NP-complete by reducing SAT (or a known NP-complete problem) to them.

## Worked reduction: 3-SAT $\leq_P$ Independent Set

**3-SAT** = decide whether a Boolean formula in 3-CNF is satisfiable.
**Independent Set** = decide whether a graph has $k$ pairwise non-adjacent vertices.

**Reduction.** Given a 3-CNF formula $\phi$ with $m$ clauses:

1. Create a vertex for each literal occurrence in $\phi$. Total: $3m$ vertices.
2. Within each clause, connect all 3 vertices to each other (forming triangles).
3. Across clauses, connect every literal $x$ to every occurrence of $\neg x$.
4. Set $k = m$.

**Claim.** $\phi$ is satisfiable iff this graph has an independent set of size $m$.

**Forward.** Given a satisfying assignment, pick one true literal from each clause. The triangle edges prevent picking two from the same clause; the negation edges prevent picking $x$ and $\neg x$. So the chosen $m$ vertices form an independent set.

**Backward.** Given an independent set of size $m$, the triangle constraint means at most one vertex per clause; with $m$ clauses, exactly one per clause. The negation constraint means no $x$ and $\neg x$ are both chosen, so the chosen literals are consistent. Set those literals to true; this satisfies $\phi$.

The reduction is computable in polynomial time, and yes-instances map to yes-instances. So Independent Set is NP-hard, and (since it's in NP) NP-complete.

## The standard catalog

CS 161 expects you to know that the following are NP-complete and to have at least a one-sentence sketch of how to reduce between them:

- **SAT** and **3-SAT** (Cook–Levin, then Karp's reduction)
- **Independent Set**, **Vertex Cover**, **Clique** (mutually equivalent under simple complement / size transformations)
- **Hamiltonian Path** and **Hamiltonian Cycle**
- **Traveling Salesman** (decision version: is there a tour of length $\leq k$?)
- **Subset Sum** and **Knapsack** (knapsack is NP-hard despite the $\Theta(nW)$ DP — see below)
- **Graph Coloring** (specifically 3-Coloring)
- **Set Cover**

Useful reduction relationships (arrows = "reduces to"):

```
3-SAT ──> Independent Set ──> Clique
   │              │
   ├──> 3-Coloring   └──> Vertex Cover
   │
   └──> Subset Sum ──> Knapsack ──> Bin Packing
```

## "Pseudo-polynomial" and weak NP-hardness

Knapsack admits a $\Theta(nW)$ dynamic-programming algorithm — see {{GUIDE:cs161-dynamic-programming}}. Why is it still NP-hard?

Because the running time depends on the **value** of $W$, not the **bit length** of $W$. Encoding $W$ requires $\log_2 W$ bits, so a polynomial-in-input-size algorithm would need to run in $O(\text{poly}(n, \log W))$. The DP runs in $O(nW) = O(n \cdot 2^{\log W})$, which is exponential in input size when $W$ is large.

Knapsack is **weakly NP-hard** — solvable in pseudo-polynomial time. Other problems (3-SAT, TSP) are **strongly NP-hard**: even bounding the numerical inputs to $O(\text{poly}(n))$ doesn't help.

## What to do when you face an NP-hard problem

The course doesn't end at "this is hard, give up." Practical strategies:

1. **Special cases.** Many NP-hard problems become polynomial on restricted inputs (e.g., 2-SAT is polynomial; Vertex Cover on trees is polynomial).
2. **Approximation algorithms.** Find a polynomial-time algorithm that returns a solution provably within some factor of optimal. Vertex Cover has a 2-approximation; TSP with metric distances has a 1.5-approximation (Christofides).
3. **Heuristics with no guarantee.** Simulated annealing, branch and bound, ILP solvers — work great in practice for many real instances even though worst-case behavior is exponential.
4. **Parameterized algorithms.** Algorithms whose running time is polynomial in $n$ but exponential in some other parameter $k$ (FPT — Fixed Parameter Tractable). Useful when the parameter is small in practice.

## Why this matters

The reason $P$ vs $NP$ is the most famous open problem in CS is that thousands of problems we genuinely care about — protein folding, optimal scheduling, secure encryption, perfect game playing — sit in NP but not (apparently) in P. Resolving the question would either deliver a universal solver (P = NP) or formally settle that some problems are intrinsically intractable.

For now, learn to **recognize** NP-hard problems quickly. The hour you spend trying to find a polynomial-time algorithm for an NP-hard problem is an hour wasted; the hour you spend constructing an approximation algorithm is an hour well-spent.

Take {{QUIZ:cs161-np-completeness-quiz}} for practice — it covers the definitions, one reduction from scratch, and "is this problem in P, NP, or NP-complete?" classification questions.

## Looking forward

CS 161's algorithm narrative — from {{GUIDE:cs161-asymptotic-analysis}} through {{GUIDE:cs161-graph-algorithms}} — has been about finding efficient algorithms. NP-completeness reframes the question: when no efficient algorithm exists, what do we do instead? That question motivates a large fraction of modern algorithms research.
