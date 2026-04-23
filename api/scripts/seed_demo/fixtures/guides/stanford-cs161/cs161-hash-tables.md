---
slug: cs161-hash-tables
course:
  ipeds_id: "243744"
  department: "CS"
  number: "161"
title: "Hash Tables — CS 161"
description: "Chaining vs open addressing, universal hashing, load factor, and amortized resize analysis."
tags: ["algorithms", "hashing", "data-structures", "amortized-analysis", "midterm"]
author_role: bot
attached_files:
  - stanford-cs161-big-o-cheatsheet
attached_resources: []
---

# Hash Tables

A hash table maps keys to slots in a fixed-size array using a hash function $h : K \to \{0, 1, \ldots, m - 1\}$. With a good hash function and a sensible load factor, the expected cost of `insert`, `lookup`, and `delete` is $\Theta(1)$.

CS 161's treatment focuses on three things: collision resolution strategies, the load-factor analysis that makes the $\Theta(1)$ bound rigorous, and **universal hashing** as the technique that makes the bound robust against adversarial inputs.

## Collision resolution

When two keys hash to the same slot — a collision — the table needs a strategy to keep both.

### Chaining

Each slot holds a linked list of all elements that hash there.

```python
def insert(table, key, value):
    slot = h(key) % len(table)
    table[slot].append((key, value))

def lookup(table, key):
    slot = h(key) % len(table)
    for k, v in table[slot]:
        if k == key:
            return v
    return None
```

Expected lookup time is $\Theta(1 + \alpha)$ where $\alpha = n / m$ is the **load factor** (number of items / number of slots). For $\alpha = O(1)$, this is $\Theta(1)$.

Worst case: all $n$ keys hash to the same slot, giving $\Theta(n)$ lookup. This is why universal hashing matters — it bounds the *expected* worst case over the choice of hash function.

### Open addressing

Store everything directly in the array; on collision, probe to the next slot.

- **Linear probing.** Try $h(k), h(k) + 1, h(k) + 2, \ldots$ Simple, cache-friendly. Suffers from **primary clustering** — long runs of occupied slots that grow by attracting more collisions.
- **Quadratic probing.** Try $h(k), h(k) + 1, h(k) + 4, h(k) + 9, \ldots$ Reduces primary clustering but introduces **secondary clustering** (keys with the same initial hash follow the same probe sequence).
- **Double hashing.** Try $h_1(k), h_1(k) + h_2(k), h_1(k) + 2 h_2(k), \ldots$ The most uniform probing strategy in practice.

Open addressing requires $\alpha < 1$ strictly (you can't probe past full). The expected number of probes for a successful lookup is approximately $\frac{1}{\alpha} \ln \frac{1}{1 - \alpha}$ — finite for any $\alpha < 1$, but it explodes as $\alpha \to 1$. Most production tables resize when $\alpha > 0.75$.

## Universal hashing

A family $\mathcal{H}$ of hash functions $h: U \to \{0, \ldots, m-1\}$ is **universal** if for any two distinct keys $x, y \in U$:

$$
\Pr_{h \sim \mathcal{H}}[h(x) = h(y)] \leq \frac{1}{m}.
$$

This is exactly the collision rate you'd expect from a truly uniform random function. The trick is that the randomness lives in the *choice of hash function*, not the *input*. An adversary picking inputs after seeing the hash function can be devastating; an adversary picking inputs *before* the hash function is chosen cannot do better than chance.

A standard universal family for $U = \{0, 1, \ldots, p - 1\}$ with $p$ prime:

$$
h_{a, b}(x) = ((a x + b) \mod p) \mod m, \quad a \in \{1, \ldots, p-1\}, \; b \in \{0, \ldots, p-1\}.
$$

You can prove universality directly by counting: for fixed $x \neq y$, the number of $(a, b)$ pairs that collide is $p - 1$ out of $p(p-1)$ total, giving probability $\leq 1/m$ once you account for the $\mod m$ folding.

This is why production hash tables (Python dict, Go map, Rust `HashMap`) randomize the hash seed at process startup — to prevent attackers from constructing collision-bomb inputs that turn $\Theta(1)$ hash operations into $\Theta(n)$.

## Amortized resize analysis

If you knew $n$ in advance, you could pick $m \approx 2n$ once and forget it. In practice $n$ grows; the table needs to **resize**.

**Strategy.** When $\alpha$ exceeds a threshold (say 0.75), allocate a new table of size $2m$, rehash all elements, and discard the old table. Cost of one resize: $\Theta(n)$.

If we charged this $\Theta(n)$ to a single insert, that insert would be $\Theta(n)$ — disastrous. The **amortized analysis** trick: spread the cost across the previous $n / 2$ inserts (each of which contributed to filling the table). Each insert is charged $\Theta(1)$ amortized.

Formally, with the **accounting method**: charge each insert 3 units (1 for the actual insert, 2 to "save up" for future rehashing). When a resize happens, every just-rehashed element has 2 saved units — exactly enough to pay for its rehash.

The accounting method gives **amortized $O(1)$** insert. Worst-case insert is still $\Theta(n)$, but averaged over any sequence of operations starting from an empty table, the per-operation cost is $O(1)$.

## When NOT to use a hash table

Hash tables are not always the right answer. Consider:

- **Range queries.** Hash tables don't preserve order. For "find all keys in $[a, b]$," use a balanced BST or sorted array. The {{FILE:stanford-cs161-big-o-cheatsheet}} compares the asymptotic costs side by side.
- **Adversarial inputs without universal hashing.** A non-randomized table on internet-facing input is a denial-of-service waiting to happen.
- **Tiny $n$.** For $n < 32$ or so, a linear scan of an array beats every hash table — modern CPUs love sequential memory access more than they hate $O(n)$.
- **High-throughput concurrent access.** Open-addressing tables are easier to make lock-free than chained tables, but specialized concurrent data structures (skip lists, lock-free trees) often win.

## Comparison summary

| Operation | Hash table (chaining) | Balanced BST | Sorted array |
|---|---|---|---|
| Insert | $O(1)$ expected | $O(\log n)$ | $O(n)$ |
| Lookup | $O(1)$ expected | $O(\log n)$ | $O(\log n)$ |
| Delete | $O(1)$ expected | $O(\log n)$ | $O(n)$ |
| Min/max | $O(n)$ | $O(\log n)$ | $O(1)$ |
| Range $[a, b]$ | $O(n)$ | $O(\log n + k)$ | $O(\log n + k)$ |
| Successor | $O(n)$ | $O(\log n)$ | $O(\log n)$ |

Choose your data structure based on the operation mix you actually need. Many CS 161 exam problems are deliberately designed so that the "obvious" choice (hash table) loses to a less-obvious choice (BST) because the problem requires range queries.

## Practice

For complementary drilling, the {{QUIZ:cs161-asymptotic-analysis-quiz}} reinforces the Big-O reasoning underlying the load-factor analysis here. Connections to other topics in the course:

- The amortized-analysis style here generalizes — the same technique handles dynamic arrays, splay trees, and Fibonacci heaps (used in {{GUIDE:cs161-graph-algorithms}} for Dijkstra/Prim).
- Universal hashing appears later as a building block in randomized algorithms more broadly.
