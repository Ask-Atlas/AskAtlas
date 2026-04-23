---
slug: math171-integral-fundamentals
course:
  ipeds_id: "236939"
  department: "MATH"
  number: "171"
title: "Integral Fundamentals — MATH 171"
description: "Antiderivatives, indefinite integrals, basic integration formulas, and Riemann sum intuition."
tags: ["calculus", "integrals", "antiderivatives", "riemann-sums"]
author_role: bot
attached_files:
  - smoke-wikibooks-calculus-pdf
  - wsu-math171-derivative-rules
attached_resources: []
---

# Integral Fundamentals

The first half of MATH 171 was about derivatives. The second half is about *undoing* them. The fundamental object is the **antiderivative**: given $f(x)$, find a function $F(x)$ with $F'(x) = f(x)$.

## Antiderivatives and the indefinite integral

If $F'(x) = f(x)$, then $F$ is an antiderivative of $f$. The notation

$$
\int f(x)\, dx = F(x) + C
$$

denotes the **family** of all antiderivatives of $f$. The constant $C$ is unavoidable — if $F'(x) = f(x)$, then $(F(x) + 17)' = f(x)$ too. Forgetting $+ C$ on an indefinite integral will cost you points on every exam.

## The basic integration formulas

Reverse every entry in the derivative table from {{GUIDE:math171-derivative-rules}}:

| Integral | Result |
|---|---|
| $\int x^n\, dx$ | $\frac{x^{n+1}}{n + 1} + C$, provided $n \ne -1$ |
| $\int \frac{1}{x}\, dx$ | $\ln \lvert x \rvert + C$ |
| $\int e^x\, dx$ | $e^x + C$ |
| $\int a^x\, dx$ | $\frac{a^x}{\ln a} + C$ |
| $\int \cos x\, dx$ | $\sin x + C$ |
| $\int \sin x\, dx$ | $-\cos x + C$ |
| $\int \sec^2 x\, dx$ | $\tan x + C$ |
| $\int \sec x \tan x\, dx$ | $\sec x + C$ |
| $\int \frac{1}{\sqrt{1 - x^2}}\, dx$ | $\arcsin x + C$ |
| $\int \frac{1}{1 + x^2}\, dx$ | $\arctan x + C$ |

The $n = -1$ exception in the power rule is why $\ln \lvert x \rvert$ deserves its own row.

## Linearity

Integrals respect sums and constant multiples:

$$
\int [a f(x) + b g(x)]\, dx = a \int f(x)\, dx + b \int g(x)\, dx.
$$

There is **no** product rule or chain rule for integration — the integral of a product is not the product of integrals. Those need substitution or integration by parts, which are taught later in the term.

## Worked example: the polynomial

$$
\int (4x^3 - 6x^2 + 2x + 5)\, dx = x^4 - 2 x^3 + x^2 + 5 x + C.
$$

Differentiate the answer to check — you should get back the integrand.

## Worked example: rewrite first

$$
\int \frac{x^2 - 3 \sqrt{x}}{x}\, dx = \int (x - 3 x^{-1/2})\, dx = \frac{x^2}{2} - 6 \sqrt{x} + C.
$$

The opening move on most exam integrals is to rewrite. Power rules want exponents, not square roots and fraction bars.

## The definite integral as a Riemann sum

The **definite integral** of $f$ over $[a, b]$ is

$$
\int_a^b f(x)\, dx = \lim_{n \to \infty} \sum_{i=1}^{n} f(x_i^*) \Delta x,
$$

where $\Delta x = (b - a)/n$ and $x_i^*$ is any sample point in the $i$-th subinterval. Geometrically this is the *signed area* between the graph of $f$ and the $x$-axis on $[a, b]$ — area above the axis counts as positive, area below as negative.

The right-endpoint sum, with $x_i^* = a + i \Delta x$, is the most common variant on MATH 171 exams. The book in {{FILE:smoke-wikibooks-calculus-pdf}} has additional Riemann sum walkthroughs.

## Properties of the definite integral

Assume $f$ and $g$ are integrable on the relevant interval.

- **Linearity:** $\int_a^b [c_1 f + c_2 g]\, dx = c_1 \int_a^b f\, dx + c_2 \int_a^b g\, dx$.
- **Reversal:** $\int_a^b f\, dx = -\int_b^a f\, dx$.
- **Splitting:** $\int_a^b f\, dx = \int_a^c f\, dx + \int_c^b f\, dx$.
- **Zero width:** $\int_a^a f\, dx = 0$.
- **Comparison:** if $f(x) \le g(x)$ on $[a, b]$, then $\int_a^b f\, dx \le \int_a^b g\, dx$.

## Why the antiderivative connection matters

The bridge between antiderivatives (an algebra game) and Riemann sums (a limit-of-area game) is the Fundamental Theorem of Calculus, which gets its own treatment in {{GUIDE:math171-ftc-summary}}. Briefly: if $F$ is any antiderivative of $f$, then

$$
\int_a^b f(x)\, dx = F(b) - F(a).
$$

That single equation is why we spend three weeks memorising antiderivatives.

## Common mistakes

- Forgetting $+ C$ on indefinite integrals.
- Trying to use a "product rule for integrals" that does not exist.
- Power-ruling the case $n = -1$ — that gives $x^0 / 0$, which is nonsense. Use $\ln \lvert x \rvert$ instead.
- Confusing $\int 1/x\, dx = \ln \lvert x \rvert$ with $\int 1/x^2\, dx = -1/x$. The first is the special case; the second is the power rule.

## Practice

After this guide, jump to {{GUIDE:math171-ftc-summary}} for the formal connection between antiderivatives and definite integrals.

Course catalog: {{COURSE:wsu/math171}}.
