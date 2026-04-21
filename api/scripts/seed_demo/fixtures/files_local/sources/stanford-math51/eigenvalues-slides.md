---
slug: stanford-math51-eigenvalues-slides
title: "Eigenvalues, Eigenvectors, and Diagonalization"
mime: application/vnd.openxmlformats-officedocument.presentationml.presentation
filename: eigenvalues-slides.pptx
course: stanford/math51
description: "Slide deck on eigenvalues, eigenvectors, and diagonalization with worked examples and geometric intuition."
author_role: bot
---

# Eigenvalues, Eigenvectors, and Diagonalization

Slide deck for MATH 51, Week 7.

## Slide 1 — Motivation

Why eigenvalues? Because they let us **pick a basis in which a linear map is just a list of scalings**. Every big application — PCA, PageRank, differential equations, vibration modes — is some version of "find the axes on which $A$ acts simply."

## Slide 2 — The Definition

A nonzero vector $v$ is an **eigenvector** of $A \in \mathbb{R}^{n \times n}$ with **eigenvalue** $\lambda$ if

$$A v = \lambda v.$$

Geometrically: $A$ stretches $v$ by $\lambda$ without rotating it off its own line.

## Slide 3 — The Characteristic Polynomial

Rearrange $Av = \lambda v$:

$$(A - \lambda I) v = 0.$$

For a nonzero $v$ to exist, $A - \lambda I$ must be singular:

$$\det(A - \lambda I) = 0.$$

This polynomial in $\lambda$ has degree $n$, so it has $n$ roots in $\mathbb{C}$ (with multiplicity).

## Slide 4 — Finding Eigenvectors

For each root $\lambda_i$, solve

$$(A - \lambda_i I) v = 0$$

by row reduction. The kernel is the **eigenspace** $E_{\lambda_i}$.

## Slide 5 — Worked Example

$$A = \begin{bmatrix} 2 & 1 \\ 0 & 3 \end{bmatrix}$$

Characteristic polynomial: $(2 - \lambda)(3 - \lambda) = 0$, so $\lambda_1 = 2$, $\lambda_2 = 3$.

- $E_2$: solve $\begin{bmatrix} 0 & 1 \\ 0 & 1 \end{bmatrix} v = 0 \Rightarrow v = [1, 0]^T$
- $E_3$: solve $\begin{bmatrix} -1 & 1 \\ 0 & 0 \end{bmatrix} v = 0 \Rightarrow v = [1, 1]^T$

## Slide 6 — Key Identities

- $\text{tr}(A) = \sum_i \lambda_i$
- $\det(A) = \prod_i \lambda_i$
- $A$ and $A^T$ share eigenvalues
- Eigenvalues of $A^k$ are $\lambda_i^k$
- Eigenvalues of $A^{-1}$ are $1/\lambda_i$ (when $A$ is invertible)

## Slide 7 — Diagonalization

$A$ is **diagonalizable** iff $\mathbb{R}^n$ (or $\mathbb{C}^n$) has a basis of eigenvectors of $A$.

Let $P = [v_1 \mid v_2 \mid \dots \mid v_n]$ and $D = \text{diag}(\lambda_1, \dots, \lambda_n)$. Then

$$A = P D P^{-1}.$$

## Slide 8 — Why Diagonalization Is Useful

Powers become trivial:

$$A^k = P D^k P^{-1}, \quad D^k = \text{diag}(\lambda_1^k, \dots, \lambda_n^k).$$

Matrix functions reduce to scalar functions: $e^A = P e^D P^{-1}$, etc.

## Slide 9 — When It Fails

Not every matrix is diagonalizable. Classic counterexample:

$$\begin{bmatrix} 0 & 1 \\ 0 & 0 \end{bmatrix}$$

Only one eigenvalue ($\lambda = 0$) with a one-dimensional eigenspace — not enough to span $\mathbb{R}^2$.

## Slide 10 — Symmetric Matrices

**Spectral theorem:** if $A = A^T$, then $A$ is diagonalizable by an **orthogonal** matrix:

$$A = Q \Lambda Q^T, \quad Q^T Q = I.$$

Eigenvalues are real; eigenvectors from distinct eigenvalues are orthogonal.

## Slide 11 — Comparison Table

| Matrix | Diagonalizable? | Where eigenvectors live |
|---|---|---|
| Generic $n \times n$ | Usually (over $\mathbb{C}$) | $\mathbb{C}^n$ |
| Symmetric real | Always | Orthogonal basis of $\mathbb{R}^n$ |
| Triangular w/ repeated $\lambda$ | Maybe not | Fewer than $n$ independent |
| Rotation in $\mathbb{R}^2$ (non-trivial) | Not over $\mathbb{R}$ | Complex eigenvectors |

## Slide 12 — Takeaways

- Eigenvalues = scalings along special directions.
- Solve $\det(A - \lambda I) = 0$, then the kernel for each $\lambda$.
- Diagonalization turns $A^k$ and $e^A$ into scalar work.
- Symmetric matrices get the cleanest version of the story.
