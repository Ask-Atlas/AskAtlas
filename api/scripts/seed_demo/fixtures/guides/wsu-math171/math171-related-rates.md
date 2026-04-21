---
slug: math171-related-rates
course:
  ipeds_id: "236939"
  department: "MATH"
  number: "171"
title: "Related Rates — MATH 171"
description: "A four-step recipe for related-rates word problems with worked ladders, balloons, and shadows."
tags: ["calculus", "derivatives", "related-rates", "word-problems"]
author_role: bot
quiz_slug: math171-related-rates-quiz
attached_files:
  - wsu-math171-related-rates-worksheet
  - wsu-math171-derivative-rules
attached_resources: []
---

# Related Rates

Related-rates problems are the section of MATH 171 that students dread the most, almost entirely because the *setup* is the hard part. Once the equation is on the page and you have differentiated with respect to $t$, the rest is bookkeeping. This guide gives you a recipe and four worked problems.

## The four-step recipe

1. **Draw a picture and label every quantity that varies.** Use letters, not numbers — even quantities that *happen* to equal $5$ at the moment of interest must be a letter until the very last step.
2. **Write an equation that relates the variables.** This is geometry: Pythagoras, similar triangles, area or volume formulas, trig identities.
3. **Differentiate both sides with respect to $t$.** Every variable that depends on $t$ contributes a $\frac{d}{dt}$ factor — this is implicit differentiation in disguise (revisit {{GUIDE:math171-implicit-differentiation}} if needed).
4. **Plug in the instantaneous values and solve for the requested rate.** Numerical substitution happens *after* differentiation, never before.

The most common mistake on the midterm is doing step 4 first — substituting numbers into the equation before differentiating. Constants have derivative $0$, so if you do that you will get nonsense.

## Worked problem 1: the sliding ladder

A 10-ft ladder is leaning against a wall. The bottom is sliding away from the wall at $1$ ft/s. How fast is the top sliding down when the bottom is $6$ ft from the wall?

**Setup.** Let $x$ be the distance from the wall to the foot of the ladder and $y$ be the height of the top. By Pythagoras, $x^2 + y^2 = 100$.

**Differentiate.** $2x \frac{dx}{dt} + 2y \frac{dy}{dt} = 0$, so $\frac{dy}{dt} = -\frac{x}{y} \frac{dx}{dt}$.

**Plug in.** When $x = 6$, Pythagoras gives $y = 8$. With $dx/dt = 1$,

$$
\frac{dy}{dt} = -\frac{6}{8} \cdot 1 = -\frac{3}{4}\ \text{ft/s}.
$$

The negative sign means the top is moving *down*, as expected.

## Worked problem 2: the balloon

A spherical balloon is being inflated at $20$ cubic cm per second. How fast is the radius increasing when the radius is $5$ cm?

**Setup.** $V = \frac{4}{3}\pi r^3$.

**Differentiate.** $\frac{dV}{dt} = 4\pi r^2 \frac{dr}{dt}$.

**Plug in.** $20 = 4\pi (25) \frac{dr}{dt}$, so $\frac{dr}{dt} = \frac{1}{5\pi}\ \text{cm/s}$.

## Worked problem 3: similar triangles (the shadow)

A 6-ft tall person walks away from a 15-ft lamppost at $4$ ft/s. How fast is the tip of the shadow moving?

**Setup.** Let $x$ be the distance from the lamppost to the person, and $s$ be the length of the shadow. By similar triangles,

$$
\frac{15}{x + s} = \frac{6}{s} \quad \Longrightarrow \quad 15 s = 6(x + s) \quad \Longrightarrow \quad 9s = 6x \quad \Longrightarrow \quad s = \tfrac{2}{3} x.
$$

**Differentiate.** $\frac{ds}{dt} = \tfrac{2}{3} \frac{dx}{dt} = \tfrac{2}{3}(4) = \tfrac{8}{3}$ ft/s.

The tip of the shadow moves at $\frac{dx}{dt} + \frac{ds}{dt} = 4 + \tfrac{8}{3} = \tfrac{20}{3}$ ft/s. Note: the *shadow tip* is moving faster than the person, but the shadow's *length* grows at only $\tfrac{8}{3}$ ft/s.

## Worked problem 4: the cone

Water drains from an inverted cone (radius 3, height 6) at $2$ cubic m/min. How fast is the water level dropping when the depth is $4$ m?

**Setup.** Let $h$ be the depth of the water and $r$ the radius of the water surface. By similar triangles, $r/h = 3/6$, so $r = h/2$. The volume of the water cone is

$$
V = \tfrac{1}{3}\pi r^2 h = \tfrac{1}{3}\pi \left(\tfrac{h}{2}\right)^2 h = \tfrac{\pi}{12} h^3.
$$

Crucially, we eliminated $r$ *before* differentiating, so we only have one variable.

**Differentiate.** $\frac{dV}{dt} = \tfrac{\pi}{4} h^2 \frac{dh}{dt}$.

**Plug in.** $-2 = \tfrac{\pi}{4}(16)\frac{dh}{dt}$, so $\frac{dh}{dt} = -\tfrac{1}{2\pi}\ \text{m/min}$.

The negative sign reflects that water is leaving, so depth decreases. Use $dV/dt = -2$ because the volume is *decreasing*.

## More practice

The worksheet at {{FILE:wsu-math171-related-rates-worksheet}} has 12 additional problems with answers. Spend an hour with it before the midterm — the patterns repeat.

## Common mistakes

- Substituting numbers before differentiating.
- Forgetting to use a similar-triangles relation to reduce to one variable when the geometry has two related lengths (cones, shadows).
- Forgetting the chain rule. Every variable that changes in time contributes a $d\!/dt$ factor.
- Sign errors. If a quantity is decreasing, its rate is negative.

## Practice

When you can recite the four-step recipe and reproduce the ladder problem from a blank sheet, take {{QUIZ:math171-related-rates-quiz}}.

Course catalog: {{COURSE:wsu/math171}}.
