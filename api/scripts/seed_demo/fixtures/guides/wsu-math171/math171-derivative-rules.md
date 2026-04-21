---
slug: math171-derivative-rules
course:
  ipeds_id: "236939"
  department: "MATH"
  number: "171"
title: "Derivative Rules — MATH 171"
description: "Power, sum, product, and quotient rules with the standard library of derivatives you must memorise."
tags: ["calculus", "derivatives", "rules", "midterm"]
author_role: bot
quiz_slug: math171-derivative-rules-quiz
attached_files:
  - wsu-math171-derivative-rules
  - wsu-math171-limits-cheatsheet
attached_resources: []
---

# Derivative Rules

Once you accept that $f'(x)$ is the slope of the tangent line at $x$, almost all of MATH 171 is choosing the right shortcut so you never actually have to compute

$$
f'(x) = \lim_{h \to 0} \frac{f(x + h) - f(x)}{h}
$$

by hand. This guide covers the algebraic rules. The chain rule gets its own treatment in {{GUIDE:math171-chain-rule}}.

## The library you must memorise

By the end of week three, the following derivatives should be reflex. Test yourself against {{FILE:wsu-math171-derivative-rules}}.

| Function | Derivative |
|---|---|
| $c$ (constant) | $0$ |
| $x^n$ | $n x^{n-1}$ |
| $e^x$ | $e^x$ |
| $a^x$ | $a^x \ln a$ |
| $\ln x$ | $1/x$ |
| $\log_a x$ | $1/(x \ln a)$ |
| $\sin x$ | $\cos x$ |
| $\cos x$ | $-\sin x$ |
| $\tan x$ | $\sec^2 x$ |
| $\sec x$ | $\sec x \tan x$ |
| $\arcsin x$ | $1/\sqrt{1 - x^2}$ |
| $\arctan x$ | $1/(1 + x^2)$ |

## The power rule

For any real number $n$,

$$
\frac{d}{dx}\left[x^n\right] = n x^{n-1}.
$$

This works for negative and fractional exponents too, which is the whole reason rewriting $\sqrt{x}$ as $x^{1/2}$ and $1/x^3$ as $x^{-3}$ is such a common opening move on exam problems.

**Example.**

$$
\frac{d}{dx}\left[3x^4 - \frac{2}{x^2} + 5\sqrt{x}\right] = 12x^3 + \frac{4}{x^3} + \frac{5}{2\sqrt{x}}.
$$

## Sum, difference, and constant multiple

These three are the easy ones — they let you differentiate term by term.

$$
\frac{d}{dx}\left[c \cdot f(x)\right] = c \cdot f'(x), \qquad \frac{d}{dx}\left[f \pm g\right] = f' \pm g'.
$$

There is no analogous "split rule" for products or quotients — those need their own machinery.

## The product rule

$$
\frac{d}{dx}\left[f(x) g(x)\right] = f'(x) g(x) + f(x) g'(x).
$$

A useful mnemonic: *first derivative of one times the other, plus the other way around*. The most common error is to write $(fg)' = f'g'$, which is wrong — try it on $x \cdot x = x^2$ and you get $1 \cdot 1 = 1$ instead of $2x$.

**Example.**

$$
\frac{d}{dx}\left[x^2 \sin x\right] = 2x \sin x + x^2 \cos x.
$$

## The quotient rule

$$
\frac{d}{dx}\left[\frac{f(x)}{g(x)}\right] = \frac{f'(x) g(x) - f(x) g'(x)}{[g(x)]^2}.
$$

Mnemonic: *low D-high minus high D-low, all over low squared*. The minus sign and the order of subtraction are the two things students get wrong most often. Always write the denominator squared first so you do not forget it.

**Example.**

$$
\frac{d}{dx}\left[\frac{x^2 + 1}{x - 3}\right] = \frac{2x(x - 3) - (x^2 + 1)(1)}{(x - 3)^2} = \frac{x^2 - 6x - 1}{(x - 3)^2}.
$$

## Why the rules work — a single example

The product rule can be derived from the limit definition:

$$
\begin{aligned}
(fg)'(x) &= \lim_{h \to 0} \frac{f(x+h) g(x+h) - f(x) g(x)}{h} \\
&= \lim_{h \to 0} \frac{f(x+h) g(x+h) - f(x) g(x+h) + f(x) g(x+h) - f(x) g(x)}{h} \\
&= \lim_{h \to 0} \left[ \frac{f(x+h) - f(x)}{h} g(x+h) + f(x) \frac{g(x+h) - g(x)}{h} \right] \\
&= f'(x) g(x) + f(x) g'(x).
\end{aligned}
$$

The trick is the "add and subtract $f(x) g(x+h)$" step. You do not need to reproduce this on the midterm, but seeing it once makes the rule less mysterious.

## Higher derivatives

The second derivative $f''(x)$ is just $\frac{d}{dx}\left[f'(x)\right]$. The signs tell you about concavity: $f'' > 0$ means concave up, $f'' < 0$ means concave down. We will lean on this in {{GUIDE:math171-optimization}}.

## Common mistakes

- Forgetting that $\frac{d}{dx}[\sin x] = \cos x$ requires $x$ to be in radians. The formulas in this guide assume radian measure.
- Trying to use the power rule on $a^x$ — it does not apply, because the exponent (not the base) is the variable. The correct derivative is $a^x \ln a$.
- Mixing up $(fg)' = f'g'$ (wrong) with the actual product rule.
- Dropping the minus sign in the quotient rule.

## Practice

Once you can recite the table from memory and apply the product/quotient rules without looking, take {{QUIZ:math171-derivative-rules-quiz}}. Then graduate to {{GUIDE:math171-chain-rule}}, which is the rule you will use most often in real problems.

Course landing page: {{COURSE:wsu/math171}}.
