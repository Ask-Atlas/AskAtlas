---
slug: psych1-learning-conditioning
course:
  ipeds_id: "243744"
  department: "PSYCH"
  number: "1"
title: "Learning and Conditioning — PSYCH 1"
description: "Classical and operant conditioning — the rules behind how organisms (including you) update behaviour from experience."
tags: ["learning", "conditioning", "behavioural-psychology", "reinforcement", "midterm"]
author_role: bot
quiz_slug: psych1-learning-conditioning-quiz
attached_files:
  - stanford-psych1-diagnostic-glossary
attached_resources: []
---

# Learning and Conditioning

Learning, in the technical sense psychology uses the word, means a
*relatively durable change in behaviour or knowledge that results from
experience*. The change has to be lasting (not just transient
fatigue), and it has to be caused by experience (not by maturation or
injury). Almost every change in your behaviour over the last year fits
that definition.

Two paradigms have shaped how the field studies learning:
**classical** (Pavlovian) conditioning and **operant** (instrumental)
conditioning. They look superficially similar but capture different
relationships, and the distinction is on every PSYCH 1 exam.

## Classical conditioning

Classical conditioning is about *predictive associations* between
stimuli. The organism is essentially passive — it doesn't have to do
anything to be conditioned.

Pavlov's setup, in the canonical vocabulary:

- **Unconditioned stimulus (US)** — meat powder. Triggers salivation
  reflexively.
- **Unconditioned response (UR)** — salivation to meat powder. The
  built-in reflex.
- **Neutral stimulus** — bell. Doesn't initially produce salivation.
- **Conditioned stimulus (CS)** — the bell, *after* it has been paired
  with meat powder repeatedly.
- **Conditioned response (CR)** — salivation to the bell alone.

The CR is similar to but typically smaller than the UR. Several
phenomena follow from the basic procedure:

- **Acquisition** — the conditioning curve as CS-US pairings are
  presented.
- **Extinction** — present the CS without the US repeatedly; the CR
  diminishes. The association is *suppressed*, not erased.
- **Spontaneous recovery** — after extinction and a delay, the CR
  returns at reduced strength when the CS is presented again.
- **Generalisation** — stimuli similar to the CS also elicit the CR.
  Pavlov's dogs salivated to bells of similar pitch.
- **Discrimination** — with differential training (one CS paired,
  another not), the response narrows to the predictive stimulus.

> The textbook gloss "the bell predicts food" is closer to the truth
> than "the bell substitutes for food." Classical conditioning is the
> brain learning what predicts what.

### Higher-order conditioning, taste aversion, and biological prep

A few classic findings push the basic story:

- **Higher-order (second-order) conditioning** — once a CS reliably
  predicts the US, you can pair a *new* neutral stimulus with the
  established CS, and the new stimulus also acquires the CR. The chain
  weakens fast — third-order conditioning is fragile.
- **Conditioned taste aversion (Garcia effect)** — pair a flavour with
  illness *once*, even hours later, and the organism avoids that
  flavour for life. Classical conditioning isn't a uniform rule that
  treats all CS-US pairs equally; biology has prepared certain
  associations to learn faster than others. A flashing light paired
  with nausea doesn't condition the same way.
- **Watson's Little Albert** — a small child was conditioned to fear a
  white rat by pairing it with a loud noise, with generalisation to
  other furry stimuli. Ethically indefensible by modern standards, but
  formative for the field.

## Operant conditioning

Operant conditioning is about *consequences*. The organism emits a
behaviour, and what happens next changes the future probability of
that behaviour.

Thorndike's **law of effect** states the core idea: behaviours followed
by satisfying consequences are more likely to recur; behaviours
followed by aversive consequences are less likely. Skinner formalised
this into a 2×2 grid that you should be able to draw from memory.

|                    | Add stimulus            | Remove stimulus            |
|--------------------|-------------------------|----------------------------|
| **Increases behaviour** | Positive reinforcement | Negative reinforcement |
| **Decreases behaviour** | Positive punishment    | Negative punishment    |

"Positive" and "negative" in this grid mean *added* and *removed*, not
"good" and "bad."

- **Positive reinforcement** — give a treat after the dog sits. Sitting
  goes up.
- **Negative reinforcement** — fasten the seatbelt to silence the
  beeping. Buckling goes up. (This is *not* punishment. Removing an
  aversive stimulus to *increase* behaviour is reinforcement.)
- **Positive punishment** — yelp when the puppy bites. Biting goes down.
- **Negative punishment** — confiscate the phone when the teen breaks
  curfew. Curfew-breaking goes down.

The single most common student error here is calling negative
reinforcement "punishment." If a procedure makes the behaviour *more*
likely, it is reinforcement, by definition.

### Schedules of reinforcement

How often you reinforce matters as much as whether you reinforce.

- **Continuous reinforcement** — reward every correct response.
  Acquisition is fast, extinction is also fast.
- **Partial / intermittent reinforcement** — reward only some
  responses. Slower acquisition, but much more *resistant to
  extinction*. Variable schedules in particular generate behaviour
  that persists for a very long time without reward.

The four main partial schedules:

| Schedule | Rule | Typical pattern |
|---|---|---|
| Fixed ratio (FR) | Reward every Nth response | High rate, brief post-reinforcement pause |
| Variable ratio (VR) | Reward after average N responses | Highest rate, very steady — slot machines |
| Fixed interval (FI) | Reward first response after N seconds | Scallop pattern: low then high near deadline |
| Variable interval (VI) | Reward first response after average N seconds | Steady, moderate rate |

Variable-ratio schedules produce the most extinction-resistant
behaviour we know how to generate, which is the technical reason
gambling is so behaviourally sticky.

### Shaping

For complex behaviours the organism never spontaneously emits in full,
**shaping** trains successive approximations: reinforce any move toward
the goal behaviour, then progressively narrow the criterion. The
trained-pigeon-playing-ping-pong demos rest on this.

## Cognitive and social learning

Pure stimulus-response accounts can't cover everything. Two
complications worth knowing:

- **Latent learning** (Tolman) — rats allowed to wander a maze with no
  reward later learn the maze faster than rats with no exploration
  experience, suggesting they had built a *cognitive map* even without
  reinforcement.
- **Observational learning** (Bandura) — children who watched an adult
  attack a Bobo doll were more aggressive toward it themselves. Humans
  learn powerful behavioural repertoires by watching others, with no
  direct reinforcement of their own.

These extend rather than overturn the conditioning framework. Most
modern accounts treat learning as inherently *associative and
predictive* — closer to the Rescorla-Wagner formalism, which models
classical conditioning as the brain reducing prediction error
between expected and actual outcomes.

For the precise vocabulary you'll see on test items, the
{{FILE:stanford-psych1-diagnostic-glossary}} cross-references CS / US /
CR / UR with the operant grid.

## Where this comes back

Conditioning is the substrate behind much of {{GUIDE:psych1-memory}}
(implicit memory in particular), and the predictive-error story
connects directly to dopamine in {{GUIDE:psych1-neuroscience-primer}}.
Many real-world examples in {{GUIDE:psych1-social-psychology}} —
prejudice acquisition, advertising — are partly classical conditioning
in the wild.

Once you can fill in Skinner's 2×2 from a blank page and explain why
variable-ratio schedules are extinction-resistant, take
{{QUIZ:psych1-learning-conditioning-quiz}}.
