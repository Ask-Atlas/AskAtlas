---
slug: psych1-sensation-and-perception
course:
  ipeds_id: "243744"
  department: "PSYCH"
  number: "1"
title: "Sensation and Perception — PSYCH 1"
description: "How transduced energy becomes seen, heard, and felt experience — from the retina to the bistable Necker cube."
tags: ["perception", "sensation", "vision", "psychophysics", "midterm"]
author_role: bot
quiz_slug: psych1-sensation-and-perception-quiz
attached_files:
  - stanford-psych1-perception-illusions
  - unsplash-png-brain-6
attached_resources: []
---

# Sensation and Perception

> **Sensation** is the conversion of physical energy (light, sound,
> pressure, chemical concentration) into neural signal.
> **Perception** is what your brain *makes* of those signals — and the
> two come apart more often than intuition suggests.

This split is the organising idea for the unit. Sensation is bottom-up
and largely deterministic: photons hit a photoreceptor, ion channels
open, an action potential propagates. Perception is top-down,
probabilistic, and inference-rich: your visual system is constantly
hypothesising about what's out there, given the noisy and ambiguous
data the eyes deliver. Most of the famous illusions in this section
exist because the inference machinery has been pushed off its usual
operating point.

## Psychophysics

Psychophysics is the quantitative study of the relation between
physical stimulus magnitude and reported sensation magnitude.

- **Absolute threshold** — the smallest stimulus you can detect 50% of
  the time. (50% is convention; detection is graded, not all-or-none.)
- **Difference threshold (just-noticeable difference, JND)** — the
  smallest change between two stimuli you can reliably tell apart.
- **Weber's law** — the JND is a roughly constant *proportion* of the
  baseline stimulus. Adding 1 lb to a 5-lb dumbbell is noticeable;
  adding 1 lb to a 50-lb dumbbell is not.
- **Signal detection theory** — separates *sensitivity* (how well your
  system actually distinguishes signal from noise, captured by `d′`)
  from *response bias* (how willing you are to say "yes, I detected
  it"). Two observers with the same `d′` can give very different hit
  rates depending on their bias.

Signal detection is the cleanest demonstration in PSYCH 1 that
"perception" is not a passive readout. Your willingness to report a
faint sound depends on the cost of false alarms, your prior, and your
mood — none of which change the physical signal.

## Vision: from photon to percept

Light enters the eye through the cornea and pupil, focuses on the
**retina**, and is transduced by two kinds of photoreceptor:

- **Rods** — about 120 million; one type, sensitive in dim light, no
  colour information, concentrated in the periphery.
- **Cones** — about 6 million; three types tuned to short, medium, and
  long wavelengths, give us colour, concentrated in the **fovea**
  (the high-acuity centre of the retina).

Photoreceptors feed into bipolar cells, which feed into ganglion cells.
The ganglion-cell axons leave the eye as the optic nerve. The point
where the nerve exits has no photoreceptors — that's your **blind
spot**. You don't notice it because the brain interpolates across it.

Beyond the retina, the pathway crosses at the optic chiasm
(information from the right visual field of *both eyes* ends up in
the left hemisphere), synapses in the **lateral geniculate nucleus**
of the thalamus, and projects to **primary visual cortex (V1)** in the
occipital lobe. From V1 the signal splits into:

- **Dorsal stream ("where/how")** — toward parietal cortex; spatial
  location, motion, action guidance.
- **Ventral stream ("what")** — toward temporal cortex; object
  recognition, faces (fusiform face area), reading (visual word form
  area).

Patients with **prosopagnosia** (face blindness) usually have ventral-
stream damage; patients with **akinetopsia** (motion blindness) have
dorsal-stream damage. The double dissociation is strong evidence that
these streams really are functionally distinct.

## Colour

Two complementary theories together explain colour vision:

- **Trichromatic theory** (Young-Helmholtz) — three cone types whose
  relative activations encode hue. Explains why most people are
  trichromatic and why colour-blindness usually traces to a missing or
  mutated cone opsin.
- **Opponent-process theory** (Hering) — downstream of the cones, the
  visual system codes colour in opposing pairs: red/green, blue/yellow,
  black/white. Explains negative afterimages: stare at red, look at
  white, see green.

These aren't rival theories — they describe different processing
stages.

## Hearing

Sound is pressure waves. The **pinna** funnels them into the ear
canal; the **eardrum** vibrates; the **ossicles** (malleus, incus,
stapes) amplify and transmit the vibration to the **oval window** of
the **cochlea**. Inside the cochlea, the basilar membrane acts as a
mechanical frequency analyser: high frequencies displace it near the
base, low frequencies near the apex. Hair cells along the membrane
transduce that displacement into neural signal that travels via the
auditory nerve to brainstem nuclei, the medial geniculate, and
auditory cortex in the temporal lobe.

Two coding schemes operate together:

- **Place coding** — *which* hair cells fire encodes pitch (works best
  for high frequencies).
- **Temporal coding** — *when* they fire encodes pitch via the timing
  of action potentials (works best for low frequencies, up to a few
  hundred Hz).

## The other senses, briefly

- **Touch / somatosensation** — multiple receptor types (pressure,
  temperature, pain) project to the somatosensory cortex along a
  topographic *homunculus*; body parts with high acuity (lips, hands)
  get disproportionately large cortical real estate.
- **Smell (olfaction)** — chemoreceptors in the nasal epithelium;
  unique among senses in that signals reach cortex *without* synapsing
  in the thalamus first.
- **Taste (gustation)** — five basic qualities: sweet, salty, sour,
  bitter, umami.
- **Proprioception** — sense of body position, mediated by receptors
  in muscles and joints. Not one of the "five senses" of childhood,
  but you'd notice immediately if it went away.

## Perceptual organisation

Once the signal arrives in cortex, perception assembles it into
coherent objects. The **Gestalt principles** describe regularities the
visual system uses for that grouping:

- **Proximity** — nearby elements group together.
- **Similarity** — elements that look alike group together.
- **Continuity** — we perceive smooth, continuous lines rather than
  abrupt direction changes.
- **Closure** — we fill in gaps to perceive complete shapes.
- **Common fate** — elements moving together group together.

These are not arbitrary preferences. They reflect statistical
regularities of the natural world: objects tend to be made of similar,
contiguous, smoothly-bounded surfaces. The grouping rules are roughly
the right *prior* given that statistical structure.

## Top-down inference and illusions

> Perception is the brain's best guess about the cause of the input,
> given the input plus prior knowledge.

Most illusions exploit this. The **Müller-Lyer illusion** (lines with
inward vs. outward arrowheads appear different lengths) seems to
recruit machinery that interprets the arrowheads as depth cues. The
**Necker cube** is bistable because two 3-D interpretations are
equally consistent with the same 2-D input. The **moon illusion** —
the moon looks bigger near the horizon — has multiple proposed
explanations all of which involve the brain combining the retinal
image with apparent-distance cues.

A worked walkthrough of half a dozen classic illusions is in
{{FILE:stanford-psych1-perception-illusions}}; for a higher-resolution
brain image to anchor the cortical pathway, see
{{FILE:unsplash-png-brain-6}}.

## Attention is part of perception

You don't perceive everything that hits your retinas. **Inattentional
blindness** demonstrations (the gorilla walking through the basketball
game) and **change blindness** demonstrations (pairs of images
alternating with a brief mask) show that conscious perception requires
attention, not just sensation. Practical takeaway: divided attention
genuinely costs you perception, not just speed.

## Where this comes back

The sensory systems are the input side of every cognitive process in
the rest of the course — encoding into memory
({{GUIDE:psych1-memory}}), classical conditioning where the CS is a
sensory stimulus ({{GUIDE:psych1-learning-conditioning}}), and the
perceptual roots of social judgement
({{GUIDE:psych1-social-psychology}}).

When you can explain the rod/cone split, the dorsal/ventral split, and
why the Necker cube is bistable from your own notes, take
{{QUIZ:psych1-sensation-and-perception-quiz}}.
