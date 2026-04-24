---
slug: stanford-cs161-graph-algorithms-worksheet
title: "Graph Algorithms Worksheet: BFS, DFS, and Dijkstra"
mime: application/vnd.openxmlformats-officedocument.wordprocessingml.document
filename: graph-algorithms-worksheet.docx
course: stanford/cs161
description: "Practice worksheet covering BFS, DFS, and Dijkstra with traces, pseudocode, and short-answer prompts."
author_role: bot
---

# Graph Algorithms Worksheet

**Instructions.** Work each problem in order. Show intermediate state (queue, stack, or priority queue contents) for trace problems. Give a one-line justification for any complexity claim.

## Part 1: Representations

**Q1.** Given an undirected graph with `$|V| = n$` and `$|E| = m$`, compare adjacency matrix vs. adjacency list.

| Operation | Matrix | List |
|-----------|--------|------|
| Space | `$\Theta(n^2)$` | `$\Theta(n + m)$` |
| Edge exists? | `$O(1)$` | `$O(\deg(v))$` |
| Iterate neighbors | `$O(n)$` | `$O(\deg(v))$` |

When is the matrix preferable? When does the list win?

## Part 2: Breadth-First Search

```python
def bfs(graph, source):
    dist = {source: 0}
    parent = {source: None}
    queue = [source]
    while queue:
        u = queue.pop(0)
        for v in graph[u]:
            if v not in dist:
                dist[v] = dist[u] + 1
                parent[v] = u
                queue.append(v)
    return dist, parent
```

**Q2.** Given the graph `$V = \{A, B, C, D, E, F\}$` with edges `$\{(A,B), (A,C), (B,D), (C,D), (C,E), (D,F), (E,F)\}$`, run BFS from `$A$`. Write:

- The order vertices are dequeued.
- `dist[v]` for every `$v$`.
- The BFS tree edges.

**Q3.** Prove BFS computes shortest paths (edge count) in an unweighted graph. Hint: induction on distance from source.

**Q4.** What is the runtime of BFS using an adjacency list? Justify. _(Expected: `$\Theta(n + m)$`.)_

## Part 3: Depth-First Search

```python
def dfs(graph):
    color = {v: "WHITE" for v in graph}
    discover, finish = {}, {}
    time = [0]

    def visit(u):
        time[0] += 1
        discover[u] = time[0]
        color[u] = "GRAY"
        for v in graph[u]:
            if color[v] == "WHITE":
                visit(v)
        time[0] += 1
        finish[u] = time[0]
        color[u] = "BLACK"

    for u in graph:
        if color[u] == "WHITE":
            visit(u)
    return discover, finish
```

**Q5.** Run DFS on the directed graph `$\{(1,2), (1,3), (2,4), (3,4), (4,1), (5,3)\}$` starting from vertex 1, then 5. Label each edge as **tree**, **back**, **forward**, or **cross**.

**Q6.** How do you detect a cycle in a directed graph with DFS? In an undirected graph?

**Q7.** Write pseudocode for topological sort using DFS finish times.

## Part 4: Dijkstra's Algorithm

```python
import heapq

def dijkstra(graph, source):
    dist = {v: float("inf") for v in graph}
    dist[source] = 0
    pq = [(0, source)]
    while pq:
        d, u = heapq.heappop(pq)
        if d > dist[u]:
            continue
        for v, w in graph[u]:
            alt = dist[u] + w
            if alt < dist[v]:
                dist[v] = alt
                heapq.heappush(pq, (alt, v))
    return dist
```

**Q8.** Run Dijkstra from `$s$` on a weighted graph with edges `$(s,a,4), (s,b,1), (b,a,2), (a,t,1), (b,t,5)$`. List the priority queue state after each extraction.

**Q9.** Why does Dijkstra fail on graphs with negative edges? Construct a minimal counterexample.

**Q10.** Give the runtime with a binary heap: `$O((n + m) \log n)$`. With a Fibonacci heap: `$O(m + n \log n)$`.

## Part 5: Synthesis

**Q11.** You must find the shortest path from `$s$` to `$t$` on a graph where every edge has weight 1 or 2. Design an algorithm that runs in `$O(n + m)$`. Hint: modified BFS with a deque.

**Q12.** Given an undirected graph, describe how to find all connected components in linear time.
