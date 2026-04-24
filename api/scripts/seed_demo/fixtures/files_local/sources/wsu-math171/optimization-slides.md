---
slug: wsu-math171-optimization-slides
title: "Optimization: Max/Min Problems Step by Step"
mime: application/vnd.openxmlformats-officedocument.presentationml.presentation
filename: optimization-slides.pptx
course: wsu/math171
description: "Slide deck walking through optimization problems — the five-step method plus classic fence, box, and can examples."
author_role: bot
---

# Optimization: Max/Min Problems

## What Is an Optimization Problem?

Find the largest or smallest value of a quantity (area, volume, cost, time) subject to a constraint.

## The Five-Step Method

1. Draw and label a diagram.
2. Write an equation for the quantity to optimize (the **objective**).
3. Write an equation for any constraint.
4. Reduce the objective to one variable using the constraint.
5. Take the derivative, find critical points, and verify max vs min.

## Step 3: Critical Points

A critical point of $f$ is where $f'(x) = 0$ or $f'(x)$ does not exist. Candidates for max/min come from critical points **and** endpoints of the domain.

## Extreme Value Theorem

A continuous function on a closed interval $[a, b]$ attains both a max and a min on that interval.

## First Derivative Test

If $f'$ changes from positive to negative at $c$, then $c$ is a local max. If $f'$ changes from negative to positive, $c$ is a local min.

## Second Derivative Test

If $f'(c) = 0$ and $f''(c) > 0$, then $c$ is a local min. If $f''(c) < 0$, $c$ is a local max.

## Example 1: Fence Problem

A farmer has 200 ft of fencing and wants a rectangular pen against a river (no fence needed on river side). Maximize area.

## Fence — Setup

Let $x$ = side perpendicular to river, $y$ = side parallel. Fencing used: $2x + y = 200$.

## Fence — Objective

Area $A(x,y) = xy$. Solve constraint: $y = 200 - 2x$.

## Fence — Reduce

$A(x) = x(200 - 2x) = 200x - 2x^2$. Domain: $0 \leq x \leq 100$.

## Fence — Optimize

$A'(x) = 200 - 4x = 0 \implies x = 50$. Then $y = 100$. Max area = $5000$ ft$^2$.

## Fence — Verify

$A''(x) = -4 < 0$, so $x = 50$ is a local max. Endpoint check: $A(0) = A(100) = 0$. Confirmed.

## Example 2: Box Problem

An open-top box is made from a 12 in $\times$ 12 in square by cutting squares of side $x$ from each corner and folding. Maximize volume.

## Box — Setup

Base: $(12 - 2x) \times (12 - 2x)$. Height: $x$. Domain: $0 < x < 6$.

## Box — Objective

$V(x) = x(12 - 2x)^2 = x(144 - 48x + 4x^2) = 144x - 48x^2 + 4x^3$.

## Box — Optimize

$V'(x) = 144 - 96x + 12x^2 = 12(x^2 - 8x + 12) = 12(x - 2)(x - 6)$.

Critical points: $x = 2$ and $x = 6$. Reject $x = 6$ (boundary — zero volume).

## Box — Answer

At $x = 2$: $V = 2 \cdot 8^2 = 128$ in$^3$. $V'$ goes from + to −, confirming max.

## Example 3: Can Problem

Design a cylindrical can with volume 355 cm$^3$ that minimizes surface area.

## Can — Setup

$V = \pi r^2 h = 355$ ⟹ $h = \frac{355}{\pi r^2}$.

## Can — Objective

$S = 2\pi r^2 + 2\pi r h = 2\pi r^2 + \frac{710}{r}$.

## Can — Optimize

$S'(r) = 4\pi r - \frac{710}{r^2} = 0 \implies r^3 = \frac{710}{4\pi} \implies r = \sqrt[3]{\frac{710}{4\pi}} \approx 3.84$ cm.

Then $h = \frac{355}{\pi r^2} \approx 7.67$ cm — note $h = 2r$, the classic "height equals diameter" rule.

## Key Takeaways

- Reduce to **one** variable before differentiating.
- Always check endpoints and the sign test.
- Many optimization answers satisfy a clean geometric ratio — a good sanity check.
