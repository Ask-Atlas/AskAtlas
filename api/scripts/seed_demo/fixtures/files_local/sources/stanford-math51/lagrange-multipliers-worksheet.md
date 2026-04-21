---
slug: stanford-math51-lagrange-multipliers-worksheet
title: "Lagrange Multipliers Worksheet"
mime: application/vnd.openxmlformats-officedocument.wordprocessingml.document
filename: lagrange-multipliers-worksheet.docx
course: stanford/math51
description: "Student worksheet on constrained optimization using Lagrange multipliers, with guided problems and space for work."
author_role: bot
---

# Lagrange Multipliers Worksheet

**Name:** __________________________  **Section:** ________

## Learning Goals

By the end of this worksheet you should be able to:

1. Set up the Lagrange system for a constraint $g(x) = c$.
2. Solve $\nabla f = \lambda \nabla g$ for critical points.
3. Classify candidates as constrained maxima, minima, or saddle points.
4. Handle two-constraint problems using $\nabla f = \lambda \nabla g + \mu \nabla h$.

## Core Method

To optimize $f(x, y, z)$ subject to $g(x, y, z) = c$:

1. Compute $\nabla f$ and $\nabla g$.
2. Solve the system
   $$\nabla f = \lambda \nabla g, \quad g(x, y, z) = c.$$
3. Evaluate $f$ at each candidate.
4. Compare values; on a compact constraint set the extremes are guaranteed.

## Problem 1 — Rectangle on an Ellipse

Maximize $f(x, y) = xy$ subject to $\frac{x^2}{4} + \frac{y^2}{9} = 1$.

**(a)** Write $\nabla f$ and $\nabla g$.

$\nabla f = \underline{\hspace{4cm}}$

$\nabla g = \underline{\hspace{4cm}}$

**(b)** Solve the Lagrange system. Show all work:

_Work space:_

&nbsp;

&nbsp;

&nbsp;

**(c)** What is the maximum value of $f$?  $f_{\max} = \underline{\hspace{2cm}}$

## Problem 2 — Closest Point to a Plane

Find the point on the plane $x + 2y + 3z = 6$ closest to the origin.

_Hint:_ Minimize $f(x, y, z) = x^2 + y^2 + z^2$ (squared distance) under the plane constraint. The squared-distance objective has the same extremizers as the distance itself but easier derivatives.

**(a)** Set up $\nabla f = \lambda \nabla g$:

$[2x, 2y, 2z] = \lambda [\underline{\hspace{2cm}}]$

**(b)** Solve for $(x, y, z)$:

&nbsp;

&nbsp;

**(c)** Compute the minimum distance:  $d_{\min} = \underline{\hspace{2cm}}$

## Problem 3 — Two Constraints

Maximize $f(x, y, z) = x + y + z$ subject to
$$x^2 + y^2 = 2, \quad z = x + y.$$

Use $\nabla f = \lambda \nabla g + \mu \nabla h$. Identify $g$ and $h$ first.

_Work space:_

&nbsp;

&nbsp;

&nbsp;

## Problem 4 — Conceptual

**(a)** In one sentence, explain why $\nabla f$ and $\nabla g$ must be parallel at a constrained extremum.

_Answer:_ ________________________________________________________________

**(b)** What does $\lambda$ represent geometrically or physically? Give one example from economics or physics.

_Answer:_ ________________________________________________________________

## Self-Check Table

| Step | Did you do it? |
|---|---|
| Computed $\nabla f$ and $\nabla g$ correctly | [ ] |
| Wrote all Lagrange equations | [ ] |
| Used the constraint equation explicitly | [ ] |
| Compared $f$ at every candidate point | [ ] |
| Stated the answer with units / context | [ ] |

_Turn in at the end of section. Staple any extra work._
