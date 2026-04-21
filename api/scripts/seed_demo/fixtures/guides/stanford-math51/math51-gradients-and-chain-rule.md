---
slug: math51-gradients-and-chain-rule
course:
  ipeds_id: "243744"
  department: "MATH"
  number: "51"
title: "Gradients and the Multivariable Chain Rule — MATH 51"
description: "Partial derivatives, the gradient vector, directional derivatives, and the chain rule for compositions of multivariable functions."
tags: ["multivariable-calculus", "gradient", "chain-rule", "partial-derivatives", "week-6"]
author_role: bot
quiz_slug: math51-gradients-and-chain-rule-quiz
attached_files:
  - stanford-math51-gradient-chain-rule
attached_resources: []
---

# Gradients and the Multivariable Chain Rule

A scalar-valued function $f : \mathbb{R}^n \to \mathbb{R}$ has $n$
"slopes" at each point — one per coordinate direction. Bundle them
into a vector and you get the **gradient**, the single most important
object in multivariable calculus. Nearly every optimization,
constraint, and physics problem in the second half of MATH 51 is the
gradient acting on something.

This guide assumes you can compute a derivative in one variable and
that you know what a vector is (see
{{GUIDE:math51-vectors-and-dot-product}} if not).

## Partial derivatives

For $f(x, y)$, the **partial derivative** with respect to $x$ at
$(x_0, y_0)$ is

$$ \frac{\partial f}{\partial x}(x_0, y_0) \;=\; \lim_{h \to 0} \frac{f(x_0 + h,\, y_0) - f(x_0, y_0)}{h}. $$

Operationally: hold every other variable constant and differentiate as
if $x$ were the only variable.

Example. $f(x, y) = x^2 y + \sin(xy)$.

$$ \frac{\partial f}{\partial x} = 2 x y + y \cos(x y), \qquad \frac{\partial f}{\partial y} = x^2 + x \cos(x y). $$

The notation $f_x$ is shorthand for $\partial f / \partial x$.

## The gradient

For $f : \mathbb{R}^n \to \mathbb{R}$, the **gradient** at $\mathbf{p}$
is the vector of partial derivatives:

$$ \nabla f(\mathbf{p}) \;=\; \left( \frac{\partial f}{\partial x_1}(\mathbf{p}),\; \frac{\partial f}{\partial x_2}(\mathbf{p}),\; \ldots,\; \frac{\partial f}{\partial x_n}(\mathbf{p}) \right). $$

Three properties to memorize:

1. **Direction of steepest ascent.** $\nabla f(\mathbf{p})$ points in
   the direction in which $f$ increases fastest at $\mathbf{p}$.
2. **Magnitude is the rate.** $\|\nabla f(\mathbf{p})\|$ is the
   maximum directional rate of change of $f$ at $\mathbf{p}$.
3. **Perpendicular to level sets.** $\nabla f(\mathbf{p})$ is
   orthogonal to the level set $\{f = f(\mathbf{p})\}$ at $\mathbf{p}$.

That last property is what makes gradients work as **normals** to
surfaces, which we'll exploit in the Lagrange multiplier guide.

## The directional derivative

The rate of change of $f$ at $\mathbf{p}$ in the direction of a unit
vector $\hat{\mathbf{u}}$ is

$$ D_{\hat{\mathbf{u}}} f(\mathbf{p}) \;=\; \nabla f(\mathbf{p}) \cdot \hat{\mathbf{u}}. $$

Two takeaways:

- The maximum is $\|\nabla f(\mathbf{p})\|$, achieved when
  $\hat{\mathbf{u}}$ points along $\nabla f(\mathbf{p})$.
- $D_{\hat{\mathbf{u}}} f = 0$ exactly when $\hat{\mathbf{u}}$ is
  tangent to the level set.

If $\mathbf{u}$ isn't unit length, normalize first:
$\hat{\mathbf{u}} = \mathbf{u} / \|\mathbf{u}\|$.

## Linearization

The first-order Taylor approximation of $f$ near $\mathbf{p}$:

$$ f(\mathbf{p} + \mathbf{h}) \;\approx\; f(\mathbf{p}) + \nabla f(\mathbf{p}) \cdot \mathbf{h}. $$

The gradient *is* the derivative of $f$ in the linear-algebraic sense
— the unique linear map that best approximates $f$ at $\mathbf{p}$.

## The chain rule, two flavours

### Flavour 1: composition with a curve

Let $f : \mathbb{R}^n \to \mathbb{R}$ and let
$\mathbf{r}(t) : \mathbb{R} \to \mathbb{R}^n$ be a smooth curve.
Define $g(t) = f(\mathbf{r}(t))$. Then

$$ g'(t) \;=\; \nabla f(\mathbf{r}(t)) \cdot \mathbf{r}'(t). $$

This is the standard "chain rule for paths." It says: the rate of
change of $f$ along the curve is the gradient of $f$ dotted with the
velocity of the curve.

### Flavour 2: full multivariable chain rule

For $f : \mathbb{R}^n \to \mathbb{R}$ depending on intermediate
variables $u_1, \ldots, u_n$, each of which depends on
$s_1, \ldots, s_m$, the partial of $f$ with respect to $s_j$ is

$$ \frac{\partial f}{\partial s_j} \;=\; \sum_{i=1}^n \frac{\partial f}{\partial u_i} \, \frac{\partial u_i}{\partial s_j}. $$

The intuition: changing $s_j$ propagates through every intermediate
$u_i$ that depends on it; sum the contributions.

In matrix form, if you stack the partials into Jacobians, the chain
rule is just matrix multiplication of those Jacobians.

For a one-page summary of all the gradient-and-chain-rule formulas,
keep {{FILE:stanford-math51-gradient-chain-rule}} open during practice.

## Worked example

Let $f(x, y) = x^2 + 3 y^2$ and $\mathbf{r}(t) = (\cos t,\, \sin t)$.
Find $g'(t)$ where $g(t) = f(\mathbf{r}(t))$.

1. $\nabla f = (2x, 6y)$. Along $\mathbf{r}(t)$: $\nabla f = (2\cos t, 6 \sin t)$.
2. $\mathbf{r}'(t) = (-\sin t, \cos t)$.
3. $g'(t) = (2\cos t)(-\sin t) + (6 \sin t)(\cos t) = 4 \sin t \cos t = 2 \sin(2t)$.

Sanity check: $g(t) = \cos^2 t + 3 \sin^2 t = 1 + 2 \sin^2 t$, so
$g'(t) = 4 \sin t \cos t = 2 \sin(2t)$. ✓

## Common bugs

- Treating partial derivatives as ordinary derivatives without holding
  the other variables constant. The "holding constant" is part of the
  definition.
- Forgetting to normalize when computing a directional derivative.
- Writing $\nabla f$ as a number rather than a vector.
- Confusing the gradient with the total derivative for vector-valued
  $f$ — the gradient is a row/column of the Jacobian; the full
  Jacobian is the matrix.

## Practice

Take {{QUIZ:math51-gradients-and-chain-rule-quiz}}. Once gradients
feel automatic, the constrained-optimization machinery in
{{GUIDE:math51-lagrange-multipliers}} is mostly bookkeeping on top of
them.
