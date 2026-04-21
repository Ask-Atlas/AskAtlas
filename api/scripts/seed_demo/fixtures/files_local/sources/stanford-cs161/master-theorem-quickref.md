---
slug: stanford-cs161-master-theorem-quickref
title: "Master Theorem Quick Reference"
mime: text/plain
filename: master-theorem-quickref.txt
course: stanford/cs161
description: "Plain-text quick reference for the master theorem with the three cases and worked examples."
author_role: bot
---

MASTER THEOREM QUICK REFERENCE
==============================

Stanford CS 161 - Design and Analysis of Algorithms

The master theorem solves recurrences of the form:

    T(n) = a * T(n/b) + f(n)

where:
  a >= 1  is the number of subproblems
  b  > 1  is the factor by which input size shrinks
  f(n)    is the work done outside the recursive calls

Define the "watershed" function:

    n^(log_b a)

Compare f(n) against this watershed. Three cases follow.


CASE 1: WORK DOMINATED BY LEAVES
--------------------------------

If f(n) = O(n^(log_b a - epsilon)) for some epsilon > 0,

then T(n) = Theta(n^(log_b a)).

Intuition: The recursion tree has more work near the leaves than at
the root. The total is dominated by the number of leaves.

Example: T(n) = 8 T(n/2) + n^2
  log_b a = log_2 8 = 3
  f(n) = n^2 grows slower than n^3
  => T(n) = Theta(n^3)


CASE 2: BALANCED WORK
---------------------

If f(n) = Theta(n^(log_b a) * log^k n) for some k >= 0,

then T(n) = Theta(n^(log_b a) * log^(k+1) n).

Intuition: Each level of the recursion tree does the same amount of
work, up to a polylog factor. Total work = work-per-level * depth.

Example: T(n) = 2 T(n/2) + n
  log_b a = log_2 2 = 1
  f(n) = n = Theta(n^1)
  => T(n) = Theta(n log n)   (this is merge sort)

Example: T(n) = 2 T(n/2) + n log n
  f(n) = Theta(n^1 * log n), k = 1
  => T(n) = Theta(n log^2 n)


CASE 3: WORK DOMINATED BY ROOT
------------------------------

If f(n) = Omega(n^(log_b a + epsilon)) for some epsilon > 0,
AND f(n) satisfies the regularity condition
    a * f(n/b) <= c * f(n) for some c < 1, for large n,

then T(n) = Theta(f(n)).

Intuition: Work at the root dominates. The recursive calls together
do less work than one level up.

Example: T(n) = 2 T(n/2) + n^2
  log_b a = 1
  f(n) = n^2 grows faster than n^1
  Regularity: 2 * (n/2)^2 = n^2/2 <= (1/2) * n^2. OK.
  => T(n) = Theta(n^2)


WHAT THE MASTER THEOREM DOES NOT COVER
--------------------------------------

The master theorem fails when:

  1. The gap between f(n) and n^(log_b a) is not polynomial.
     Example: T(n) = 2 T(n/2) + n / log n.

  2. Subproblem sizes are uneven.
     Example: T(n) = T(n/3) + T(2n/3) + n.

  3. The recursion is not of the form a T(n/b) + f(n).
     Example: T(n) = T(n-1) + n  (subtract, not divide).

For these, use the recursion tree method or substitution method.


CHEAT TABLE OF COMMON RECURRENCES
---------------------------------

Recurrence                             Solution
----------                             --------
T(n) = T(n/2) + 1                      Theta(log n)       (binary search)
T(n) = T(n/2) + n                      Theta(n)
T(n) = 2 T(n/2) + 1                    Theta(n)           (traversal)
T(n) = 2 T(n/2) + n                    Theta(n log n)     (merge sort)
T(n) = 2 T(n/2) + n^2                  Theta(n^2)
T(n) = 3 T(n/2) + n                    Theta(n^(log_2 3))
T(n) = 3 T(n/2) + n^2                  Theta(n^2)
T(n) = 4 T(n/2) + n                    Theta(n^2)
T(n) = 4 T(n/2) + n^2                  Theta(n^2 log n)
T(n) = 7 T(n/2) + n^2                  Theta(n^(log_2 7)) (Strassen)
T(n) = 3 T(n/2) + n                    Theta(n^(log_2 3)) (Karatsuba)
T(n) = T(n-1) + 1                      Theta(n)
T(n) = T(n-1) + n                      Theta(n^2)
T(n) = 2 T(n-1) + 1                    Theta(2^n)


DECISION CHECKLIST
------------------

Given T(n) = a T(n/b) + f(n):

  [ ] Compute log_b a.
  [ ] Compare f(n) polynomially against n^(log_b a).
  [ ] If f is polynomially smaller -> Case 1.
  [ ] If f is within a log factor  -> Case 2.
  [ ] If f is polynomially larger and regular -> Case 3.
  [ ] If none of the above apply   -> use recursion tree.
