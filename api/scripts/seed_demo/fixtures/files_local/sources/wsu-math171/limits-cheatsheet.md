---
slug: wsu-math171-limits-cheatsheet
title: "Limits Cheatsheet: Epsilon-Delta, Limit Laws, and L'Hopital's Rule"
mime: application/pdf
filename: limits-cheatsheet.pdf
course: wsu/math171
description: "Quick reference for epsilon-delta definition, limit laws, indeterminate forms, and L'Hopital's Rule with worked examples."
author_role: bot
---

# Limits Cheatsheet

## 1. The Epsilon-Delta Definition

We say $\lim_{x \to a} f(x) = L$ if for every $\varepsilon > 0$ there exists a $\delta > 0$ such that

$$0 < |x - a| < \delta \implies |f(x) - L| < \varepsilon.$$

Intuition: no matter how tight a tolerance $\varepsilon$ you demand around $L$, we can find a neighborhood around $a$ (of radius $\delta$) where $f(x)$ stays inside that tolerance.

**Worked example.** Prove $\lim_{x \to 3}(2x - 1) = 5$.

Given $\varepsilon > 0$, we need $|(2x - 1) - 5| < \varepsilon$, i.e. $|2x - 6| < \varepsilon$, i.e. $|x - 3| < \varepsilon/2$. Pick $\delta = \varepsilon/2$.

## 2. Limit Laws

If $\lim_{x \to a} f(x) = L$ and $\lim_{x \to a} g(x) = M$, then:

| Law | Statement |
|---|---|
| Sum | $\lim (f + g) = L + M$ |
| Difference | $\lim (f - g) = L - M$ |
| Product | $\lim (f \cdot g) = L \cdot M$ |
| Quotient | $\lim (f / g) = L / M$ if $M \neq 0$ |
| Power | $\lim f^n = L^n$ |
| Root | $\lim \sqrt[n]{f} = \sqrt[n]{L}$ (n even requires $L \geq 0$) |

**Squeeze Theorem.** If $g(x) \leq f(x) \leq h(x)$ near $a$ and $\lim g = \lim h = L$, then $\lim f = L$.

Classic use: $\lim_{x \to 0} x^2 \sin(1/x) = 0$ because $-x^2 \leq x^2 \sin(1/x) \leq x^2$.

## 3. Special Limits Worth Memorizing

$$\lim_{x \to 0} \frac{\sin x}{x} = 1, \qquad \lim_{x \to 0} \frac{1 - \cos x}{x} = 0, \qquad \lim_{x \to 0} \frac{1 - \cos x}{x^2} = \frac{1}{2}.$$

$$\lim_{n \to \infty} \left(1 + \frac{1}{n}\right)^n = e.$$

## 4. Indeterminate Forms

The seven indeterminate forms: $\frac{0}{0}, \frac{\infty}{\infty}, 0 \cdot \infty, \infty - \infty, 0^0, \infty^0, 1^\infty$.

When you get one of these, do not stop — transform the expression.

## 5. L'Hopital's Rule

If $\lim_{x \to a} f(x)/g(x)$ is of the form $0/0$ or $\infty/\infty$, and $f, g$ are differentiable near $a$ with $g'(x) \neq 0$, then

$$\lim_{x \to a} \frac{f(x)}{g(x)} = \lim_{x \to a} \frac{f'(x)}{g'(x)}$$

provided the right-hand limit exists.

**Worked example 1.** $\lim_{x \to 0} \frac{\sin x}{x} \stackrel{\mathrm{LH}}{=} \lim_{x \to 0} \frac{\cos x}{1} = 1.$

**Worked example 2.** $\lim_{x \to \infty} \frac{\ln x}{x} \stackrel{\mathrm{LH}}{=} \lim_{x \to \infty} \frac{1/x}{1} = 0.$

**Worked example 3 (rewrite to 0/0).** $\lim_{x \to 0^+} x \ln x = \lim_{x \to 0^+} \frac{\ln x}{1/x} \stackrel{\mathrm{LH}}{=} \lim_{x \to 0^+} \frac{1/x}{-1/x^2} = \lim_{x \to 0^+} (-x) = 0.$

## 6. One-Sided Limits and Continuity

$f$ is continuous at $a$ iff $\lim_{x \to a^-} f(x) = \lim_{x \to a^+} f(x) = f(a)$. All three must agree and the point value must exist.
