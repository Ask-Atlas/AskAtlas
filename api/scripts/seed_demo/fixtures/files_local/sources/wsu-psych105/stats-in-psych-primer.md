---
slug: wsu-psych105-stats-in-psych-primer
title: "Statistics in Psychology: A Primer"
mime: application/pdf
filename: stats-in-psych-primer.pdf
course: wsu/psych105
description: "Primer on the descriptive and inferential statistics behind psychology research: distributions, t-tests, effect size, p-values."
author_role: bot
---

# Statistics in Psychology: A Primer

Psychology is an empirical science, which means every claim in your textbook rests on a number someone computed. This primer introduces the statistics you will see in Psych 105 readings — enough to read a Results section without glazing over.

## Descriptive Statistics

Descriptives summarize a sample. Three families matter:

- **Central tendency**: mean (`$\bar{x}$`), median, mode. The median resists outliers; the mean does not.
- **Variability**: range, variance (`$s^2$`), standard deviation (`$s$`). Standard deviation is in the same units as the data — always report it.
- **Distribution shape**: skew (asymmetry) and kurtosis (tail heaviness).

The **normal distribution** (bell curve) is the workhorse. About 68% of values fall within 1 SD of the mean, 95% within 2 SD, 99.7% within 3 SD — the empirical rule.

## Z-Scores

A z-score converts a raw value into "standard deviations above the mean":

`$z = \frac{x - \mu}{\sigma}$`

A z of `$+1.96$` or lower than `$-1.96$` is in the outer 5% of a normal distribution. Z-scores let you compare scores across tests on different scales (an IQ of 130 is `$z = +2$`; an SAT of 1400 is roughly the same z).

## Sampling and the Central Limit Theorem

We never measure whole populations — we sample. The **central limit theorem** says that if you take many samples of size `$n$` from any population, the distribution of sample means approaches normal as `$n$` grows, with standard error `$\sigma / \sqrt{n}$`. This is why larger studies are more trustworthy.

## Hypothesis Testing: The t-Test

A t-test asks: are two group means different by more than chance?

- **Null hypothesis (`$H_0$`)**: no difference.
- **Alternative (`$H_1$`)**: there is a difference.
- **p-value**: probability of observing a difference this large *if `$H_0$` were true*.

Convention: if `$p < 0.05$`, reject `$H_0$` and call the effect "statistically significant". A p-value is **not** the probability that `$H_0$` is true, and **not** a measure of effect size.

## Effect Size: Cohen's d

Statistical significance tells you whether an effect exists; effect size tells you how big it is.

`$d = \frac{\bar{x}_1 - \bar{x}_2}{s_{pooled}}$`

Cohen's rule of thumb: `$d = 0.2$` small, `$0.5$` medium, `$0.8$` large. A significant `$p$` with a tiny `$d$` usually means a huge sample, not an important finding.

## Correlation vs. Causation

Pearson's `$r$` ranges from `$-1$` (perfect negative) to `$+1$` (perfect positive). `$r = 0.3$` is typical in social science; `$r^2$` tells you the proportion of variance explained. Correlation does not imply causation because of:

- **Third variables** (ice-cream sales and drowning both rise in summer)
- **Reverse causation** (does exercise cause happiness, or vice versa?)
- **Selection effects** (who volunteered for the study?)

Only randomized experiments license causal claims.

## The Replication Crisis

Since roughly 2011, psychology has been reckoning with the fact that many classic findings fail to replicate. Causes: p-hacking, publication bias, low-powered studies. Responses: pre-registration, open data, larger samples, and reporting effect sizes alongside p-values. When you read a single study, ask: has this been replicated?

## Minimum Literacy Checklist

- Can you state `$H_0$` and `$H_1$` for the study?
- Is the sample size reported and adequate?
- Are effect sizes (not just p-values) given?
- Is the design experimental or correlational?
- Has the finding replicated?
