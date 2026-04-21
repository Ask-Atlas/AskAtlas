---
slug: math51-linear-systems
course:
  ipeds_id: "243744"
  department: "MATH"
  number: "51"
title: "Linear Systems and Gaussian Elimination — MATH 51"
description: "Row reduction, RREF, pivots, and reading off the full solution set from an augmented matrix."
tags: ["linear-algebra", "linear-systems", "row-reduction", "rref", "midterm"]
author_role: bot
attached_files:
  - stanford-math51-matrix-operations-cheatsheet
attached_resources: []
---

# Linear Systems and Gaussian Elimination

Almost every concrete computation in MATH 51 — finding null spaces,
checking linear independence, inverting a matrix, solving for unknown
coefficients — is secretly the same problem: solve a linear system
$A x = b$. Master row reduction once and you collect dividends for the
rest of the course.

This guide assumes you can multiply matrices fluently. If not, work
through {{GUIDE:math51-matrix-operations}} first and come back.

## What "linear system" means

A linear system is a finite collection of equations of the form

$$ \begin{aligned} a_{11} x_1 + a_{12} x_2 + \cdots + a_{1n} x_n &= b_1, \\ a_{21} x_1 + a_{22} x_2 + \cdots + a_{2n} x_n &= b_2, \\ &\;\;\vdots \\ a_{m1} x_1 + a_{m2} x_2 + \cdots + a_{mn} x_n &= b_m. \end{aligned} $$

In matrix form this is just $A x = b$, where $A \in \mathbb{R}^{m \times n}$,
$x \in \mathbb{R}^n$, $b \in \mathbb{R}^m$.

The **augmented matrix** $[A \mid b]$ glues $A$ and $b$ side by side.
All of Gaussian elimination is done on this object.

## The three elementary row operations

Each of these preserves the *solution set* of the linear system:

1. **Swap** two rows.
2. **Scale** a row by a nonzero constant.
3. **Replace** row $R_i$ with $R_i + c R_j$ for $j \neq i$.

Two augmented matrices that differ by a sequence of these operations
are called **row equivalent** — they encode the same solution set.

## Row Echelon Form (REF)

A matrix is in **row echelon form** if:

- All all-zero rows are at the bottom.
- Each leading nonzero entry (a **pivot**) is strictly to the right
  of the pivot in the row above.

REF is enough to read off whether the system has a solution and to
back-substitute. **Reduced** row echelon form (RREF) goes further:

- Every pivot is $1$.
- Every column containing a pivot has zeros in all other entries.

RREF is unique for a given matrix, which makes it the canonical form.

## The procedure (top-down, then bottom-up)

1. Find the leftmost column with a nonzero entry. Swap rows so the
   nonzero entry is on top.
2. Scale that row so the pivot is $1$.
3. Use the pivot to clear all entries below it (row replacement).
4. Move down one row and to the right of the pivot column. Repeat.
5. Once you've reached REF, sweep upward: clear entries above each
   pivot too. You're now in RREF.

A worked $3 \times 4$ example. Start with

$$ \left[\begin{array}{ccc|c} 1 & 2 & 1 & 4 \\ 2 & 5 & 3 & 11 \\ 1 & 3 & 3 & 9 \end{array}\right]. $$

Replace $R_2 \leftarrow R_2 - 2 R_1$ and $R_3 \leftarrow R_3 - R_1$:

$$ \left[\begin{array}{ccc|c} 1 & 2 & 1 & 4 \\ 0 & 1 & 1 & 3 \\ 0 & 1 & 2 & 5 \end{array}\right]. $$

Replace $R_3 \leftarrow R_3 - R_2$:

$$ \left[\begin{array}{ccc|c} 1 & 2 & 1 & 4 \\ 0 & 1 & 1 & 3 \\ 0 & 0 & 1 & 2 \end{array}\right]. $$

Now sweep up. $R_2 \leftarrow R_2 - R_3$, then
$R_1 \leftarrow R_1 - R_3$, then $R_1 \leftarrow R_1 - 2 R_2$:

$$ \left[\begin{array}{ccc|c} 1 & 0 & 0 & 1 \\ 0 & 1 & 0 & 1 \\ 0 & 0 & 1 & 2 \end{array}\right]. $$

Read the solution: $x_1 = 1$, $x_2 = 1$, $x_3 = 2$.

## Reading the answer off RREF

Once you're in RREF, classify the columns:

- **Pivot columns** — variables that get a unique value.
- **Non-pivot columns** — variables that become *free parameters*.

Three outcomes are possible:

1. **Inconsistent.** A row of the form $[0\;0\;\cdots\;0 \mid c]$ with
   $c \neq 0$ appears. No solutions exist.
2. **Unique solution.** Every column of $A$ is a pivot column, and the
   system is consistent.
3. **Infinitely many solutions.** The system is consistent and at
   least one column of $A$ has no pivot. Free variables parametrize a
   line, plane, or higher-dimensional flat of solutions.

## Why pivots count

The number of pivots of $A$ is the **rank** of $A$. Rank tells you
both the dimension of the column space and the number of constraints
the system actually imposes (the rest of the rows are redundant).

A formula you'll use constantly: for $A \in \mathbb{R}^{m \times n}$,

$$ \dim(\operatorname{Null}(A)) \;=\; n - \operatorname{rank}(A). $$

This is the **rank-nullity theorem**, and it is just bookkeeping on
pivot columns vs free columns.

For a quick formula refresher, glance at
{{FILE:stanford-math51-matrix-operations-cheatsheet}}.

## Common procedural slips

- Forgetting to scale a row before using it to clear another column.
  Your pivots don't *have* to be $1$ in REF, but they must be in RREF.
- Doing an "operation" that secretly changes the solution set, e.g.
  $R_i \leftarrow R_i + c R_i$ (this scales by $1+c$, not what you
  meant).
- Stopping at REF and trying to read off the answer directly. RREF is
  what makes it trivial.
- Mis-handling free variables — they are real solutions, not "leftover
  noise."

## Practice

There is no quiz attached to this guide; the row-reduction skill is
tested implicitly in the matrix and eigenvalue quizzes
({{QUIZ:math51-matrix-operations-quiz}} and
{{QUIZ:math51-eigenvalues-eigenvectors-quiz}}) where you'll need to
solve linear systems en route to the answer.

The next conceptual step is eigenvalues — see
{{GUIDE:math51-eigenvalues-eigenvectors}}.
