---
slug: psych105-research-methods-and-ethics
course:
  ipeds_id: "236939"
  department: "PSYCH"
  number: "105"
title: "Research Methods and Ethics — PSYCH 105"
description: "Experimental, correlational, and descriptive methods plus the IRB ethics framework that gates every PSYCH 105 study you'll read."
tags: ["research-methods", "ethics", "experiments", "irb", "midterm"]
author_role: bot
quiz_slug: psych105-research-methods-and-ethics-quiz
attached_files:
  - wsu-psych105-research-ethics-primer
attached_resources: []
---

# Research Methods and Ethics

Almost every other unit in PSYCH 105 cites empirical findings — Skinner's
boxes, Loftus's misinformation studies, Ainsworth's Strange Situation. None
of those mean anything without the methodology that produced them. Spend
the first two weeks getting fluent in research design and the rest of the
semester gets dramatically easier, because you'll be able to read a result
and immediately ask the right follow-up question: *what kind of study was
this, and what can it actually tell us?*

## The three families of method

Psychology research falls into three broad camps. Each answers a different
question and carries different limits.

> **Descriptive** — what's happening? Case studies, naturalistic observation,
> surveys. High in ecological validity, but you can't infer causation and
> you may have selection or response bias.

> **Correlational** — do two variables move together? You measure both,
> compute a correlation coefficient (`r`), and report the strength and
> direction of the relationship. **Correlation is not causation** — a
> third variable, reverse causation, or coincidence can all produce the
> same `r`.

> **Experimental** — does X cause Y? You manipulate the **independent
> variable**, hold everything else constant, and measure the **dependent
> variable**. Random assignment to conditions controls for individual
> differences. Only experiments support causal claims.

Memorize that ladder. On the exam, the most common trap is being asked
whether a result implies causation when the underlying study was
correlational.

## Anatomy of an experiment

A clean experiment in PSYCH 105 looks like this:

1. Operationalize your variables. "Stress" isn't measurable — cortisol
   level or self-reported anxiety on a 1–7 scale is.
2. Recruit a sample large enough to detect the effect you care about.
3. Randomly assign participants to the experimental and control
   conditions.
4. Run a **double-blind** procedure when feasible — neither the
   participant nor the experimenter knows who's in which condition.
5. Measure the dependent variable.
6. Compare conditions with the appropriate statistical test.

A control condition is non-negotiable. Without one, you can't separate
the effect of your manipulation from placebo, expectancy, or simple
regression to the mean.

## Threats to validity

You'll hear two flavors all semester:

- **Internal validity** — can we conclude the IV caused the DV in *this*
  sample? Threatened by confounds, demand characteristics, and
  experimenter effects.
- **External validity** — do the results generalize to other people,
  settings, or times? Threatened by WEIRD samples (Western, Educated,
  Industrialized, Rich, Democratic — typically college sophomores) and
  artificial lab conditions.

A famous study can be high on one and low on the other. Milgram's
obedience research has strong internal validity (the manipulation
worked) but contested external validity (would people obey the same
way outside a Yale lab?).

## The ethics gate

The Institutional Review Board (IRB) reviews every study with human
participants before it runs. The framework you're responsible for in
PSYCH 105 traces back to the Belmont Report:

| Principle | What it means in practice |
|---|---|
| Respect for persons | Informed consent, voluntary participation, right to withdraw |
| Beneficence | Maximize benefit, minimize harm; weigh risks honestly |
| Justice | Burdens and benefits of research distributed fairly |

In operational terms, that translates to:

- **Informed consent** — participants know what they're agreeing to,
  in language they can understand. Children give *assent* and
  guardians give *consent*.
- **Debriefing** — after any deception, participants are told the
  true purpose and given a chance to ask questions.
- **Confidentiality** — data are stored without identifiers when
  possible, and identifiers are not shared.
- **Right to withdraw** — at any time, without penalty, even after
  the session has started.

The {{FILE:wsu-psych105-research-ethics-primer}} primer expands on
how IRB protocols are written and reviewed at WSU.

## Notorious cases that shaped the rules

- **Tuskegee Syphilis Study (1932–1972)** — withheld penicillin from
  Black sharecroppers without consent. Drove the National Research
  Act and the Belmont Report.
- **Milgram (1961)** — deception about electric shocks. Drove
  modern debriefing requirements.
- **Stanford Prison Experiment (1971)** — failure of researcher
  oversight; participants harmed. Informs current limits on
  deception and the requirement that the experimenter is not also
  the rescuer.
- **Little Albert (1920)** — conditioned a fear response in an
  infant with no consent and no extinction protocol. Modern IRB
  would never approve.

## Sampling and generalization

A result from a sample of 50 WSU undergrads tells you something about
WSU undergrads. Whether it tells you anything about, say, retired
farmers in rural Idaho is an empirical question — usually not
answered in the original paper. Watch for:

- **Convenience sampling** — easy but biased. Most psych studies
  use it.
- **Random sampling** — every member of the population has an equal
  chance of being selected. Rare in lab studies, common in surveys.
- **Stratified sampling** — random within demographic strata; used
  when you want representative subgroups.

## Statistical literacy in one paragraph

You don't need to compute t-tests by hand for the midterm, but you
do need to read a result and know what `p < .05` means: assuming the
null hypothesis is true, the probability of observing data this
extreme is less than 5%. It does **not** mean the effect is large,
nor that there's a 95% chance the result is "real". Effect size
(Cohen's `d`, `r²`) tells you about magnitude; `p` tells you about
unlikeliness under the null.

## Replication and the "crisis"

Around 2015, large-scale replication projects found that many
classic psychology results don't reproduce. The field's response
has been preregistration (declaring your hypothesis and analysis
plan before collecting data), open data, and larger samples. When
you read a study from before about 2012, treat the reported
effect as a candidate finding, not a settled fact.

## Study cues for the exam

- If the question says "researchers found X is associated with Y",
  the design is correlational. Don't be tricked into a causal
  conclusion.
- If the question shows random assignment to conditions, it's an
  experiment, and causal language is appropriate.
- If the question describes deception, look for whether debriefing
  is mentioned — that's usually the trap.

For practice on these distinctions, run {{QUIZ:psych105-research-methods-and-ethics-quiz}}
before moving on to the conditioning unit. The methods you learn
here are the lens through which we'll evaluate every learning,
memory, and social-psych study in {{COURSE:wsu/psych105}}.
