---
slug: psych1-research-methods
course:
  ipeds_id: "243744"
  department: "PSYCH"
  number: "1"
title: "Research Methods — PSYCH 1"
description: "How psychologists actually know things: experiments, correlation, sampling, and the difference between a finding and a vibe."
tags: ["research-methods", "statistics", "experiments", "intro", "midterm"]
author_role: bot
attached_files:
  - stanford-psych1-research-methods-worksheet
  - stanford-psych1-diagnostic-glossary
attached_resources: []
---

# Research Methods in Psychology

Psychology is a science, which means every claim in this course rests on
evidence collected through a method. The single most useful thing PSYCH 1
can teach you is how to read a study and say, with calibrated confidence,
*how much should I update on this?* That habit is more durable than any
specific finding.

This guide is the toolkit. Every later topic — perception, learning,
memory, social — uses these methods, so it pays to get them straight now.

## The basic question every study answers

> An experiment is a procedure for asking nature a *yes/no* question
> about cause: does manipulating X change Y, holding everything else
> equal?

Most other research designs (correlational, observational, longitudinal,
case study) answer weaker questions — usually about *association* rather
than *cause*. They are not lesser; they are appropriate to different
problems. But you should always know which question a study answered
before you decide what it shows.

## Variables and operationalisation

- **Independent variable (IV):** what the experimenter manipulates.
- **Dependent variable (DV):** what gets measured.
- **Operational definition:** the specific, measurable procedure that
  stands in for an abstract construct ("aggression" → "number of hot
  sauce drops added to a stranger's drink in the lab paradigm").

Operational definitions matter because they determine what the study
*actually* measured. Two studies that both claim to measure "memory" but
operationalise it differently (recall vs. recognition vs. priming) may
genuinely disagree without contradicting each other.

## Experimental design

The gold-standard intervention study has four ingredients:

1. **Random assignment** of participants to conditions. This is what
   gives you causal inference: it makes the groups equivalent on
   *everything* in expectation, including variables you didn't measure
   or even think of.
2. **A control condition** that differs from the treatment only in the
   IV.
3. **Blinding** when feasible — participants shouldn't know which
   condition they're in (single-blind), and ideally neither should the
   experimenters interacting with them (double-blind).
4. **A pre-registered analysis plan** so that the test you run is the
   test you said you would run.

Without random assignment, you have a quasi-experiment at best. Without
a control, you have a demonstration. Without blinding, you have a
demand-characteristics problem. These aren't pedantic distinctions;
they're the load-bearing structure of any causal claim.

## Correlational designs

Sometimes you can't randomly assign. You can't randomly assign people to
"smoker" or "non-smoker" — so the lung-cancer evidence is correlational,
not experimental, and the case had to be built across many studies and
mechanisms. Two big rules apply:

- **Correlation does not equal causation.** A correlation between X and
  Y is consistent with X→Y, Y→X, a third variable Z causing both, or
  selection effects. The correlation alone does not pick between them.
- **Effect sizes matter more than p-values.** A correlation of r = 0.05
  can be "statistically significant" in a giant sample and still
  predict almost nothing about an individual.

## Sampling

> A sample is only as good as the population it was drawn from and the
> mechanism that selected it.

A study of 250 Stanford undergraduates is informative about Stanford
undergraduates. It is suggestive, not conclusive, about humans in
general — and the gap matters whenever the construct is plausibly
shaped by culture, age, education, or wealth. The acronym **WEIRD**
(Western, Educated, Industrialised, Rich, Democratic) flags samples
that may not generalise widely; a striking share of psychology's
historical findings come from such samples.

Other sampling concerns to watch for:

- **Self-selection:** who chose to participate, and how does that
  differ from the people who didn't?
- **Survivorship:** which participants made it into the final dataset
  versus dropping out partway through?
- **Ceiling and floor effects:** if everyone scores at the maximum, you
  cannot detect a difference even if one exists.

## Statistics in one page

You don't need a statistics class to read a paper. You need three
intuitions.

1. **Variation is everywhere.** Two random samples from the same
   population will differ. Statistics tells you whether the difference
   you observed is bigger than the difference you'd expect from noise
   alone.
2. **A p-value is the probability of seeing a result *at least this
   extreme*, under the assumption that the null hypothesis (no real
   effect) is true.** It is not the probability that the hypothesis is
   true, and a p < 0.05 result is not "proven."
3. **Confidence intervals are usually more informative than p-values.**
   "The effect is somewhere between 0.2 and 1.4 standard deviations"
   tells you the size *and* the uncertainty in one shot.

For symbol definitions and a refresher on the common test names,
{{FILE:stanford-psych1-diagnostic-glossary}} has the cheat-sheet.

## Common pitfalls in interpreting psychology

- **Replication crisis (≈ 2011–present):** many landmark social and
  cognitive findings have failed to replicate at full effect size.
  Treat single-study, surprising results as provisional.
- **HARKing** — Hypothesising After Results are Known. Reporting
  exploratory findings as if they were predicted inflates false
  positives.
- **p-hacking** — running many small analyses until one crosses
  p < 0.05. The fix is pre-registration of the planned analysis.
- **File-drawer effect** — null results often go unpublished, biasing
  the published literature toward positive findings.
- **Reverse causation in longitudinal data** — if depression predicts
  later social withdrawal, withdrawal might equally predict later
  depression. Direction matters.

## A short reading rubric

When you read a study summary in this course (or in the news), walk it
through this checklist:

- [ ] What was the IV and the DV, in operational terms?
- [ ] Was there random assignment? If not, what alternative
      explanations does the design fail to rule out?
- [ ] What was the sample, and how was it selected?
- [ ] How big was the effect, not just whether it was significant?
- [ ] Has it replicated?

The companion worksheet at
{{FILE:stanford-psych1-research-methods-worksheet}} drills these
questions on five published study abstracts. Doing it once will change
how you read every study after.

## Where this comes back

- {{GUIDE:psych1-neuroscience-primer}} cites correlational neuroimaging
  evidence — the operational-definition lens applies directly.
- {{GUIDE:psych1-social-psychology}} discusses several classic studies
  whose modern replication results are mixed.
- The exam-prep guide ({{GUIDE:psych1-exam-prep}}) returns to this
  rubric as a way of thinking about applied questions.

For the catalog overview of this course, see
{{COURSE:stanford/psych1}}.
