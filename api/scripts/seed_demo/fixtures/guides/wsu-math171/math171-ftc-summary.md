---
slug: math171-ftc-summary
course:
  ipeds_id: "236939"
  department: "MATH"
  number: "171"
title: "The Fundamental Theorem of Calculus — MATH 171"
description: "Both parts of the FTC, the evaluation theorem, and worked examples that connect derivatives and integrals."
tags: ["calculus", "ftc", "integrals", "antiderivatives", "final-exam"]
author_role: bot
attached_files:
  - wsu-math171-ftc-summary
  - smoke-wikibooks-calculus-pdf
attached_resources: []
---

# The Fundamental Theorem of Calculus

The Fundamental Theorem of Calculus (FTC) is the punch line of the entire course. It says that two things that look completely unrelated — the slope of a tangent line and the area under a curve — are inverse operations. Once you internalise this, the second half of MATH 171 collapses into a much simpler picture.

## Setup

Let $f$ be continuous on $[a, b]$. Define a new function $g$ by integrating $f$ from the left endpoint up to a moving right endpoint:

$$
g(x) = \int_a^x f(t)\, dt.
$$

The variable of integration is $t$ (a *dummy*); the variable of $g$ is $x$. Mixing those two up is the most common notational mistake on exam problems.

## FTC Part 1 (the differentiation part)

If $f$ is continuous on $[a, b]$ and $g(x) = \int_a^x f(t)\, dt$, then $g$ is differentiable on $(a, b)$ and

$$
g'(x) = f(x).
$$

In words: differentiating an integral with respect to its upper limit gives back the integrand. The integral is the *antiderivative*. The cheatsheet at {{FILE:wsu-math171-ftc-summary}} has a clean diagram of the moving-area picture that motivates this.

### With the chain rule

When the upper limit is a function $u(x)$ instead of just $x$, multiply by the chain-rule factor:

$$
\frac{d}{dx} \int_a^{u(x)} f(t)\, dt = f(u(x)) \cdot u'(x).
$$

**Example.** $\frac{d}{dx} \int_0^{x^2} \cos(t)\, dt = \cos(x^2) \cdot 2x$.

### When the lower limit moves

Use the reversal property — flipping the limits flips the sign:

$$
\int_{u(x)}^{b} f(t)\, dt = -\int_b^{u(x)} f(t)\, dt.
$$

Then apply Part 1 as usual.

## FTC Part 2 (the evaluation part)

If $f$ is continuous on $[a, b]$ and $F$ is any antiderivative of $f$, then

$$
\int_a^b f(x)\, dx = F(b) - F(a).
$$

This is the workhorse for computing definite integrals. The standard notation is $\left[ F(x) \right]_a^b = F(b) - F(a)$.

**Example.** $\int_0^{\pi/2} \cos x\, dx = \left[ \sin x \right]_0^{\pi/2} = \sin(\pi/2) - \sin(0) = 1$.

**Example.** $\int_1^4 (3x^2 + 2x)\, dx = \left[ x^3 + x^2 \right]_1^4 = (64 + 16) - (1 + 1) = 78$.

**Example.** $\int_1^e \frac{1}{x}\, dx = \left[ \ln \lvert x \rvert \right]_1^e = \ln e - \ln 1 = 1$.

## Why the two parts go together

Read them side by side:

- Part 1: differentiating an integral gives back the integrand.
- Part 2: integrating a derivative gives back the original function (up to the boundary values).

Differentiation and integration are inverses. The "$+ C$" in indefinite integrals is exactly what gets killed off by the subtraction $F(b) - F(a)$ in Part 2.

## Worked example: combined Part 1 and chain rule

Compute $\frac{d}{dx} \int_{x}^{x^2} \sqrt{1 + t^3}\, dt$.

Split using the splitting property at any convenient constant — say $0$:

$$
\int_x^{x^2} \sqrt{1 + t^3}\, dt = \int_0^{x^2} \sqrt{1 + t^3}\, dt - \int_0^{x} \sqrt{1 + t^3}\, dt.
$$

Differentiating each piece with Part 1:

$$
\frac{d}{dx} \int_x^{x^2} \sqrt{1 + t^3}\, dt = \sqrt{1 + x^6} \cdot 2x - \sqrt{1 + x^3}.
$$

Notice we never had to find a closed-form antiderivative for $\sqrt{1 + t^3}$ — none exists in elementary functions. Part 1 lets you differentiate integrals you cannot evaluate.

## Net change interpretation

If $F'(t) = v(t)$ is a velocity and $F(t)$ is position, then

$$
F(b) - F(a) = \int_a^b v(t)\, dt
$$

is the **net displacement** between $t = a$ and $t = b$. To get *total distance travelled* you integrate $\lvert v(t) \rvert$ instead, splitting at sign changes.

## Common mistakes

- Confusing the integration variable with the limit variable. In $\int_0^{x^2} \cos t\, dt$, the answer depends on $x$, not on $t$.
- Forgetting the chain-rule factor when the upper limit is a function.
- Using FTC Part 2 with a function that is not continuous on the interval. The theorem requires continuity.
- Treating $+ C$ as still required after evaluating a definite integral. It cancels in $F(b) - F(a)$, so omit it.

## Practice

For richer integrals — the ones where guessing an antiderivative does not work — head to the integration techniques primer.

Course catalog: {{COURSE:wsu/math171}}.
