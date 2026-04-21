---
slug: wsu-math171-derivative-rules
title: "Derivative Rules: Power, Product, Quotient, Chain, and Implicit"
mime: application/pdf
filename: derivative-rules.pdf
course: wsu/math171
description: "All the core differentiation rules at a glance — power, product, quotient, chain, and implicit — with worked examples."
author_role: bot
---

# Derivative Rules

## Definition

The derivative of $f$ at $x$ is

$$f'(x) = \lim_{h \to 0} \frac{f(x + h) - f(x)}{h}$$

when the limit exists. Every rule below is a shortcut that saves you from re-doing this limit.

## Core Rules Table

| Rule | Formula |
|---|---|
| Constant | $\frac{d}{dx}[c] = 0$ |
| Power | $\frac{d}{dx}[x^n] = n x^{n-1}$ |
| Constant multiple | $\frac{d}{dx}[c f(x)] = c f'(x)$ |
| Sum / difference | $(f \pm g)' = f' \pm g'$ |
| Product | $(f g)' = f' g + f g'$ |
| Quotient | $(f/g)' = \dfrac{f' g - f g'}{g^2}$ |
| Chain | $\dfrac{d}{dx}[f(g(x))] = f'(g(x)) \cdot g'(x)$ |
| Exponential | $\frac{d}{dx}[e^x] = e^x$ |
| Exp. base $a$ | $\frac{d}{dx}[a^x] = a^x \ln a$ |
| Natural log | $\frac{d}{dx}[\ln x] = \frac{1}{x}$ |
| Sine / cosine | $\frac{d}{dx}[\sin x] = \cos x$, $\frac{d}{dx}[\cos x] = -\sin x$ |
| Tangent | $\frac{d}{dx}[\tan x] = \sec^2 x$ |
| Inverse sine | $\frac{d}{dx}[\arcsin x] = \frac{1}{\sqrt{1 - x^2}}$ |
| Inverse tan | $\frac{d}{dx}[\arctan x] = \frac{1}{1 + x^2}$ |

## Product Rule, Worked

Differentiate $f(x) = x^2 \sin x$.

$$f'(x) = 2x \cdot \sin x + x^2 \cdot \cos x.$$

## Quotient Rule, Worked

Differentiate $f(x) = \dfrac{x^2 + 1}{x - 3}$.

$$f'(x) = \frac{2x(x - 3) - (x^2 + 1)(1)}{(x - 3)^2} = \frac{x^2 - 6x - 1}{(x - 3)^2}.$$

## Chain Rule, Worked

Differentiate $f(x) = \sin(3x^2 + 1)$.

Let $u = 3x^2 + 1$. Then $f = \sin u$, $f'(x) = \cos(u) \cdot u' = \cos(3x^2 + 1) \cdot 6x$.

Another: $\frac{d}{dx}[e^{x^2}] = e^{x^2} \cdot 2x$.

## Implicit Differentiation

When $y$ is defined implicitly by an equation in $x$ and $y$, differentiate both sides with respect to $x$, treating $y$ as a function of $x$, then solve for $\frac{dy}{dx}$.

**Example.** Find $\frac{dy}{dx}$ for $x^2 + y^2 = 25$.

Differentiate: $2x + 2y \frac{dy}{dx} = 0$, so $\frac{dy}{dx} = -\frac{x}{y}$.

**Example.** $x^3 + y^3 = 6xy$ (folium of Descartes).

$3x^2 + 3y^2 y' = 6y + 6x y'$. Group: $y'(3y^2 - 6x) = 6y - 3x^2$. So $y' = \dfrac{6y - 3x^2}{3y^2 - 6x} = \dfrac{2y - x^2}{y^2 - 2x}$.

## Higher-Order Derivatives

$f''(x) = \frac{d}{dx}[f'(x)]$ is the second derivative. Geometrically, $f''>0$ means concave up, $f''<0$ means concave down. Inflection points occur where $f''$ changes sign.

## Common Pitfalls

- Do not forget the chain rule when differentiating compositions — even $(x^2 + 1)^{10}$ needs it.
- The quotient rule's numerator is $f'g - fg'$, **not** $fg' - f'g$. Order matters.
- $\frac{d}{dx}[x^x]$ is neither power nor exponential — use logarithmic differentiation: $\ln y = x \ln x$, so $y'/y = \ln x + 1$, giving $y' = x^x(\ln x + 1)$.
