---
slug: wsu-cpts260-lab-setup-spim
title: "SPIM / MARS Lab Setup Guide"
mime: text/plain
filename: lab-setup-spim.txt
course: wsu/cpts260
description: "Step-by-step instructions for installing the SPIM and MARS MIPS simulators on Windows, macOS, and Linux."
author_role: bot
---

CPTS 260 LAB SETUP - SPIM AND MARS
===================================

This course uses two MIPS simulators. You only need one working installation
for assignments, but having both makes debugging easier.

1. WHICH SIMULATOR?

   - SPIM: original UC Berkeley simulator, C source, QtSpim GUI.
   - MARS: Java-based, friendlier UI, richer debugger, used by most
     textbooks. Recommended for this course.

2. INSTALL MARS (RECOMMENDED)

   Requirement: Java 8 or newer.

   - Download Mars4_5.jar from https://courses.missouristate.edu/KenVollmar/MARS/
   - Place it anywhere, e.g. ~/tools/mars/
   - Launch:
         java -jar Mars4_5.jar

   On macOS you may need to right-click and choose Open the first time
   because the jar is unsigned.

3. INSTALL QTSPIM

   Linux (Debian/Ubuntu):
         sudo apt install spim

   macOS (Homebrew):
         brew install spim
         brew install --cask qtspim     # GUI variant

   Windows: download the installer from
   https://spimsimulator.sourceforge.net/ and run it.

4. VERIFY THE INSTALL

   Create a file hello.asm with this content:

         .data
         msg: .asciiz "Hello from MIPS\n"
         .text
         .globl main
         main:
             li   $v0, 4
             la   $a0, msg
             syscall
             li   $v0, 10
             syscall

   In MARS: File > Open, then Run > Assemble (F3), Run > Go (F5).
   In SPIM: spim -file hello.asm

   You should see "Hello from MIPS" in the console pane.

5. RECOMMENDED SETTINGS

   Under Settings in MARS, enable:
   - Assemble all files in directory
   - Initialize program counter to global 'main'
   - Permit extended (pseudo) instructions

   Leave delayed branching OFF unless an assignment calls for it.

6. SUBMITTING WORK

   Submissions are the .asm source plus a short screenshot of the
   registers pane at halt. Do NOT zip an entire MARS install; only the
   source files are required.

7. COMMON PROBLEMS

   - "java: command not found" on macOS: install via
     `brew install --cask temurin`.
   - MARS hangs on large loops: Run > Pause, check the instruction
     counter, look for an unintentional infinite loop.
   - SPIM exits with "Cannot open input file": you are in the wrong
     working directory. Use an absolute path or `cd` first.

8. OFFICE HOURS

   Bring your screen and the exact .asm file that is failing. Half of
   reported "SPIM bugs" are missing .globl main directives.
