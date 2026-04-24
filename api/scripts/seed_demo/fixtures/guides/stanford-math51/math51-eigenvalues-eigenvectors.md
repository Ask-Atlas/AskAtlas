---
slug: math51-eigenvalues-eigenvectors
course:
  ipeds_id: "243744"
  department: "MATH"
  number: "51"
title: "Eigenvalues and Eigenvectors — MATH 51"
description: "The characteristic polynomial, eigenspaces, and what eigenvalues say about a linear map."
tags: ["linear-algebra", "eigenvalues", "eigenvectors", "diagonalization", "midterm"]
author_role: bot
quiz_slug: math51-eigenvalues-eigenvectors-quiz
attached_files:
  - stanford-math51-eigenvalues-slides
  - stanford-math51-matrix-operations-cheatsheet
attached_resources: []
---

# Eigenvalues and Eigenvectors

For a linear map $A : \mathbb{R}^n \to \mathbb{R}^n$, most input
vectors get rotated and stretched in some complicated way. But a few
special directions are only *stretched* — they don't change direction
at all. Those directions are **eigenvectors**, and the stretch factors
are **eigenvalues**. Once you find them, the matrix often becomes
trivial to compute with.

A quick refresher on matrix arithmetic and determinants is in
{{GUIDE:math51-matrix-operations}}; this guide assumes both.

## The defining equation

A nonzero vector $\mathbf{v} \in \mathbb{R}^n$ is an **eigenvector**
of $A$ with **eigenvalue** $\lambda \in \mathbb{R}$ when

$$ A \mathbf{v} = \lambda \mathbf{v}. $$

Equivalently, $(A - \lambda I) \mathbf{v} = \mathbf{0}$.

The "$\mathbf{v} \neq \mathbf{0}$" requirement is essential — if we
allowed the zero vector, every $\lambda$ would qualify and the concept
would be empty.

## Finding eigenvalues

For $\mathbf{v} \neq \mathbf{0}$ to satisfy $(A - \lambda I)\mathbf{v} = \mathbf{0}$,
the matrix $A - \lambda I$ must be singular. So:

$$ \det(A - \lambda I) = 0. $$

This is the **characteristic equation**. Expanded out,
$\det(A - \lambda I)$ is the **characteristic polynomial**
$p_A(\lambda)$, a degree-$n$ polynomial in $\lambda$. Its roots are
the eigenvalues.

## Worked $2\times 2$ example

Let

$$ A = \begin{pmatrix} 4 & -2 \\ 1 & 1 \end{pmatrix}. $$

Then

$$ \det(A - \lambda I) = \det\begin{pmatrix} 4 - \lambda & -2 \\ 1 & 1 - \lambda \end{pmatrix} = (4 - \lambda)(1 - \lambda) - (-2)(1). $$

Expanding: $(4 - \lambda)(1 - \lambda) + 2 = \lambda^2 - 5\lambda + 6$.
Roots: $\lambda = 2$ and $\lambda = 3$.

For each eigenvalue, find a nonzero $\mathbf{v}$ in the null space of
$A - \lambda I$:

- $\lambda = 2$: $A - 2I = \begin{pmatrix} 2 & -2 \\ 1 & -1 \end{pmatrix}$. Null space is spanned by $\mathbf{v}_1 = (1, 1)$.
- $\lambda = 3$: $A - 3I = \begin{pmatrix} 1 & -2 \\ 1 & -2 \end{pmatrix}$. Null space is spanned by $\mathbf{v}_2 = (2, 1)$.

Check: $A \mathbf{v}_1 = (4-2, 1+1) = (2, 2) = 2 \mathbf{v}_1$. ✓

## Eigenspaces

For a fixed eigenvalue $\lambda$, the **eigenspace**
$E_\lambda = \operatorname{Null}(A - \lambda I)$ is a subspace of
$\mathbb{R}^n$. Its dimension is the **geometric multiplicity** of
$\lambda$.

The number of times $\lambda$ appears as a root of $p_A$ is the
**algebraic multiplicity**.

Two facts:

1. Geometric multiplicity is at least $1$ and at most algebraic
   multiplicity.
2. If they're equal *for every eigenvalue*, $A$ is **diagonalizable**.

## Diagonalization

If $A$ has $n$ linearly independent eigenvectors, stack them as
columns of a matrix $P$ and put the eigenvalues on the diagonal of
$D$:

$$ A = P D P^{-1}. $$

This is called **eigendecomposition**. Powers become trivial:

$$ A^k = P D^k P^{-1}, $$

and $D^k$ is just the eigenvalues raised to the $k$th power.

In the $2\times 2$ example above:

$$ P = \begin{pmatrix} 1 & 2 \\ 1 & 1 \end{pmatrix}, \quad D = \begin{pmatrix} 2 & 0 \\ 0 & 3 \end{pmatrix}. $$

If you want a one-page reminder of the formulas, the lecture deck is
{{FILE:stanford-math51-eigenvalues-slides}}.

## Why eigenvalues matter

Eigenvalues encode geometric and dynamic information about $A$:

- **Stretch factors.** $|\lambda|$ tells you how much $A$ stretches
  along $\mathbf{v}$.
- **Stability.** If you iterate $A$ on a vector and all $|\lambda| < 1$,
  the iterates collapse to $\mathbf{0}$. If any $|\lambda| > 1$, they
  blow up.
- **Determinant.** $\det A = \lambda_1 \lambda_2 \cdots \lambda_n$
  (with multiplicity). $A$ is invertible iff none of the $\lambda_i$
  are zero.
- **Trace.** $\operatorname{tr}(A) = \sum_i A_{ii} = \sum_i \lambda_i$.
  This is a quick sanity check on your eigenvalues.

## When eigenvectors fail to span

A famous trap matrix:

$$ N = \begin{pmatrix} 0 & 1 \\ 0 & 0 \end{pmatrix}. $$

Characteristic polynomial: $\lambda^2$. So $\lambda = 0$ with algebraic
multiplicity $2$. But $\operatorname{Null}(N) = \operatorname{span}\{(1, 0)\}$,
geometric multiplicity $1$. Not diagonalizable.

You're not expected to perform Jordan decomposition in MATH 51, but
you should recognise that $A$ being non-diagonalizable is *the*
phenomenon, not a glitch.

## Pre-quiz checklist

- Can you set up $\det(A - \lambda I)$ for a $2\times 2$ or $3\times 3$
  matrix without copy-paste errors?
- Given an eigenvalue, can you write down the corresponding eigenspace?
- Do you know which two scalars (trace, determinant) you can use to
  quickly check your eigenvalues?

If yes, take {{QUIZ:math51-eigenvalues-eigenvectors-quiz}}. Then move
on to the multivariable side of the course in
{{GUIDE:math51-gradients-and-chain-rule}}.
