---
slug: cs161-graph-algorithms
course:
  ipeds_id: "243744"
  department: "CS"
  number: "161"
title: "Graph Algorithms — CS 161"
description: "BFS, DFS, topological sort, Dijkstra, Bellman-Ford, Floyd-Warshall, and the SCC pipeline."
tags: ["algorithms", "graphs", "shortest-paths", "bfs", "dfs", "final"]
author_role: bot
quiz_slug: cs161-graph-algorithms-quiz
attached_files:
  - stanford-cs161-graph-algorithms-worksheet
  - stanford-cs161-big-o-cheatsheet
attached_resources: []
---

# Graph Algorithms

A graph $G = (V, E)$ models pairwise relationships. The CS 161 graph unit gives you a **toolbox** of around eight algorithms; the exam tests whether you can pick the right tool for a given problem.

This guide is organized by problem type, not by algorithm. When you read a problem statement, classify it first; then the algorithm choice is automatic.

## Representation

Two standard choices:

- **Adjacency list.** `adj[u]` is a list of vertices reachable from `u`. Space $\Theta(V + E)$. Optimal for sparse graphs (most real-world graphs).
- **Adjacency matrix.** `M[u][v]` is the edge weight (or 0/1). Space $\Theta(V^2)$. Optimal for dense graphs and constant-time edge queries.

Default to adjacency list unless you have specific reasons. The trade-off shows up in algorithm complexity: e.g., Prim's MST is $\Theta(E \log V)$ with a heap on a list, but $\Theta(V^2)$ with a matrix — so for dense graphs, matrix is faster.

## Traversals

### BFS — Breadth First Search

```python
from collections import deque

def bfs(adj, s):
    dist = {s: 0}
    q = deque([s])
    while q:
        u = q.popleft()
        for v in adj[u]:
            if v not in dist:
                dist[v] = dist[u] + 1
                q.append(v)
    return dist
```

Time $\Theta(V + E)$. Computes shortest-path distances in **unweighted** graphs.

### DFS — Depth First Search

```python
def dfs(adj, s, visited):
    visited.add(s)
    for v in adj[s]:
        if v not in visited:
            dfs(adj, v, visited)
```

Time $\Theta(V + E)$. The classification of edges into tree, back, forward, cross edges is what powers many downstream algorithms (cycle detection, topological sort, SCC).

## Topological sort

For a directed acyclic graph (DAG), produce a linear ordering of vertices such that every edge $(u, v)$ has $u$ before $v$.

Two standard implementations:

1. **DFS-based.** Run DFS from every unvisited vertex; when DFS finishes a vertex, prepend it to the output. Reverse the post-order is the topological order.
2. **Kahn's algorithm.** Repeatedly remove a vertex with in-degree 0 and add it to the output; decrement in-degrees of its neighbors.

Both are $\Theta(V + E)$. Kahn's is easier to explain and yields a natural cycle detector — if the queue empties before all vertices are removed, the input wasn't a DAG.

## Single-source shortest paths

| Algorithm | Edge weights | Complexity | Notes |
|---|---|---|---|
| BFS | Unweighted | $\Theta(V + E)$ | All edges treated as weight 1 |
| Dijkstra | Non-negative | $\Theta((V + E) \log V)$ with binary heap | Fails on negative edges |
| Bellman–Ford | Any (incl. negative) | $\Theta(VE)$ | Detects negative-weight cycles |
| Topological + relax | DAG only | $\Theta(V + E)$ | Use when graph is acyclic |

### Dijkstra

```python
import heapq

def dijkstra(adj, s):  # adj[u] = list of (v, weight)
    dist = {s: 0}
    pq = [(0, s)]
    while pq:
        d, u = heapq.heappop(pq)
        if d > dist[u]:
            continue
        for v, w in adj[u]:
            nd = d + w
            if nd < dist.get(v, float("inf")):
                dist[v] = nd
                heapq.heappush(pq, (nd, v))
    return dist
```

Correctness rests on the **invariant**: when a vertex is popped from the priority queue, its current distance is final. This invariant breaks if any edge has negative weight — that's why Dijkstra fails there.

### Bellman–Ford

```
init dist[s] = 0, dist[v] = infinity for v != s
repeat |V| - 1 times:
    for each edge (u, v) with weight w:
        if dist[u] + w < dist[v]:
            dist[v] = dist[u] + w
# detect negative cycle
for each edge (u, v) with weight w:
    if dist[u] + w < dist[v]:
        return "negative cycle"
```

Time $\Theta(VE)$. After $k$ iterations, `dist[v]` equals the shortest distance using at most $k$ edges. Hence $V - 1$ iterations suffice for any simple path.

## All-pairs shortest paths

**Floyd–Warshall.** $\Theta(V^3)$ time, $\Theta(V^2)$ space. Cubic but conceptually trivial:

```python
def floyd_warshall(M):  # M[u][v] = edge weight or infinity
    n = len(M)
    for k in range(n):
        for i in range(n):
            for j in range(n):
                M[i][j] = min(M[i][j], M[i][k] + M[k][j])
```

The DP subproblem: "shortest path from $i$ to $j$ that uses only intermediate vertices in $\{1, \ldots, k\}$." Adding vertex $k+1$ either helps or doesn't.

For sparse graphs with non-negative weights, **Johnson's algorithm** (run Bellman–Ford once to reweight, then Dijkstra from every vertex) is $\Theta(V^2 \log V + VE)$ — better than Floyd–Warshall when $E \ll V^2$.

## Strongly connected components (Kosaraju)

For directed graphs:

1. Run DFS on $G$; push each vertex onto a stack when DFS finishes it.
2. Compute the reverse graph $G^T$.
3. Pop vertices off the stack one at a time; each DFS in $G^T$ from a popped vertex (skipping already-visited vertices) discovers exactly one SCC.

Time $\Theta(V + E)$. The two-pass structure surprises everyone the first time. The exchange argument: in the post-order from step 1, the LAST finishing vertex must lie in a "source" SCC of the SCC DAG. Reversing edges turns sources into sinks, so DFS from that vertex stays inside the SCC.

## Decision flowchart

```
Are edges weighted?
├─ NO ──> BFS
└─ YES ─> Are weights non-negative?
         ├─ YES ──> Is the graph a DAG?
         │        ├─ YES ──> Topological sort + relaxation
         │        └─ NO  ──> Dijkstra
         └─ NO  ──> Bellman–Ford (also detects negative cycles)
```

For all-pairs problems, prefer Floyd–Warshall on dense graphs and Johnson's algorithm on sparse ones.

For more practice problems, work through the {{FILE:stanford-cs161-graph-algorithms-worksheet}}, then take the {{QUIZ:cs161-graph-algorithms-quiz}}.

## What's next

The MST piece of this material lives in {{GUIDE:cs161-greedy-algorithms}}. The hardest graph problems — Hamiltonian path, traveling salesman, vertex cover — turn out to be NP-complete; we cover the consequences in {{GUIDE:cs161-np-completeness}}.
