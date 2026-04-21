---
slug: math171-implicit-differentiation
course:
  ipeds_id: "236939"
  department: "MATH"
  number: "171"
title: "Implicit Differentiation — MATH 171"
description: "Differentiating curves defined implicitly, plus tangent-line problems and logarithmic differentiation."
tags: ["calculus", "derivatives", "implicit", "tangent-lines"]
author_role: bot
attached_files:
  - wsu-math171-derivative-rules
attached_resources: []
---

# Implicit Differentiation

Up to this point you have always been able to write $y$ explicitly as a function of $x$. But many of the curves that show up in MATH 171 — circles, ellipses, the folium of Descartes — are not the graph of a function. They are defined implicitly by an equation like

$$
x^2 + y^2 = 25 \quad \text{or} \quad x^3 + y^3 = 6xy.
$$

Implicit differentiation lets you find $dy/dx$ without solving for $y$.

## The core idea

Treat $y$ as a function of $x$. Whenever you differentiate a term containing $y$, apply the chain rule with $y$ as the inside function — that produces an extra factor of $dy/dx$. Then solve for $dy/dx$.

If you struggled in {{GUIDE:math171-chain-rule}}, go back and shore that up first. Implicit differentiation is just the chain rule in disguise.

## Worked example: the circle

Differentiate $x^2 + y^2 = 25$ with respect to $x$:

$$
\frac{d}{dx}[x^2] + \frac{d}{dx}[y^2] = \frac{d}{dx}[25].
$$

The first term is $2x$. The second term, by the chain rule, is $2y \cdot \frac{dy}{dx}$. The right side is $0$. So

$$
2x + 2y \frac{dy}{dx} = 0 \quad \Longrightarrow \quad \frac{dy}{dx} = -\frac{x}{y}.
$$

This formula gives the slope of the tangent line at any point $(x, y)$ on the circle of radius $5$. Try it at $(3, 4)$: slope $-3/4$, perpendicular to the radius from the origin — exactly what geometry predicts.

## Worked example: a more interesting curve

Find $dy/dx$ for $x^3 + y^3 = 6xy$.

Differentiate both sides:

$$
3x^2 + 3y^2 \frac{dy}{dx} = 6y + 6x \frac{dy}{dx}.
$$

Notice the right side: $6xy$ is a product, so it needs the product rule, and $y$ contributes a $dy/dx$ factor. Now collect $dy/dx$ terms:

$$
3y^2 \frac{dy}{dx} - 6x \frac{dy}{dx} = 6y - 3x^2 \quad \Longrightarrow \quad \frac{dy}{dx} = \frac{6y - 3x^2}{3y^2 - 6x} = \frac{2y - x^2}{y^2 - 2x}.
$$

The cheatsheet in {{FILE:wsu-math171-derivative-rules}} has more practice curves of this form.

## Procedure

1. Differentiate both sides with respect to $x$, treating $y$ as a function of $x$.
2. Every term involving $y$ gets a factor of $dy/dx$ from the chain rule.
3. Use the product rule on terms like $xy$ or $x^2 y$.
4. Collect all $dy/dx$ terms on one side.
5. Factor and divide.

## Tangent-line problems

The most common exam problem is "find the equation of the tangent line at $(a, b)$." The recipe is:

1. Use implicit differentiation to find $dy/dx$ as a function of $x$ and $y$.
2. Substitute $(a, b)$ to get the slope $m$.
3. Write the line: $y - b = m(x - a)$.

For the circle example at $(3, 4)$: slope $-3/4$, tangent line $y - 4 = -\frac{3}{4}(x - 3)$.

## Logarithmic differentiation

Some derivatives are nasty enough to deserve a special trick: take the natural log first, then differentiate implicitly. This is the cleanest way to handle expressions of the form $f(x)^{g(x)}$, like $y = x^x$.

For $y = x^x$ with $x > 0$:

$$
\ln y = x \ln x.
$$

Differentiate both sides with respect to $x$:

$$
\frac{1}{y} \frac{dy}{dx} = \ln x + x \cdot \frac{1}{x} = \ln x + 1.
$$

Multiply through by $y = x^x$:

$$
\frac{dy}{dx} = x^x (\ln x + 1).
$$

Without the log trick this is genuinely painful — you cannot use the power rule (the exponent is variable) and you cannot use the exponential rule directly (the base is variable).

## Common mistakes

- Forgetting the chain rule on $y$. Writing $\frac{d}{dx}[y^2] = 2y$ instead of $2y \frac{dy}{dx}$ is the most frequent error on MATH 171 midterms.
- Forgetting the product rule on terms like $xy$ — this is *not* a single variable raised to a power.
- Mixing notation. Stay consistent: either always write $dy/dx$ or always write $y'$.
- Trying to "solve for $y$ first" on curves where $y$ has no closed form. Implicit differentiation is what you do precisely when you cannot do that.

## Practice

After implicit differentiation comes its most natural application: {{GUIDE:math171-related-rates}}, where two variables that depend on time are linked by an equation and you differentiate with respect to $t$.

Course catalog: {{COURSE:wsu/math171}}.
