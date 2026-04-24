---
slug: stanford-math51-gradient-chain-rule
title: "Gradient and the Multivariable Chain Rule"
mime: application/pdf
filename: gradient-chain-rule.pdf
course: stanford/math51
description: "Definitions and working rules for the gradient, directional derivatives, and the multivariable chain rule in MATH 51."
author_role: bot
---

# Gradient and the Multivariable Chain Rule

This note covers the three core objects that organize differential calculus of scalar fields: the gradient $\nabla f$, the directional derivative $D_u f$, and the multivariable chain rule.

## 1. Partial Derivatives

For $f : \mathbb{R}^n \to \mathbb{R}$, the partial derivative with respect to $x_i$ is

$$\partial_{x_i} f(x) = \lim_{h \to 0} \frac{f(x + h e_i) - f(x)}{h}.$$

Each $\partial_{x_i} f$ measures how $f$ changes along one coordinate axis while the other coordinates are held fixed.

## 2. The Gradient

The **gradient** assembles all partials into a vector:

$$\nabla f = [\partial_x f, \partial_y f, \partial_z f]$$

(for a function of three variables; extend the obvious way to $n$ variables). Two geometric facts do most of the work in MATH 51:

- $\nabla f(x_0)$ points in the direction of **steepest ascent** of $f$ at $x_0$.
- $\nabla f(x_0)$ is **orthogonal** to the level set $\{x : f(x) = f(x_0)\}$ at $x_0$.

## 3. Directional Derivative

The derivative of $f$ at $x_0$ in the direction of a **unit** vector $u$ is

$$D_u f(x_0) = \nabla f(x_0) \cdot u = \|\nabla f(x_0)\| \cos\theta,$$

where $\theta$ is the angle between $\nabla f(x_0)$ and $u$. It is maximized when $u$ points along $\nabla f$, giving $\|\nabla f(x_0)\|$.

If you are handed a non-unit vector $v$, normalize first: $u = v / \|v\|$.

## 4. Multivariable Chain Rule

Let $f : \mathbb{R}^n \to \mathbb{R}$ and $r : \mathbb{R} \to \mathbb{R}^n$ be differentiable. Then $g(t) = f(r(t))$ satisfies

$$\frac{dg}{dt} = \nabla f(r(t)) \cdot r'(t).$$

More generally, for $f : \mathbb{R}^n \to \mathbb{R}$ and $x : \mathbb{R}^m \to \mathbb{R}^n$,

$$\frac{\partial f}{\partial t_j} = \sum_{i=1}^n \frac{\partial f}{\partial x_i} \frac{\partial x_i}{\partial t_j}.$$

In matrix form: $D(f \circ x) = Df \cdot Dx$.

## 5. Worked Micro-Example

Let $f(x, y) = x^2 y$ and $r(t) = (\cos t, \sin t)$.

- $\nabla f = [2xy, x^2]$
- $r'(t) = [-\sin t, \cos t]$
- $\frac{d}{dt} f(r(t)) = 2\cos t \sin t \cdot (-\sin t) + \cos^2 t \cdot \cos t$
- Simplifies to $\cos^3 t - 2 \cos t \sin^2 t$.

## Comparison Table

| Object | Type | Input | Output |
|---|---|---|---|
| Partial $\partial_{x_i} f$ | scalar | point in $\mathbb{R}^n$ | scalar |
| Gradient $\nabla f$ | vector field | point in $\mathbb{R}^n$ | vector in $\mathbb{R}^n$ |
| Directional deriv. $D_u f$ | scalar | point + unit vector | scalar |
| Chain rule | rule | composed map | derivative of composition |

## Common Pitfalls

- Forgetting to normalize $u$ when computing $D_u f$.
- Writing $\nabla f$ as a row when the problem expects a column (stay consistent).
- Confusing $\nabla f$ (scalar field input) with the Jacobian (vector field input).
