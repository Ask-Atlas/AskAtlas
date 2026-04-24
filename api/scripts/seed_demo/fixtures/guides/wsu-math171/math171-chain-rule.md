---
slug: math171-chain-rule
course:
  ipeds_id: "236939"
  department: "MATH"
  number: "171"
title: "The Chain Rule — MATH 171"
description: "Differentiating composite functions, including the most common combinations with trig, exp, and log."
tags: ["calculus", "derivatives", "chain-rule", "composition"]
author_role: bot
attached_files:
  - wsu-math171-derivative-rules
attached_resources: []
---

# The Chain Rule

The chain rule is the single most-used differentiation rule on the MATH 171 final. Almost every "real" function you meet in the second half of the course is a composition — $\sin(3x)$, $e^{x^2}$, $\sqrt{x^2 + 1}$ — and you need a way to differentiate them without unpacking the limit definition each time.

## Statement

If $h(x) = f(g(x))$, then

$$
h'(x) = f'(g(x)) \cdot g'(x).
$$

In Leibniz notation this looks even cleaner. If $y = f(u)$ and $u = g(x)$, then

$$
\frac{dy}{dx} = \frac{dy}{du} \cdot \frac{du}{dx}.
$$

The Leibniz form is the easier one to remember on an exam: the inner derivative "cancels" against the inner symbol, leaving $dy/dx$.

## The "outside-inside" mantra

For most exam problems, you can apply the chain rule by reciting:

> Differentiate the outside, leaving the inside alone, then multiply by the derivative of the inside.

For $h(x) = \sin(3x^2)$:

1. **Outside:** $\sin u$, derivative $\cos u$.
2. **Leave the inside alone:** $\cos(3x^2)$.
3. **Multiply by inside's derivative:** $\frac{d}{dx}[3x^2] = 6x$.

Result: $h'(x) = 6x \cos(3x^2)$.

## Worked examples

**Example 1.** $\frac{d}{dx}[(x^2 + 1)^{10}] = 10(x^2 + 1)^9 \cdot 2x = 20x(x^2 + 1)^9$.

**Example 2.** $\frac{d}{dx}[e^{x^2}] = e^{x^2} \cdot 2x = 2x e^{x^2}$.

**Example 3.** $\frac{d}{dx}[\ln(\sec x)] = \frac{1}{\sec x} \cdot \sec x \tan x = \tan x$.

**Example 4.** $\frac{d}{dx}\left[\sqrt{x^2 + 1}\right] = \frac{1}{2\sqrt{x^2 + 1}} \cdot 2x = \frac{x}{\sqrt{x^2 + 1}}$.

The fourth example is one you will see again in {{GUIDE:math171-related-rates}} — distance functions almost always require this.

## Multiple applications

The chain rule composes with itself. For $f(x) = \sin(\cos(3x))$:

$$
f'(x) = \cos(\cos(3x)) \cdot \frac{d}{dx}[\cos(3x)] = \cos(\cos(3x)) \cdot (-\sin(3x)) \cdot 3 = -3 \sin(3x) \cos(\cos(3x)).
$$

Work outside-in, peeling one layer at a time. Do not try to do it all in one step — that is where errors sneak in.

## Combining with the product and quotient rules

Most exam problems ask you to combine rules. The cheatsheet at {{FILE:wsu-math171-derivative-rules}} has a worked example sheet, but the core idea is: differentiate the *whole* with the product or quotient rule, and use the chain rule whenever you need to differentiate one of the factors.

**Example.** Differentiate $f(x) = x^2 \sin(3x)$.

By the product rule:

$$
f'(x) = 2x \sin(3x) + x^2 \cdot \frac{d}{dx}[\sin(3x)] = 2x \sin(3x) + x^2 \cdot 3 \cos(3x).
$$

The chain rule shows up in the second term and produces the factor of $3$.

## Why the chain rule is true (sketch)

If $g$ is differentiable at $x$ and $f$ is differentiable at $g(x)$, then near $x$,

$$
g(x + h) \approx g(x) + g'(x) \cdot h, \qquad f(g(x) + k) \approx f(g(x)) + f'(g(x)) \cdot k.
$$

Substituting $k = g(x + h) - g(x) \approx g'(x) h$ and dividing by $h$ gives the chain rule. The full proof uses the linearisation more carefully, but this is the picture.

## Common mistakes

- Forgetting to multiply by the inside derivative. $\frac{d}{dx}[\sin(3x)] = \cos(3x)$ is wrong — you need the extra factor of $3$.
- Confusing the chain rule with the product rule. The composition $\sin(3x)$ is *not* a product of $\sin$ and $3x$.
- Stopping too early on multi-layer compositions like $\sin(\cos(3x))$. Peel every layer.
- Misreading $\sqrt{x^2 + 1}$ as $\sqrt{x^2} + 1 = x + 1$. The square root is the *outside* function applied to $x^2 + 1$.

## Practice

The next guide, {{GUIDE:math171-implicit-differentiation}}, depends on you being fluent in the chain rule — every implicit derivative is secretly a chain rule application with $y$ as the inside. Get comfortable here first, then move on.

Course catalog: {{COURSE:wsu/math171}}.
