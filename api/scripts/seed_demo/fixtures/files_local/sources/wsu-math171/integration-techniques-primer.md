---
slug: wsu-math171-integration-techniques-primer
title: "Integration Techniques: A Primer on u-Substitution, Integration by Parts, and Partial Fractions"
mime: application/epub+zip
filename: integration-techniques-primer.epub
course: wsu/math171
description: "Multi-section primer covering the three core integration techniques — u-substitution, by parts, and partial fractions — with worked examples."
author_role: bot
---

# Integration Techniques: A Primer

Integration is not a single procedure the way differentiation is. It is a toolkit. Three tools cover most MATH 171 and introductory MATH 172 integrals: $u$-substitution, integration by parts, and partial fractions. This primer walks through each, one section at a time.

## Section 1 — $u$-Substitution

$u$-substitution is the chain rule in reverse. If you can spot an "inside function" $u = g(x)$ whose derivative $du = g'(x)\,dx$ also appears in the integrand, you can rewrite the integral in terms of $u$ and solve a simpler problem.

**The pattern.** $\int f(g(x)) \cdot g'(x) \, dx = \int f(u) \, du$.

**Example 1.1.** Evaluate $\int 2x \cos(x^2) \, dx$.

Let $u = x^2$, so $du = 2x\,dx$. The integral becomes $\int \cos u \, du = \sin u + C = \sin(x^2) + C$.

**Example 1.2.** Evaluate $\int \frac{x}{x^2 + 1} \, dx$.

Let $u = x^2 + 1$, so $du = 2x\,dx$, meaning $x \, dx = \tfrac{1}{2}du$. The integral becomes $\tfrac{1}{2}\int \frac{1}{u} du = \tfrac{1}{2}\ln|u| + C = \tfrac{1}{2}\ln(x^2 + 1) + C$.

**Definite integrals.** When using $u$-sub on $\int_a^b$, change the limits: if $u = g(x)$, new limits are $g(a)$ and $g(b)$. You never have to back-substitute.

## Section 2 — Integration by Parts

Integration by parts is the product rule in reverse. Starting from $(uv)' = u'v + uv'$ and integrating both sides:

$$\int u \, dv = uv - \int v \, du.$$

**Choosing $u$ and $dv$.** Use the mnemonic LIATE: prefer $u$ in this priority order — Logarithmic, Inverse trig, Algebraic, Trigonometric, Exponential.

**Example 2.1.** Evaluate $\int x e^x \, dx$.

Take $u = x$ (algebraic) and $dv = e^x\,dx$. Then $du = dx$ and $v = e^x$. Apply the formula:

$$\int x e^x \, dx = x e^x - \int e^x \, dx = x e^x - e^x + C = e^x(x - 1) + C.$$

**Example 2.2.** Evaluate $\int \ln x \, dx$.

Take $u = \ln x$, $dv = dx$. Then $du = \tfrac{1}{x}dx$, $v = x$. So

$$\int \ln x \, dx = x \ln x - \int x \cdot \tfrac{1}{x}\,dx = x \ln x - x + C.$$

**Repeated use.** Some integrals need integration by parts twice, for instance $\int x^2 \sin x \, dx$. Sometimes after two passes the original integral reappears; solve algebraically for it.

## Section 3 — Partial Fractions

Partial fractions handles rational functions $\int \frac{P(x)}{Q(x)} \, dx$ where $\deg P < \deg Q$. The idea: decompose the fraction into a sum of simpler pieces that each integrate easily.

**Step 1.** Factor $Q(x)$ completely over the reals.

**Step 2.** Write one term per factor:
- Linear factor $(x - a)$: term $\frac{A}{x - a}$.
- Repeated linear $(x - a)^n$: terms $\frac{A_1}{x - a} + \cdots + \frac{A_n}{(x - a)^n}$.
- Irreducible quadratic $(x^2 + bx + c)$: term $\frac{Ax + B}{x^2 + bx + c}$.

**Step 3.** Solve for the unknown constants by clearing denominators and matching coefficients (or plugging in convenient $x$ values).

**Example 3.1.** Evaluate $\int \frac{1}{x^2 - 1}\,dx$.

Factor: $x^2 - 1 = (x - 1)(x + 1)$. Decompose:

$$\frac{1}{(x-1)(x+1)} = \frac{A}{x-1} + \frac{B}{x+1}.$$

Clearing: $1 = A(x+1) + B(x-1)$. Plug $x = 1$: $1 = 2A$, so $A = \tfrac{1}{2}$. Plug $x = -1$: $1 = -2B$, so $B = -\tfrac{1}{2}$.

$$\int \frac{1}{x^2 - 1} dx = \tfrac{1}{2}\ln|x-1| - \tfrac{1}{2}\ln|x+1| + C = \tfrac{1}{2}\ln\left|\frac{x-1}{x+1}\right| + C.$$

## Section 4 — Choosing a Technique

1. Does a substitution $u$ simplify to $\int f(u)\,du$? Try $u$-sub first.
2. Is the integrand a product of different function types (polynomial times exponential, log times algebraic)? Try by parts.
3. Is it a rational function of polynomials? Try partial fractions, after polynomial long division if $\deg P \geq \deg Q$.

Practice shifts you from "what do I do?" to recognition at a glance.
