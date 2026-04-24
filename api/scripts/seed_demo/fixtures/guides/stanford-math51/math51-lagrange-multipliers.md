---
slug: math51-lagrange-multipliers
course:
  ipeds_id: "243744"
  department: "MATH"
  number: "51"
title: "Lagrange Multipliers — MATH 51"
description: "Constrained optimization via the gradient-alignment condition, with worked examples on circles, planes, and surfaces."
tags: ["multivariable-calculus", "optimization", "lagrange-multipliers", "constraints", "final"]
author_role: bot
quiz_slug: math51-lagrange-multipliers-quiz
attached_files:
  - stanford-math51-lagrange-multipliers-worksheet
  - stanford-math51-gradient-chain-rule
attached_resources: []
---

# Lagrange Multipliers

Suppose you want to maximize or minimize a function $f(x, y, z)$
subject to a constraint $g(x, y, z) = c$. The unconstrained recipe
("set $\nabla f = 0$") fails: the optimum doesn't have to be a
critical point of $f$ — it just has to be a critical point of $f$
*restricted to the constraint surface*. **Lagrange multipliers** are
the tool for that.

This guide assumes you're comfortable with the gradient and the
geometric "$\nabla f$ is perpendicular to level sets" picture from
{{GUIDE:math51-gradients-and-chain-rule}}.

## The geometric idea

Picture the level sets of $f$ as a family of curves (or surfaces) and
the constraint $g = c$ as one specific curve. As you walk along
$g = c$, the value of $f$ rises and falls. At an extremum, the level
set of $f$ at that height is **tangent** to the constraint curve.

Tangency is captured by the gradients pointing in the same direction
(possibly with opposite sign or scaled). So at an extremum,

$$ \nabla f(\mathbf{p}) \;=\; \lambda \, \nabla g(\mathbf{p}) $$

for some scalar $\lambda$ — the **Lagrange multiplier**.

If $\nabla f$ and $\nabla g$ weren't parallel, we could nudge along
$g = c$ in a direction with a nonzero component of $\nabla f$ — and $f$
would change. So $\mathbf{p}$ couldn't have been an extremum. That's
the entire intuition.

## The recipe

To find candidate extrema of $f$ subject to $g = c$:

1. Compute $\nabla f$ and $\nabla g$.
2. Set up the system
   $$ \nabla f = \lambda \nabla g, \qquad g(x_1, \ldots, x_n) = c. $$
3. Solve for $(x_1, \ldots, x_n, \lambda)$.
4. Evaluate $f$ at each candidate point. Compare values.
5. Don't forget boundary or non-smooth cases — Lagrange only finds
   *critical points* of the restricted function.

The system in step 2 has $n + 1$ equations in $n + 1$ unknowns.

For practice problems with full solutions, see
{{FILE:stanford-math51-lagrange-multipliers-worksheet}}.

## Worked example: rectangle inscribed in an ellipse

Maximize $A(x, y) = 4 x y$ (the area of an axis-aligned rectangle with
corner at $(x, y)$) subject to
$\frac{x^2}{a^2} + \frac{y^2}{b^2} = 1$ with $x, y > 0$.

Let $g(x, y) = \frac{x^2}{a^2} + \frac{y^2}{b^2}$.

$$ \nabla A = (4 y,\; 4 x), \qquad \nabla g = \left(\frac{2 x}{a^2},\; \frac{2 y}{b^2}\right). $$

The Lagrange condition $\nabla A = \lambda \nabla g$ gives

$$ 4 y = \lambda \cdot \frac{2 x}{a^2}, \qquad 4 x = \lambda \cdot \frac{2 y}{b^2}. $$

Solve the first for $\lambda$: $\lambda = \frac{2 a^2 y}{x}$. Substitute
into the second:

$$ 4 x = \frac{2 a^2 y}{x} \cdot \frac{2 y}{b^2} \;\Longrightarrow\; x^2 b^2 = a^2 y^2 \;\Longrightarrow\; \frac{x^2}{a^2} = \frac{y^2}{b^2}. $$

Combine with the constraint $\frac{x^2}{a^2} + \frac{y^2}{b^2} = 1$:
$\frac{x^2}{a^2} = \frac{y^2}{b^2} = \frac{1}{2}$. So
$x = a/\sqrt 2$, $y = b/\sqrt 2$.

Maximum area: $4 \cdot \frac{a}{\sqrt 2} \cdot \frac{b}{\sqrt 2} = 2 a b$.

## When the recipe needs care

### Multiple constraints

For two constraints $g_1 = c_1$, $g_2 = c_2$ in $\mathbb{R}^3$, the
constrained set is a curve and the gradient condition is

$$ \nabla f \;=\; \lambda_1 \nabla g_1 + \lambda_2 \nabla g_2. $$

You now have two multipliers. Geometrically, $\nabla f$ has to lie in
the plane spanned by $\nabla g_1$ and $\nabla g_2$ — i.e. it has zero
component along the constraint curve.

### Non-smooth constraints

Lagrange's theorem requires $\nabla g \neq 0$ at the extremum. If the
constraint surface has a corner or cusp where $\nabla g = 0$, you have
to check that point separately by inspection.

### Compactness

If the constraint set is compact (closed and bounded), $f$ continuous
guarantees a min and a max — comparing values at the Lagrange
candidates gives the answer. If it's not compact, $f$ might fail to
attain an extremum at all; you may need to study the behaviour at
infinity.

## A common student error

A frequent slip is dividing both sides of $\nabla f = \lambda \nabla g$
by a coordinate that might be zero. If a coordinate of $\nabla g$ is
zero, dividing by it can erase a real solution. Better: solve by
substitution, or treat each coordinate equation as its own constraint
and check both branches.

A related slip: forgetting $\lambda$ entirely after solving and
reporting it as part of the answer. The multiplier is *not* part of
the optimum point — it's a parameter that drops out. (Though its sign
sometimes carries economic meaning — "shadow price.")

## Practice

Take {{QUIZ:math51-lagrange-multipliers-quiz}} once you can run the
recipe end-to-end on the ellipse example without looking. For the
gradient prerequisites in one place, revisit
{{GUIDE:math51-gradients-and-chain-rule}}.

For broader course context, see {{COURSE:stanford/math51}}.
