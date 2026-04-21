---
slug: stanford-math51-matrix-operations-cheatsheet
title: "Matrix Operations Cheatsheet"
mime: application/pdf
filename: matrix-operations-cheatsheet.pdf
course: stanford/math51
description: "Quick-reference sheet for MATH 51 matrix operations: transpose, inverse, rank, determinant, and eigenvalues."
author_role: bot
---

# Matrix Operations Cheatsheet

A compact reference for the core matrix operations in MATH 51. Let $A \in \mathbb{R}^{m \times n}$ unless otherwise noted.

## 1. Transpose

The transpose $A^T$ flips rows and columns: $(A^T)_{ij} = A_{ji}$.

Key identities:

- $(A^T)^T = A$
- $(A + B)^T = A^T + B^T$
- $(AB)^T = B^T A^T$
- $A$ is **symmetric** iff $A = A^T$ (requires $m = n$).

## 2. Inverse

For a square matrix $A \in \mathbb{R}^{n \times n}$, the inverse $A^{-1}$ satisfies $A A^{-1} = A^{-1} A = I_n$. It exists iff $\det(A) \neq 0$.

For a $2 \times 2$ matrix:

$$A = \begin{bmatrix} a & b \\ c & d \end{bmatrix}, \quad A^{-1} = \frac{1}{ad - bc} \begin{bmatrix} d & -b \\ -c & a \end{bmatrix}$$

General strategy: row-reduce $[A \mid I]$ to $[I \mid A^{-1}]$.

## 3. Rank

The rank $\text{rank}(A)$ is the dimension of the column space (equivalently, the row space, or the number of pivots in RREF).

- $\text{rank}(A) \le \min(m, n)$
- $A$ is **full rank** if $\text{rank}(A) = \min(m, n)$
- **Rank-nullity:** $\text{rank}(A) + \dim(\ker A) = n$

## 4. Determinant

Defined only for square $A$. Geometrically, $|\det(A)|$ is the signed volume scaling factor of the linear map $x \mapsto Ax$.

| Property | Statement |
|---|---|
| Row swap | Flips sign of $\det$ |
| Row scale by $k$ | Multiplies $\det$ by $k$ |
| Row replacement | Leaves $\det$ unchanged |
| Product rule | $\det(AB) = \det(A)\det(B)$ |
| Transpose | $\det(A^T) = \det(A)$ |
| Invertibility | $A$ invertible $\iff \det(A) \neq 0$ |

## 5. Eigenvalues

$\lambda \in \mathbb{C}$ is an **eigenvalue** of $A$ if there exists $v \neq 0$ with $Av = \lambda v$. Solve the **characteristic polynomial**:

$$\det(A - \lambda I) = 0$$

Useful facts:

- $\sum_i \lambda_i = \text{tr}(A)$
- $\prod_i \lambda_i = \det(A)$
- Eigenvalues of $A^T$ match those of $A$
- Eigenvalues of a triangular matrix are its diagonal entries

## Comparison Table

| Operation | Input | Output | Cost | Exists when |
|---|---|---|---|---|
| Transpose | $m \times n$ | $n \times m$ | $O(mn)$ | Always |
| Inverse | $n \times n$ | $n \times n$ | $O(n^3)$ | $\det(A) \neq 0$ |
| Rank | $m \times n$ | scalar | $O(mn \cdot \min(m,n))$ | Always |
| Determinant | $n \times n$ | scalar | $O(n^3)$ via LU | $A$ square |
| Eigenvalues | $n \times n$ | $n$ roots | $O(n^3)$ iterative | $A$ square |

Use this sheet alongside worked examples from lecture and practice sets.
