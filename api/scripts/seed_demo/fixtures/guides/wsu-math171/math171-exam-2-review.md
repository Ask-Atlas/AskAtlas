---
slug: math171-exam-2-review
course:
  ipeds_id: "236939"
  department: "MATH"
  number: "171"
title: "Exam 2 Review — MATH 171"
description: "Cumulative review for the second midterm: chain rule, implicit differentiation, related rates, optimization, and the FTC."
tags: ["calculus", "exam-review", "derivatives", "ftc", "midterm"]
author_role: bot
quiz_slug: math171-exam-2-review-quiz
attached_files:
  - wsu-math171-derivative-rules
  - wsu-math171-related-rates-worksheet
  - wsu-math171-optimization-slides
  - wsu-math171-ftc-summary
attached_resources: []
---

# Exam 2 Review — Chain Rule, Implicit Diff, Related Rates, Optimization, FTC

Exam 2 is the broadest test of the semester. It covers the second half of the differentiation material, all of the application word problems, and the first taste of integration. Six questions, 75 minutes, no calculator. The bottleneck is rarely the calculus — it is being organised enough to set up word problems quickly.

## What is on the exam

| Topic | Approx. weight | Source guide |
|---|---|---|
| Chain rule (and combined with product/quotient) | 15% | {{GUIDE:math171-chain-rule}} |
| Implicit differentiation + tangent lines | 15% | {{GUIDE:math171-implicit-differentiation}} |
| Related rates word problem | 20% | {{GUIDE:math171-related-rates}} |
| Optimization word problem | 25% | {{GUIDE:math171-optimization}} |
| Indefinite and definite integrals | 15% | {{GUIDE:math171-integral-fundamentals}} |
| FTC Part 1 (differentiating an integral) | 10% | {{GUIDE:math171-ftc-summary}} |

## Diagnostic — can you do these in 30 seconds each?

1. $\frac{d}{dx}\left[\sin(3x^2)\right]$
2. Find $dy/dx$ if $x^2 + xy + y^2 = 7$.
3. $\frac{d}{dx} \int_0^{x^3} \cos t\, dt$.
4. $\int_0^1 (3 x^2 + e^x)\, dx$.

If any take longer than 30 seconds, re-read the corresponding guide before continuing. Answers (in order): $6x \cos(3x^2)$; $-(2x + y)/(x + 2y)$; $3 x^2 \cos(x^3)$; $1 + e - 1 = e$.

## Word-problem playbook

### Related rates

Use the four-step recipe from {{GUIDE:math171-related-rates}}: picture, equation, differentiate w.r.t. $t$, plug in. The worksheet at {{FILE:wsu-math171-related-rates-worksheet}} has a dozen practice problems. The traps are (a) substituting numbers before differentiating and (b) forgetting to use a similar-triangles relation to reduce variables.

### Optimization

Use the six-step recipe from {{GUIDE:math171-optimization}}: identify objective, write it as a function, use constraint to reduce to one variable, find critical points, confirm with the second-derivative test, answer the original question. The slides at {{FILE:wsu-math171-optimization-slides}} have worked solutions to the most common templates: rectangle in a region, can/box volume, distance minimisation.

## Chain rule and implicit — the two-minute drill

These two topics share one trap: forgetting the inside derivative. Practice these:

- $\frac{d}{dx}\left[e^{\sin x}\right] = e^{\sin x} \cos x$.
- $\frac{d}{dx}\left[\ln(x^2 + 1)\right] = \frac{2x}{x^2 + 1}$.
- $\frac{d}{dx}\left[\sqrt{1 + \tan x}\right] = \frac{\sec^2 x}{2 \sqrt{1 + \tan x}}$.
- For $x^2 y + y^3 = 10$: differentiate to get $2xy + x^2 y' + 3 y^2 y' = 0$, solve $y' = -2xy/(x^2 + 3y^2)$.

Cover the right-hand sides and reproduce them. If you can write each derivative in under 20 seconds, you are ready.

## FTC drill

For Part 2 (evaluating definite integrals): rewrite the integrand if needed, find an antiderivative, plug in the limits.

For Part 1 (differentiating an integral with a variable upper limit): use $\frac{d}{dx} \int_a^{u(x)} f(t)\, dt = f(u(x)) \cdot u'(x)$. The cheatsheet at {{FILE:wsu-math171-ftc-summary}} has the canonical examples.

## Time budget

For a 75-minute exam:

- 10 minutes — quick computational problems (chain rule, implicit, basic integrals).
- 25 minutes — the related-rates and optimization word problems (about 12 minutes each).
- 20 minutes — the FTC and definite integral problems.
- 20 minutes — buffer to re-check.

If you go over budget on any one problem, *move on*. Coming back fresh is faster than grinding.

## Common mistakes (cumulative list)

- Substituting numbers before differentiating in related-rates problems.
- Forgetting the constraint when setting up an optimization problem.
- Treating $\infty - \infty$ or $0 \cdot \infty$ as numbers — they are indeterminate.
- Forgetting $+ C$ on indefinite integrals.
- Confusing the integration variable $t$ with the limit variable $x$ in FTC Part 1 problems.

## Practice

When you finish this review, take {{QUIZ:math171-exam-2-review-quiz}} under exam conditions (closed-book, 25 minutes). After that, the related-rates and optimization quizzes are the most worthwhile additional drills.

Course catalog: {{COURSE:wsu/math171}}.
