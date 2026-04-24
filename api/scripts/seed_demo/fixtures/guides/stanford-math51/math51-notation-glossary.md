---
slug: math51-notation-glossary
course:
  ipeds_id: "243744"
  department: "MATH"
  number: "51"
title: "Notation Glossary — MATH 51"
description: "Vector, matrix, set, and calculus notation conventions used throughout MATH 51, with disambiguation tips."
tags: ["reference", "notation", "glossary", "vocabulary"]
author_role: bot
attached_files:
  - stanford-math51-notation-glossary
attached_resources: []
---

# MATH 51 Notation Glossary

A surprising number of "I don't understand this problem" moments are
actually "I don't recognize this symbol" moments. This glossary
collects the conventions used in MATH 51 lectures, the textbook, and
the problem sets, with disambiguation tips where conventions clash.

For a printable one-page version, see
{{FILE:stanford-math51-notation-glossary}}.

## Sets and numbers

| Symbol | Meaning |
|---|---|
| $\mathbb{N}$ | natural numbers $\{0, 1, 2, \ldots\}$ (sometimes $\{1, 2, \ldots\}$ — context tells you) |
| $\mathbb{Z}$ | integers |
| $\mathbb{Q}$ | rationals |
| $\mathbb{R}$ | real numbers |
| $\mathbb{R}^n$ | $n$-tuples of real numbers; the standard $n$-dimensional space |
| $\in$ | "is an element of"; e.g. $x \in \mathbb{R}^3$ |
| $\subseteq$ | subset (allows equality) |
| $\subset$ | proper subset (textbook convention varies) |
| $\emptyset$ | empty set |
| $|S|$ | cardinality (size) of the set $S$ |

## Vectors

| Symbol | Meaning |
|---|---|
| $\mathbf{v}$ or $\vec{v}$ | a vector — bold typeset or arrow accent |
| $v_i$ | the $i$-th component of $\mathbf{v}$ |
| $\hat{\mathbf{v}}$ | the unit vector in the direction of $\mathbf{v}$, i.e. $\mathbf{v} / \|\mathbf{v}\|$ |
| $\|\mathbf{v}\|$ | length (norm) of $\mathbf{v}$; $\sqrt{v_1^2 + \cdots + v_n^2}$ |
| $\mathbf{v} \cdot \mathbf{w}$ | dot product, a scalar |
| $\mathbf{v} \times \mathbf{w}$ | cross product (only in $\mathbb{R}^3$) |
| $\mathbf{0}$ | the zero vector — bold to distinguish from scalar $0$ |
| $\mathbf{e}_i$ | standard basis vector with $1$ in slot $i$ and $0$ elsewhere |
| $\operatorname{span}\{\mathbf{v}_1, \ldots, \mathbf{v}_k\}$ | set of all linear combinations $c_1 \mathbf{v}_1 + \cdots + c_k \mathbf{v}_k$ |

## Matrices

| Symbol | Meaning |
|---|---|
| $A$, $B$ | a matrix (capital letter, non-bold) |
| $A_{ij}$ or $a_{ij}$ | entry in row $i$, column $j$ |
| $A^T$ | transpose: rows become columns |
| $A^{-1}$ | inverse, when it exists |
| $\det A$ or $|A|$ | determinant |
| $\operatorname{tr}(A)$ | trace = sum of diagonal entries |
| $\operatorname{rank}(A)$ | dimension of the column space (= number of pivots) |
| $\operatorname{Null}(A)$ | null space (kernel); set of $\mathbf{x}$ with $A \mathbf{x} = \mathbf{0}$ |
| $\operatorname{Col}(A)$ | column space; span of the columns of $A$ |
| $I$ or $I_n$ | identity matrix (size $n$) |
| $[A \mid b]$ | augmented matrix — $A$ on the left, vector $b$ on the right |

A useful disambiguation: $|A|$ for an $n \times n$ matrix means
**determinant**, but for a number $|x|$ means **absolute value**, and
for a vector $\|\mathbf{v}\|$ means **norm**. Two bars vs one bar
matters. When in doubt, write $\det A$.

## Functions and calculus

| Symbol | Meaning |
|---|---|
| $f : \mathbb{R}^n \to \mathbb{R}^m$ | a function from $\mathbb{R}^n$ to $\mathbb{R}^m$ |
| $f_x$ or $\partial f / \partial x$ | partial derivative with respect to $x$ |
| $\nabla f$ | gradient — vector of partial derivatives |
| $D_{\hat{\mathbf{u}}} f$ | directional derivative in the direction of unit vector $\hat{\mathbf{u}}$ |
| $H_f$ | Hessian — matrix of second partials |
| $J_f$ | Jacobian matrix — for $f : \mathbb{R}^n \to \mathbb{R}^m$, the $m \times n$ matrix of first partials |
| $\circ$ | function composition: $(f \circ g)(x) = f(g(x))$ |
| $\Rightarrow$ | "implies" |
| $\iff$ or $\Longleftrightarrow$ | "if and only if" |
| $\forall$ | "for all" |
| $\exists$ | "there exists" |

## Common subscript and superscript conventions

- $\mathbf{v}^{(k)}$ — the $k$-th vector in a sequence (parenthetical
  superscript is a label, *not* a power).
- $A^k$ — $A$ multiplied by itself $k$ times.
- $A^T$ — transpose.
- $A^{-1}$ — inverse.
- $\lambda_i$ — the $i$-th eigenvalue.

## Greek letters that show up constantly

- $\alpha, \beta, \gamma$ — generic scalars or angles.
- $\theta, \varphi$ — angles, especially in polar/spherical coordinates.
- $\lambda$ — eigenvalues, Lagrange multipliers.
- $\mu$ — sometimes a second multiplier.
- $\Sigma$ — summation, also covariance matrices in stats.
- $\Pi$ — product.
- $\Delta$ — change in, finite difference.
- $\nabla$ — "del", the gradient operator.
- $\partial$ — partial derivative symbol; also boundary in topology.

## "Iff" vs "if"

In proofs and definitions, **iff** = "if and only if" — a two-way
implication. Plain **if** is one-way. Pay attention: a definition
phrased "$\mathbf{v}$ is unit if $\|\mathbf{v}\| = 1$" is really
"iff" by convention, but a theorem stated with "if" is a one-way
claim and the converse may be false.

## Common ambiguities

- **Vector vs scalar.** If a quantity is bold or has an arrow, it's a
  vector. If not, scalar. Hand-written work — use arrows so the
  grader can tell.
- **Multiplication.** $\mathbf{v} \cdot \mathbf{w}$ is dot product (scalar
  output); $\mathbf{v} \times \mathbf{w}$ is cross product (vector output);
  $A B$ is matrix multiplication. Don't substitute one for another.
- **Open vs closed intervals.** $(a, b)$ excludes endpoints,
  $[a, b]$ includes them. $(a, b)$ also denotes a 2D point —
  context will tell you which.

## Where to use this

When a problem statement looks unfamiliar, look up each unfamiliar
symbol here first before assuming you missed a concept. Most often
it's just notation drift between lecture, textbook, and homework.

For the next conceptual topic, see {{GUIDE:math51-vectors-and-dot-product}}.

For full course context, see {{COURSE:stanford/math51}}.
