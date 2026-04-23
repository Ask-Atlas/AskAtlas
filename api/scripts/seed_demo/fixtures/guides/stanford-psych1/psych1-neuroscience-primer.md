---
slug: psych1-neuroscience-primer
course:
  ipeds_id: "243744"
  department: "PSYCH"
  number: "1"
title: "Neuroscience Primer — PSYCH 1"
description: "Neurons, neurotransmitters, and the major brain regions you need for the first midterm."
tags: ["neuroscience", "biology", "brain", "neurons", "midterm"]
author_role: bot
quiz_slug: psych1-neuroscience-primer-quiz
attached_files:
  - stanford-psych1-neuroscience-primer
  - unsplash-brain-scan-6
attached_resources: []
---

# Neuroscience Primer

Psychology lives in the body. Every emotion, memory, and decision in
this course corresponds to electrical and chemical events inside about
86 billion neurons. You don't need to memorise the entire brain for
PSYCH 1 — but you do need a working mental model of the cell, the
synapse, and the major regions, because the rest of the course
constantly refers back to them.

## The neuron in 60 seconds

A neuron is a specialised cell with three jobs: receive signals,
integrate them, and decide whether to send a signal of its own.

- **Dendrites** — branching structures that receive signals from other
  neurons.
- **Soma (cell body)** — integrates the incoming signals; if the sum
  exceeds threshold, the neuron fires.
- **Axon** — carries the outgoing signal away from the soma. Often
  wrapped in **myelin** (a fatty insulator made by glial cells) that
  speeds conduction.
- **Axon terminal** — releases neurotransmitter into the **synapse**,
  the gap between this neuron and the next.

> A neuron is digital at the output (fires or doesn't) but analogue
> at the input (sums many graded signals to decide).

## The action potential

This is the cell-level mechanism behind everything that follows.

1. At rest, the inside of the neuron is more negative than the outside
   (resting potential ≈ −70 mV), maintained by the sodium-potassium
   pump.
2. Excitatory inputs depolarise the membrane (push it toward 0). When
   the membrane potential reaches threshold (≈ −55 mV), voltage-gated
   sodium channels snap open.
3. Sodium (Na⁺) rushes in. The membrane swings to ≈ +30 mV. This is
   the *spike*.
4. Sodium channels inactivate; voltage-gated potassium channels open.
   Potassium (K⁺) flows out, repolarising the membrane.
5. There's a brief **refractory period** during which the neuron can't
   fire again. This enforces the direction of propagation.

Two consequences worth remembering:

- The action potential is **all-or-none**. Stronger stimuli don't
  produce bigger spikes; they produce *more frequent* spikes.
- In myelinated axons, the spike jumps from one node of Ranvier to the
  next — **saltatory conduction**. This is why myelination matters
  clinically (multiple sclerosis attacks myelin).

## The synapse and neurotransmitters

When the action potential reaches the axon terminal, voltage-gated
calcium channels open. Calcium influx triggers vesicles of
neurotransmitter to fuse with the membrane and dump their contents
into the synaptic cleft. The neurotransmitter binds receptors on the
postsynaptic neuron and either depolarises it (excitatory) or
hyperpolarises it (inhibitory).

A handful of neurotransmitters carry most of the load on a PSYCH 1
exam:

| Neurotransmitter | Where it matters | Common context |
|---|---|---|
| Glutamate | Main excitatory transmitter in the CNS | Learning, memory (LTP) |
| GABA | Main inhibitory transmitter | Anxiolytics, alcohol target GABA receptors |
| Dopamine | Reward, motor control | Parkinson's (loss in substantia nigra), addiction |
| Serotonin | Mood, sleep, appetite | SSRIs (depression treatment) |
| Acetylcholine | Muscle activation, attention, memory | Alzheimer's pathology, neuromuscular junction |
| Norepinephrine | Arousal, alertness | Stress response, ADHD medications |
| Endorphins | Pain modulation | Opioid analgesics mimic them |

## Major brain regions

Bottom-up — older structures first, newer cortex last — is a useful
order because it parallels both evolutionary history and what you can
lose function-wise from damage at each level.

**Brainstem (medulla, pons, midbrain).** Heart rate, breathing,
arousal. Damage is often fatal because it controls the autonomic
fundamentals.

**Cerebellum.** "Little brain" tucked behind the brainstem. Coordinates
fine motor movement and balance; also contributes to certain forms of
implicit learning. Damage produces ataxia.

**Limbic system.** A loosely defined ring of structures in the middle
of the brain that handle emotion and memory:

- **Hippocampus** — required for forming new long-term *declarative*
  memories. Famous patient H.M. lost his hippocampi and could no
  longer form new explicit memories.
- **Amygdala** — fear processing, emotional salience. Lesion studies
  in animals and humans both show blunted threat detection.
- **Hypothalamus** — homeostasis (temperature, hunger, thirst), plus
  the master controller of the endocrine system via the pituitary.
- **Thalamus** — sensory relay station; nearly every sensory pathway
  except smell synapses here on its way to cortex.

**Cerebral cortex.** The wrinkled outer sheet, divided into four lobes:

- **Occipital** — primary visual cortex (V1) at the back.
- **Temporal** — auditory cortex, language comprehension (Wernicke's
  area, usually left), face recognition (fusiform face area).
- **Parietal** — somatosensory cortex, spatial attention, sensory
  integration.
- **Frontal** — motor cortex along the central sulcus, and prefrontal
  cortex anterior to that — planning, working memory, inhibition,
  personality. The Phineas Gage case is the classic demonstration of
  what you lose when you lose the prefrontal cortex.

## How we know any of this

Causal evidence in human neuroscience is hard. The four main methods,
each with different strengths:

- **Lesion studies** — natural lesions (stroke, tumour, surgery) tell
  us what a region is *necessary* for. Limited by the messiness of
  natural lesions.
- **Single-unit recording** — extracellular electrodes, mostly in
  animals, give the best temporal and spatial precision.
- **fMRI** — measures blood-oxygen-level-dependent (BOLD) signal as a
  proxy for neural activity. Excellent spatial resolution, slow
  temporal resolution. Correlational, not causal.
- **TMS / tDCS** — non-invasive stimulation that briefly disrupts or
  modulates a cortical region. Lets you ask causal questions
  ethically in healthy humans.

Bring the {{GUIDE:psych1-research-methods}} rubric to any neuroimaging
result you read about. "Region X lights up during task Y" is a
correlation, not a job description.

## Plasticity

The adult brain is not fixed. Synaptic strength changes with experience
(long-term potentiation), cortical maps reorganise after injury or
training, and adult neurogenesis occurs at least in the hippocampus.
This is the substrate for learning — picked up in detail in
{{GUIDE:psych1-learning-conditioning}} and {{GUIDE:psych1-memory}}.

For the canonical cell diagrams and regional cross-sections, see
{{FILE:stanford-psych1-neuroscience-primer}}. A higher-resolution scan
image (useful for visualising fMRI-style data) is in
{{FILE:unsplash-brain-scan-6}}.

## Practice

When the synapse → neurotransmitter → region chain is something you can
walk through unprompted, take {{QUIZ:psych1-neuroscience-primer-quiz}}.
