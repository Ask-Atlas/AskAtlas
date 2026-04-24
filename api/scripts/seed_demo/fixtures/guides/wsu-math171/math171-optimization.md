---
slug: math171-optimization
course:
  ipeds_id: "236939"
  department: "MATH"
  number: "171"
title: "Optimization — MATH 171"
description: "Closed-interval method, critical points, the second-derivative test, and worked optimization problems."
tags: ["calculus", "optimization", "critical-points", "applications"]
author_role: bot
quiz_slug: math171-optimization-quiz
attached_files:
  - wsu-math171-optimization-slides
  - wsu-math171-derivative-rules
attached_resources: []
---

# Optimization

Optimization is where derivatives finally start paying their rent. The whole apparatus of $f'(x)$ and $f''(x)$ exists to answer one question: *what value of the input makes the output as large or as small as possible?* This guide pulls together critical points, the closed-interval method, and the second-derivative test, then walks through three exam-style word problems.

## Critical points

A **critical point** of $f$ is an interior point $c$ in the domain where either

- $f'(c) = 0$, or
- $f'(c)$ does not exist.

By Fermat's theorem, every interior local max or min occurs at a critical point. Critical points are *candidates* for extrema — not guarantees.

## The closed-interval method (extreme value theorem)

If $f$ is continuous on $[a, b]$, then $f$ attains an absolute max and absolute min on $[a, b]$. To find them:

1. Find all critical points of $f$ in $(a, b)$.
2. Evaluate $f$ at each critical point and at the endpoints $a$ and $b$.
3. The largest value is the absolute max; the smallest is the absolute min.

This is brutal but reliable. Use it whenever the problem is on a closed bounded interval. The slides at {{FILE:wsu-math171-optimization-slides}} have a clean worked example.

**Example.** Find the absolute max and min of $f(x) = x^3 - 3x + 1$ on $[0, 2]$.

$f'(x) = 3x^2 - 3 = 3(x-1)(x+1)$. Critical points in $(0, 2)$: $x = 1$ (we discard $x = -1$ because it is outside the interval). Evaluate:

$$
f(0) = 1, \qquad f(1) = -1, \qquad f(2) = 3.
$$

Absolute max is $3$ at $x = 2$; absolute min is $-1$ at $x = 1$.

## First-derivative test

For a critical point $c$ where $f'(c) = 0$:

- If $f'$ changes from $+$ to $-$ at $c$, then $f(c)$ is a **local max**.
- If $f'$ changes from $-$ to $+$ at $c$, then $f(c)$ is a **local min**.
- If $f'$ does not change sign, $c$ is neither (it is a horizontal-tangent inflection).

This is the test to use when the second derivative is messy.

## Second-derivative test

Provided $f''(c)$ exists:

- $f'(c) = 0$ and $f''(c) > 0$ $\Longrightarrow$ local **min** at $c$ (concave up).
- $f'(c) = 0$ and $f''(c) < 0$ $\Longrightarrow$ local **max** at $c$ (concave down).
- $f''(c) = 0$ $\Longrightarrow$ inconclusive — fall back to the first-derivative test.

This is faster on the exam when $f''$ is easy to compute. It says nothing about absolute extrema on its own — for those you still need the closed-interval method or a global argument.

## Optimization word problems — the recipe

1. **Read the problem twice.** Identify what is being optimised (the *objective*) and what constraint is given.
2. **Draw a picture and label every variable.**
3. **Write the objective as a function of the variables.**
4. **Use the constraint to eliminate variables until only one remains.**
5. **Find critical points and confirm with the second-derivative test or closed-interval method.**
6. **Answer the original question** (sometimes the problem asks for the minimum *value*, sometimes for the *coordinates*).

## Worked problem 1: the rectangular pen

A farmer has $400$ ft of fencing and wants to enclose a rectangular pen. What dimensions maximise the area?

Let $x$ and $y$ be the side lengths. Constraint: $2x + 2y = 400$, so $y = 200 - x$. Objective:

$$
A(x) = x(200 - x) = 200x - x^2.
$$

$A'(x) = 200 - 2x$. Setting this to $0$ gives $x = 100$. $A''(x) = -2 < 0$, so this is a maximum.

The optimal pen is a $100 \times 100$ square with area $10\,000$ ft$^2$. Squares almost always win these problems, which is a good intuition check.

## Worked problem 2: the open-top box

You start with a $20 \times 30$ inch sheet of cardboard and cut squares of side $x$ from each corner, then fold up the sides to form an open-top box. What value of $x$ maximises the volume?

After folding, the box has length $30 - 2x$, width $20 - 2x$, height $x$. Volume:

$$
V(x) = x(30 - 2x)(20 - 2x) = 4x^3 - 100 x^2 + 600 x.
$$

Domain: $0 < x < 10$ (so width stays positive).

$V'(x) = 12 x^2 - 200 x + 600 = 4(3x^2 - 50 x + 150)$. The quadratic formula gives

$$
x = \frac{50 \pm \sqrt{2500 - 1800}}{6} = \frac{50 \pm \sqrt{700}}{6}.
$$

Numerically, $x \approx 3.92$ or $x \approx 12.74$. Only $x \approx 3.92$ is in the domain. Check $V''$ or compare endpoint values to confirm it is the max.

## Worked problem 3: the can

What dimensions of a closed cylindrical can of volume $V_0$ minimise the surface area?

Let $r$ be the radius and $h$ the height. Constraint: $\pi r^2 h = V_0$, so $h = V_0 / (\pi r^2)$. Surface area:

$$
S(r) = 2\pi r^2 + 2\pi r h = 2\pi r^2 + \frac{2 V_0}{r}.
$$

$S'(r) = 4\pi r - 2 V_0 / r^2$. Setting this to $0$:

$$
4\pi r = \frac{2 V_0}{r^2} \quad \Longrightarrow \quad r^3 = \frac{V_0}{2\pi}.
$$

The optimal can has $h = 2r$ — height equal to the diameter. Real soup cans are always shorter than this because of how labels and stacking work, but the math is clean.

## Common mistakes

- Forgetting to use the constraint to eliminate variables — you cannot optimise a function of two independent variables in MATH 171.
- Using the second-derivative test when $f''(c) = 0$ — that case is inconclusive, switch to the first-derivative test.
- Skipping the endpoints when the domain is closed — the absolute max may live at $a$ or $b$, not at a critical point.
- Solving for the wrong quantity. If the problem asks for the dimensions, give the dimensions; if it asks for the maximum value, give the value.

## Practice

When you can reproduce the rectangular pen problem from scratch and recite the closed-interval method without looking, take {{QUIZ:math171-optimization-quiz}}.

Course catalog: {{COURSE:wsu/math171}}.
