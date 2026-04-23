---
slug: math51-matrix-operations
course:
  ipeds_id: "243744"
  department: "MATH"
  number: "51"
title: "Matrix Operations — MATH 51"
description: "Multiplication, inverse, determinant, and how matrices act as linear maps."
tags: ["linear-algebra", "matrices", "determinant", "inverse", "midterm"]
author_role: bot
quiz_slug: math51-matrix-operations-quiz
attached_files:
  - stanford-math51-matrix-operations-cheatsheet
attached_resources: []
---

# Matrix Operations

A matrix is a rectangular grid of numbers, but it is more useful to
think of it as a **linear map** — a function that eats a vector and
spits out another vector. Every algebraic property in this guide
(multiplication, inverse, determinant) has a geometric meaning in that
"maps act on vectors" picture, and the test problems you'll see in
MATH 51 reward switching between the two views.

If vectors and dot products are a little hazy, take a quick detour
through {{GUIDE:math51-vectors-and-dot-product}} first; matrix–vector
multiplication is just a stack of dot products in disguise.

## Notation and shapes

We write $A \in \mathbb{R}^{m \times n}$ for a matrix with $m$ rows and
$n$ columns. The entry in row $i$ and column $j$ is $A_{ij}$. A column
vector in $\mathbb{R}^n$ is the same as an $n \times 1$ matrix.

The *shape rule* for any matrix product is non-negotiable: an
$m \times n$ matrix can multiply an $n \times p$ matrix, and the
result is $m \times p$. The two inner dimensions must agree, and they
"cancel."

$$ \underbrace{(m \times n)}_{A} \cdot \underbrace{(n \times p)}_{B} \;=\; \underbrace{(m \times p)}_{AB}. $$

If the inner dimensions don't match, the product is **undefined** —
this is the single most common source of "did I set this up right?"
panic on exams.

## Matrix multiplication, three ways

The same number $C_{ij}$ is the answer to all three of these
descriptions; pick the one that matches what the problem is asking.

**Entry view.** $C_{ij} = \sum_k A_{ik} B_{kj}$ — the dot product of
row $i$ of $A$ with column $j$ of $B$.

**Column view.** Column $j$ of $AB$ is $A$ times column $j$ of $B$:
$(AB)_{:,j} = A \, B_{:,j}$. So multiplying by $A$ on the left
transforms each column of $B$ in the same way.

**Row view.** Row $i$ of $AB$ is row $i$ of $A$ times $B$.

Worked example. With

$$ A = \begin{pmatrix} 1 & 2 \\ 0 & 1 \end{pmatrix}, \qquad B = \begin{pmatrix} 3 & 1 \\ 4 & 0 \end{pmatrix}, $$

the product is

$$ AB = \begin{pmatrix} 1\cdot 3 + 2\cdot 4 & 1\cdot 1 + 2\cdot 0 \\ 0\cdot 3 + 1\cdot 4 & 0\cdot 1 + 1\cdot 0 \end{pmatrix} = \begin{pmatrix} 11 & 1 \\ 4 & 0 \end{pmatrix}. $$

Importantly, $AB \neq BA$ in general. Matrix multiplication is **not**
commutative. It *is* associative ($A(BC) = (AB)C$) and distributive,
which is what lets us reorder factors when the inner-dimension rules
allow it.

For a one-page summary of the multiplication, inverse, and determinant
formulas, keep {{FILE:stanford-math51-matrix-operations-cheatsheet}}
nearby while you practice.

## The identity matrix

The identity matrix $I_n \in \mathbb{R}^{n \times n}$ has $1$s on the
main diagonal and $0$s elsewhere. It is the multiplicative identity:
$IA = AI = A$ for any compatible $A$. It plays the role of the number
$1$ in the matrix world.

## Inverse

A square matrix $A$ is **invertible** if there exists a square matrix
$A^{-1}$ such that $A A^{-1} = A^{-1} A = I$. The inverse is unique
when it exists.

Two essential facts:

- $A$ is invertible $\Longleftrightarrow$ $\det A \neq 0$.
- $(AB)^{-1} = B^{-1} A^{-1}$ (note the swap — "socks and shoes").

For a $2\times 2$ matrix $A = \begin{pmatrix} a & b \\ c & d \end{pmatrix}$ with $\det A = ad - bc \neq 0$,

$$ A^{-1} = \frac{1}{ad - bc} \begin{pmatrix} d & -b \\ -c & a \end{pmatrix}. $$

For larger matrices, you'll usually solve $A x = b$ via row reduction
rather than computing $A^{-1}$ explicitly. Computing the inverse is
expensive; row reduction is $O(n^3)$ but with a much smaller
constant, and it's numerically friendlier.

## Determinant

The **determinant** of a square matrix $A$ is a scalar that captures
how $A$ scales (signed) volume.

- $|\det A|$ = how much $A$ stretches volume.
- $\operatorname{sign}(\det A)$ = whether $A$ flips orientation
  (negative) or preserves it (positive).
- $\det A = 0$ means $A$ collapses some direction to zero — i.e., $A$
  is **singular** and not invertible.

Key properties:

1. $\det(AB) = (\det A)(\det B)$.
2. $\det(A^T) = \det A$.
3. $\det(A^{-1}) = 1 / \det A$ when $A$ is invertible.
4. Swapping two rows multiplies the determinant by $-1$.
5. Multiplying a row by $c$ multiplies the determinant by $c$.
6. Adding a multiple of one row to another leaves the determinant
   unchanged.

Properties 4–6 are exactly the elementary row operations, which is
why row reduction is the standard tool for evaluating large
determinants.

For a $3 \times 3$ matrix you can either expand along a row/column
(cofactor expansion) or use the Sarrus diagonal trick. For
$4 \times 4$ and beyond, row-reduce to upper triangular and multiply
the diagonal — much faster than blindly expanding.

## Transpose

$A^T$ swaps rows and columns: $(A^T)_{ij} = A_{ji}$. Useful identities
to keep in muscle memory:

- $(A^T)^T = A$
- $(A + B)^T = A^T + B^T$
- $(A B)^T = B^T A^T$ (another swap, like the inverse)

## Why this matters geometrically

Every $A \in \mathbb{R}^{n \times n}$ defines a linear map
$x \mapsto A x$. The columns of $A$ are the images of the standard
basis vectors $e_1, \ldots, e_n$. Knowing the columns *is* knowing the
map. The determinant tells you the signed volume of the image of the
unit cube under that map. The inverse undoes the map (when it exists).

Once this picture clicks, half the algebra problems in the course
become "what does the geometry say?" with the algebra serving as a
calculator.

## Common slips on exams

- Multiplying two matrices whose inner dimensions don't match. Always
  write the shapes next to your matrices.
- Forgetting the swap in $(AB)^{-1} = B^{-1}A^{-1}$.
- Computing $\det$ by expanding a giant matrix when row reduction
  would have taken a third of the time.
- Treating $A^{-1}$ as $1/A$ — there is no division of matrices.

## When you're ready

Take {{QUIZ:math51-matrix-operations-quiz}}. Then move on to
{{GUIDE:math51-linear-systems}}, where row reduction gets formalized
into Gaussian elimination.
