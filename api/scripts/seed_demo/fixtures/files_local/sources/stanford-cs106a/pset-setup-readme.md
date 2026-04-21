---
slug: stanford-cs106a-pset-setup-readme
title: "Problem Set Environment Setup"
mime: text/plain
filename: pset-setup-readme.txt
course: stanford/cs106a
description: "Plain-text setup guide for the CS 106A Python pset environment: install, run, submit, and troubleshoot."
author_role: bot
---

CS 106A - Problem Set Environment Setup
========================================

This file walks you through installing Python, setting up your editor,
running starter code, and submitting a problem set. Read it top to
bottom the first time. Keep it open for the rest of the quarter.

1. Install Python 3.11 or newer
-------------------------------

macOS:
  Open Terminal, then run:
    brew install python@3.11

Windows:
  Download the installer from python.org. During install, CHECK the box
  that says "Add Python to PATH". This matters.

Linux:
  Your package manager probably has it:
    sudo apt install python3 python3-venv

Verify the install:
    python3 --version

You should see something like: Python 3.11.7

2. Install VS Code
------------------

Download from code.visualstudio.com. After install, open Extensions
(Cmd+Shift+X on macOS, Ctrl+Shift+X elsewhere) and add:

  - Python (Microsoft)
  - Pylance (Microsoft)

Both come from Microsoft and are free.

3. Download the starter code
----------------------------

Each pset is distributed as a zip file on Canvas. Unzip it somewhere
sensible, like:

  ~/cs106a/assignment3/

Open that folder in VS Code:
    File -> Open Folder... -> pick the unzipped directory

4. Create a virtual environment
-------------------------------

A virtual environment keeps this pset's dependencies isolated from the
rest of your machine.

From the Terminal, cd into the pset folder and run:

    python3 -m venv .venv

Then activate it:

  macOS / Linux:
    source .venv/bin/activate

  Windows PowerShell:
    .venv\Scripts\Activate.ps1

Your prompt should now start with (.venv). Install required packages:

    pip install -r requirements.txt

If the pset ships without a requirements file, you usually only need:

    pip install pytest

5. Run the starter code
-----------------------

Every assignment has one main entry point, usually named after the
problem. For example:

    python hangman.py

To run the tests the course staff provide:

    pytest

pytest will find files named test_*.py and run them. A green dot means
one test passed; F means one failed; E means the test itself crashed.

6. Edit, save, run, repeat
--------------------------

The loop is:
  a. Read the problem statement.
  b. Write ONE small piece.
  c. Run it. Read the output.
  d. Fix the next smallest broken thing.

If you catch yourself writing for more than ten minutes without running
the program, stop and run it.

7. Submit
---------

- Double-check your file is named exactly as the handout says.
- Remove all print() statements you added for debugging.
- Zip the folder (or follow the specific submission format in the spec).
- Upload to Gradescope before the deadline.
- Confirm the submission page shows a green check.

8. Troubleshooting
------------------

"command not found: python3"
  -> The PATH entry didn't stick. Close and reopen the terminal. If that
     fails, reinstall Python with the "Add to PATH" box checked.

"ModuleNotFoundError: No module named X"
  -> Your virtual environment isn't activated, or you forgot
     pip install. Check the prompt for (.venv).

"IndentationError: unexpected indent"
  -> Mixed tabs and spaces. In VS Code, Cmd+Shift+P -> "Convert
     Indentation to Spaces".

"My program hangs."
  -> You probably wrote an infinite loop. Press Ctrl+C in the terminal
     to kill it, then re-check your loop condition.

9. Getting help
---------------

- Office hours: see the course website for the weekly schedule.
- Ed discussion board: search before posting. Someone else hit the
  same error ten minutes ago.
- LaIR (Lab for Assistance, Interaction, and Reinforcement): evenings
  in the CS building, staffed by section leaders.

Do not share code with classmates. Sharing *approaches* is fine;
sharing *solutions* is an honor code violation. When in doubt, ask.

That's the whole loop: install, activate, code, test, submit. Welcome
to CS 106A.
