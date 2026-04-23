---
slug: math51-vectors-and-dot-product
course:
  ipeds_id: "243744"
  department: "MATH"
  number: "51"
title: "Vectors and the Dot Product — MATH 51"
description: "Vector arithmetic, geometric meaning of the dot product, projections, and the angle formula."
tags: ["linear-algebra", "vectors", "dot-product", "geometry", "week-1"]
author_role: bot
attached_files:
  - stanford-math51-notation-glossary
attached_resources: []
---

# Vectors and the Dot Product

MATH 51 starts with vectors because every later topic — linear systems,
matrices, gradients, optimization — is really an operation on vectors in
disguise. If your intuition for what a vector *is* and what the dot
product *means* is solid, the rest of the course mostly follows. If it
isn't, you will spend the second half of the quarter chasing symbols
without geometry.

Throughout this guide we work in $\mathbb{R}^n$, the set of ordered
$n$-tuples of real numbers. A typical vector looks like
$\mathbf{v} = (v_1, v_2, \ldots, v_n)$, and we'll write column vectors
when we need them to interact with matrices.

## What a vector actually is

A vector has two equally valid pictures, and you should be fluent in
switching between them.

1. **Algebraic picture.** A vector is an ordered list of numbers. We
   add componentwise and scale componentwise:
   $$ (a_1, a_2) + (b_1, b_2) = (a_1 + b_1,\, a_2 + b_2), \qquad c\,(a_1, a_2) = (c a_1,\, c a_2). $$
2. **Geometric picture.** A vector is a directed arrow with a tail and
   a head. Addition is "tip-to-tail." Scaling stretches or flips the
   arrow.

Both pictures describe the same object. Pick the one that makes the
problem easiest, then translate when the other picture would be quicker.

For notation conventions used in this course (bold for vectors, hats for
unit vectors, etc.), see the {{FILE:stanford-math51-notation-glossary}}.

## Length

The **length** (or **norm**) of $\mathbf{v} = (v_1, \ldots, v_n)$ is

$$ \|\mathbf{v}\| = \sqrt{v_1^2 + v_2^2 + \cdots + v_n^2}. $$

This is just the Pythagorean theorem in $n$ dimensions. A **unit
vector** has length $1$. To turn any nonzero $\mathbf{v}$ into a unit
vector, divide by its length: $\hat{\mathbf{v}} = \mathbf{v} / \|\mathbf{v}\|$.

Worked example: $\mathbf{v} = (3, 4)$ has length
$\sqrt{9 + 16} = 5$, so $\hat{\mathbf{v}} = (3/5, 4/5)$.

## The dot product — definition

For $\mathbf{a} = (a_1, \ldots, a_n)$ and $\mathbf{b} = (b_1, \ldots, b_n)$,

$$ \mathbf{a} \cdot \mathbf{b} = a_1 b_1 + a_2 b_2 + \cdots + a_n b_n. $$

The output is a *scalar*, not a vector. That single fact dissolves a
remarkable number of student mistakes.

The dot product satisfies:

- **Symmetry:** $\mathbf{a} \cdot \mathbf{b} = \mathbf{b} \cdot \mathbf{a}$
- **Linearity in each slot:** $(c \mathbf{a} + \mathbf{a}') \cdot \mathbf{b} = c (\mathbf{a} \cdot \mathbf{b}) + (\mathbf{a}' \cdot \mathbf{b})$
- **Positive definiteness:** $\mathbf{a} \cdot \mathbf{a} = \|\mathbf{a}\|^2 \geq 0$, with equality only when $\mathbf{a} = \mathbf{0}$.

## The geometric formula

The dot product also equals

$$ \mathbf{a} \cdot \mathbf{b} = \|\mathbf{a}\| \, \|\mathbf{b}\| \, \cos\theta $$

where $\theta$ is the angle between the two vectors. This is the
formula you reach for whenever a problem mentions "angle,"
"perpendicular," or "projection."

Three immediate consequences:

- $\mathbf{a} \cdot \mathbf{b} = 0$ iff $\mathbf{a}$ and $\mathbf{b}$
  are **orthogonal** (perpendicular). This is the single most-used
  fact in the course.
- The sign of $\mathbf{a} \cdot \mathbf{b}$ tells you whether the
  angle is acute (positive), right (zero), or obtuse (negative).
- Solving for the angle: $\cos\theta = (\mathbf{a} \cdot \mathbf{b}) / (\|\mathbf{a}\| \|\mathbf{b}\|)$.

## Worked angle problem

Find the angle between $\mathbf{a} = (1, 2, 2)$ and $\mathbf{b} = (2, 0, 1)$.

1. $\mathbf{a} \cdot \mathbf{b} = 1\cdot 2 + 2\cdot 0 + 2\cdot 1 = 4$.
2. $\|\mathbf{a}\| = \sqrt{1 + 4 + 4} = 3$, $\|\mathbf{b}\| = \sqrt{4 + 0 + 1} = \sqrt 5$.
3. $\cos\theta = 4 / (3\sqrt 5) = 4\sqrt 5 / 15$.
4. $\theta = \arccos(4\sqrt 5 / 15) \approx 53.4°$.

## Projection of one vector onto another

The **projection** of $\mathbf{a}$ onto $\mathbf{b}$ is the "shadow" of
$\mathbf{a}$ in the direction of $\mathbf{b}$:

$$ \operatorname{proj}_{\mathbf{b}} \mathbf{a} \;=\; \frac{\mathbf{a} \cdot \mathbf{b}}{\mathbf{b} \cdot \mathbf{b}} \, \mathbf{b}. $$

The scalar $(\mathbf{a} \cdot \mathbf{b})/(\mathbf{b} \cdot \mathbf{b})$
tells you *how much of* $\mathbf{a}$ lies along $\mathbf{b}$. Multiply
by $\mathbf{b}$ to get the actual vector.

Decomposition trick: any $\mathbf{a}$ splits uniquely as

$$ \mathbf{a} = \operatorname{proj}_{\mathbf{b}} \mathbf{a} \;+\; \mathbf{a}_\perp, $$

where $\mathbf{a}_\perp$ is orthogonal to $\mathbf{b}$. This shows up
repeatedly later — in least squares, Gram-Schmidt, and gradient
descent.

## Things that trip students up

- $\mathbf{a} \cdot \mathbf{b}$ is a number; it has no direction.
- $\|\mathbf{a} + \mathbf{b}\| \neq \|\mathbf{a}\| + \|\mathbf{b}\|$ in
  general. (Triangle inequality goes the other way: $\leq$.)
- "Orthogonal" and "linearly independent" are different concepts. Two
  orthogonal nonzero vectors are independent, but independent vectors
  need not be orthogonal.
- The zero vector is orthogonal to *every* vector. This is a
  convention that keeps the formulas tidy; don't fight it.

## Quick checks before moving on

- Can you compute $\|\mathbf{v}\|$ in $\mathbb{R}^4$ without thinking?
- Given two vectors, can you decide in one line whether they're
  perpendicular?
- Can you write down the projection formula from memory?

If yes to all three, the next guide,
{{GUIDE:math51-matrix-operations}}, builds on these ideas to introduce
matrix multiplication as a stack of dot products. You can also test
those skills directly with {{QUIZ:math51-matrix-operations-quiz}}
once you have moved on.

For the broader course context, see {{COURSE:stanford/math51}}.
