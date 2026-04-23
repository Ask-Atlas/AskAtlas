---
slug: stanford-math51-notation-glossary
title: "MATH 51 Notation Glossary"
mime: text/plain
filename: notation-glossary.txt
course: stanford/math51
description: "Plain-text glossary of the notation used throughout Stanford MATH 51 for vectors, matrices, and calculus."
author_role: bot
---

MATH 51 NOTATION GLOSSARY
=========================

This glossary lists the notation used throughout MATH 51 (Linear Algebra,
Multivariable Calculus, and Modern Applications). Symbols are grouped by
topic. Entries are short on purpose; consult the textbook for full
definitions.

SETS AND NUMBERS
----------------

R             The real numbers.
R^n           n-tuples of real numbers; the standard n-dimensional space.
C             The complex numbers.
Z             The integers.
{x : P(x)}    The set of all x satisfying property P.

VECTORS
-------

v, w          Vectors, typically in R^n. Written as columns by default.
v_i           The i-th component of vector v.
e_i           The i-th standard basis vector (1 in slot i, 0 elsewhere).
0             The zero vector (context-dependent dimension).
v + w         Componentwise addition.
c * v         Scalar multiplication.
||v||         Euclidean norm: sqrt(v_1^2 + ... + v_n^2).
v . w         Dot product: v_1 w_1 + ... + v_n w_n.
v x w         Cross product (R^3 only).
v^T           Transpose (turns a column into a row).
proj_w v      Orthogonal projection of v onto w.

MATRICES
--------

A, B          Matrices. Usually capital Latin letters.
A_{ij}        Entry in row i, column j of A.
A^T           Transpose: (A^T)_{ij} = A_{ji}.
A^{-1}        Inverse of a square matrix A (when it exists).
I, I_n        Identity matrix (size n x n).
det(A)        Determinant of A.
tr(A)         Trace: sum of diagonal entries.
rank(A)       Rank: dimension of the column space.
ker(A)        Kernel / null space: {x : Ax = 0}.
col(A)        Column space: span of the columns.
row(A)        Row space.
A^k           A multiplied by itself k times.

LINEAR SYSTEMS AND SPACES
-------------------------

Ax = b        Standard linear system.
RREF          Reduced row echelon form.
span{v_1,...} The set of all linear combinations of the given vectors.
dim V         Dimension of vector space V.
V + W         Sum of subspaces.
V (+) W       Direct sum (intersection is trivial).
V perp        Orthogonal complement.

EIGENTHEORY
-----------

lambda        Eigenvalue (scalar, possibly complex).
v             Eigenvector (nonzero) satisfying A v = lambda v.
E_lambda      Eigenspace associated with eigenvalue lambda.
p(lambda)     Characteristic polynomial: det(A - lambda I).
P D P^{-1}    Diagonalization: P holds eigenvectors, D holds eigenvalues.
Q Lambda Q^T  Spectral decomposition for symmetric A (Q orthogonal).

CALCULUS OF SCALAR FIELDS
-------------------------

f : R^n -> R  A scalar field; input is a point in R^n.
d_{x_i} f     Partial derivative with respect to x_i
              (also written as partial f / partial x_i).
grad f        Gradient: [d_x f, d_y f, d_z f] (and similar in n dims).
D_u f         Directional derivative in the unit direction u.
H f           Hessian matrix: the matrix of second partials.
Delta         Laplacian: sum of pure second partials.

VECTOR-VALUED FUNCTIONS
-----------------------

r(t)          Vector-valued function of one variable (a parametric curve).
r'(t)         Velocity vector: componentwise derivative.
||r'(t)||     Speed.
Dx            Jacobian matrix of a multivariable map x.

OPTIMIZATION
------------

f_max, f_min  Maximum and minimum values of f (possibly constrained).
grad f = 0    Unconstrained first-order condition.
lambda        Lagrange multiplier (constrained optimization).
H f at x_0    Local classification via the Hessian:
              positive definite  -> local min
              negative definite  -> local max
              indefinite         -> saddle point

COMMON SHORTHANDS
-----------------

iff           If and only if.
s.t.          Subject to / such that.
w.r.t.        With respect to.
LHS, RHS      Left-hand side, right-hand side of an equation.
WLOG          Without loss of generality.

Keep this file handy during problem sets. When notation in the book
differs slightly, rely on the instructor's conventions from lecture.
