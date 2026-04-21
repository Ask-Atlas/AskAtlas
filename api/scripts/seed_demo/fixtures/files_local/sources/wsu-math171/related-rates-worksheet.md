---
slug: wsu-math171-related-rates-worksheet
title: "Related Rates Worksheet"
mime: application/vnd.openxmlformats-officedocument.wordprocessingml.document
filename: related-rates-worksheet.docx
course: wsu/math171
description: "Student worksheet with four related-rates problems — ladder, balloon, cone, and shadow — plus a solution checklist."
author_role: bot
---

# Related Rates Worksheet

**Instructions.** For each problem: (1) draw a labeled diagram, (2) list the given rates and the rate you want, (3) write an equation relating the variables, (4) differentiate both sides with respect to time $t$, (5) substitute known values **after** differentiating, and (6) solve.

Show all work. Include units in your final answer.

---

## Problem 1 — The Sliding Ladder

A 10-ft ladder leans against a vertical wall. The bottom of the ladder slides away from the wall at $\frac{dx}{dt} = 2$ ft/sec. How fast is the top sliding down the wall when the bottom is 6 ft from the wall?

Hints:
- Let $x$ = horizontal distance, $y$ = height on wall.
- Constraint: $x^2 + y^2 = 100$.
- Differentiate: $2x \frac{dx}{dt} + 2y \frac{dy}{dt} = 0$.
- Find $y$ from $x = 6$ (Pythagoras), then solve for $\frac{dy}{dt}$.

Your answer: $\frac{dy}{dt} = \underline{\phantom{xxx}}$ ft/sec.

---

## Problem 2 — The Spherical Balloon

Air is pumped into a spherical balloon at a rate of $\frac{dV}{dt} = 100$ cm$^3$/sec. How fast is the radius increasing when the diameter is 50 cm?

Hints:
- $V = \frac{4}{3}\pi r^3$, so $\frac{dV}{dt} = 4\pi r^2 \frac{dr}{dt}$.
- At diameter 50 cm, $r = 25$ cm.
- Solve for $\frac{dr}{dt}$.

Your answer: $\frac{dr}{dt} = \underline{\phantom{xxx}}$ cm/sec.

---

## Problem 3 — The Leaky Conical Tank

Water drains from an inverted conical tank with height 12 ft and top radius 6 ft at a rate of $\frac{dV}{dt} = -2$ ft$^3$/min. How fast is the water level falling when the water is 4 ft deep?

Hints:
- By similar triangles, $\frac{r}{h} = \frac{6}{12} = \frac{1}{2}$, so $r = h/2$.
- $V = \frac{1}{3}\pi r^2 h = \frac{1}{3}\pi (h/2)^2 h = \frac{\pi h^3}{12}$.
- $\frac{dV}{dt} = \frac{\pi h^2}{4} \cdot \frac{dh}{dt}$.
- Substitute $h = 4$, solve for $\frac{dh}{dt}$.

Your answer: $\frac{dh}{dt} = \underline{\phantom{xxx}}$ ft/min.

---

## Problem 4 — The Lamppost Shadow

A 6-ft-tall person walks away from a 15-ft lamppost at 5 ft/sec. How fast is the tip of her shadow moving along the ground?

Hints:
- Let $x$ = distance from post to person, $s$ = distance from post to shadow tip. Shadow length is $s - x$.
- Similar triangles: $\frac{15}{s} = \frac{6}{s - x}$, giving $15(s - x) = 6s$, so $s = \frac{15}{9}x = \frac{5}{3}x$.
- Differentiate: $\frac{ds}{dt} = \frac{5}{3} \cdot \frac{dx}{dt}$.

Your answer: $\frac{ds}{dt} = \underline{\phantom{xxx}}$ ft/sec.

---

## Self-Check

- [ ] Did you differentiate **before** substituting numbers?
- [ ] Did you include the correct sign for rates that are decreasing?
- [ ] Are your final units consistent (ft/sec vs cm/sec vs ft/min)?
- [ ] Is your answer's sign physically reasonable (top of ladder goes down, so $\frac{dy}{dt} < 0$)?

**Expected answers:** P1: $-\frac{3}{2}$ ft/sec. P2: $\frac{100}{2500\pi} = \frac{1}{25\pi}$ cm/sec. P3: $-\frac{1}{2\pi}$ ft/min. P4: $\frac{25}{3}$ ft/sec.
