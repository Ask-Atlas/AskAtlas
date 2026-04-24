---
slug: math171-limits-and-continuity
course:
  ipeds_id: "236939"
  department: "MATH"
  number: "171"
title: "Limits and Continuity — MATH 171"
description: "Limit laws, one-sided limits, the squeeze theorem, and the formal definition of continuity."
tags: ["calculus", "limits", "continuity", "midterm", "foundations"]
author_role: bot
quiz_slug: math171-limits-and-continuity-quiz
attached_files:
  - wsu-math171-limits-cheatsheet
  - smoke-wikibooks-calculus-pdf
attached_resources: []
---

# Limits and Continuity

Limits are the gateway drug of calculus. Almost every theorem you see in MATH 171 — derivatives, integrals, the FTC — is ultimately a statement about a limit. If you can read $\lim_{x \to a} f(x) = L$ fluently, the rest of the semester is mostly bookkeeping.

## What a limit actually means

The informal idea is: $\lim_{x \to a} f(x) = L$ means that as $x$ gets arbitrarily close to $a$ (but not equal to $a$), $f(x)$ gets arbitrarily close to $L$. The formal $\epsilon$-$\delta$ version (which we will only sketch) says:

$$
\forall \epsilon > 0,\ \exists \delta > 0\ \text{such that}\ 0 < |x - a| < \delta \implies |f(x) - L| < \epsilon.
$$

You will not be asked to write an $\epsilon$-$\delta$ proof on the midterm, but you should recognise the shape — it explains why "the value at $a$ doesn't matter."

## The basic limit laws

Provided $\lim_{x \to a} f(x)$ and $\lim_{x \to a} g(x)$ both exist:

- **Sum:** $\lim (f + g) = \lim f + \lim g$
- **Constant multiple:** $\lim (cf) = c \lim f$
- **Product:** $\lim (fg) = (\lim f)(\lim g)$
- **Quotient:** $\lim (f/g) = (\lim f)/(\lim g)$, provided $\lim g \ne 0$
- **Power:** $\lim f^n = (\lim f)^n$ for integer $n \ge 1$

If both limits exist, you may move the limit inside any of these operations. The whole game in the first month of MATH 171 is recognising which form you have and applying the right rule.

## Direct substitution

For polynomial, rational (with non-zero denominator), trig, exponential, log, and root functions, you can usually just plug in:

$$
\lim_{x \to 2} (3x^2 - 5x + 1) = 3(4) - 5(2) + 1 = 3.
$$

When direct substitution gives $0/0$, you have an **indeterminate form** and need to do real work — factor, rationalise, or apply a known special limit.

## The classic 0/0 trick: factor and cancel

$$
\lim_{x \to 3} \frac{x^2 - 9}{x - 3} = \lim_{x \to 3} \frac{(x-3)(x+3)}{x-3} = \lim_{x \to 3} (x + 3) = 6.
$$

The cancellation is legal because the limit only cares about $x$ near $3$, not $x = 3$ itself.

## One-sided limits

Sometimes a function approaches different values from the left and the right:

$$
\lim_{x \to 0^-} \frac{|x|}{x} = -1, \qquad \lim_{x \to 0^+} \frac{|x|}{x} = +1.
$$

The two-sided limit $\lim_{x \to 0} \frac{|x|}{x}$ does **not** exist, because the one-sided limits disagree. A two-sided limit exists if and only if both one-sided limits exist and are equal.

## Limits at infinity

For rational functions, compare the leading degrees of numerator and denominator:

$$
\lim_{x \to \infty} \frac{2x^3 + x}{5x^3 - 4} = \frac{2}{5}.
$$

If the numerator's degree is smaller, the limit is $0$. If it is larger, the limit is $\pm\infty$ (use the leading coefficients to determine the sign).

## The squeeze theorem

If $g(x) \le f(x) \le h(x)$ near $a$ (except possibly at $a$), and $\lim_{x \to a} g(x) = \lim_{x \to a} h(x) = L$, then $\lim_{x \to a} f(x) = L$.

The textbook example is

$$
\lim_{x \to 0} x^2 \sin\!\left(\frac{1}{x}\right) = 0,
$$

which you cannot evaluate by direct substitution but can squeeze between $-x^2$ and $x^2$.

## Continuity

A function $f$ is **continuous at $a$** when all three of the following hold:

1. $f(a)$ is defined,
2. $\lim_{x \to a} f(x)$ exists,
3. $\lim_{x \to a} f(x) = f(a)$.

If any of those fails, $f$ is **discontinuous** at $a$. The three classic discontinuity types you should be able to name on sight are:

- **Removable** — limit exists but disagrees with (or doesn't equal) $f(a)$. Patching $f(a)$ fixes it.
- **Jump** — left and right limits both exist but disagree. Cannot be patched.
- **Infinite** — the function blows up (vertical asymptote).

## The Intermediate Value Theorem

If $f$ is continuous on $[a, b]$ and $N$ is any value strictly between $f(a)$ and $f(b)$, then there exists $c \in (a, b)$ with $f(c) = N$. In MATH 171 we mostly use this to argue that an equation has a root in an interval, e.g. "since $f(0) = -1$ and $f(2) = 5$, $f$ has a root in $(0, 2)$." The cheatsheet at {{FILE:wsu-math171-limits-cheatsheet}} has a worked IVT example you can model your homework on.

## Common mistakes

- Treating $0/0$ as if it equals $0$ or $1$. It is an *indeterminate form* — the limit could be anything.
- Forgetting that the quotient law requires $\lim g \ne 0$.
- Assuming a function is continuous because it is "smooth-looking." Continuity is a point-by-point property; piecewise definitions can break it.
- Writing $\lim_{x \to a} f(x) = f(a)$ as a definition instead of as the *conclusion* you reach when $f$ is continuous.

## Practice

When you can apply the limit laws without consulting them and recite the three conditions for continuity from memory, take {{QUIZ:math171-limits-and-continuity-quiz}}. Limits show up everywhere — they will reappear in {{GUIDE:math171-derivative-rules}} as the definition of $f'(x)$ and again in {{GUIDE:math171-integral-fundamentals}} as the definition of the definite integral.

For the broader course catalog see {{COURSE:wsu/math171}}.
