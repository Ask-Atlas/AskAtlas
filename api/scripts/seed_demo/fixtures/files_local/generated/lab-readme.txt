CPTS 121 Lab Environment Setup

This guide walks you through compiling and running your first C program for CPTS 121. Read it once front to back before touching a terminal. Every command here assumes you are in the directory that contains your .c file.

1.  Get a Compiler

On the lab machines, gcc is already installed. From a terminal, run this to confirm:

    gcc --version

If you see a version number (for example, gcc 11.x or newer), you are ready. If the command is not found, stop and ask your TA. Do not install your own compiler on a lab machine.

On your personal laptop:

    - macOS: run "xcode-select --install" to install the command-line tools, which include clang. clang accepts the same flags as gcc for everything we do in this course.
    - Linux: install the build-essential package using your distro's package manager. For Ubuntu that is "sudo apt install build-essential".
    - Windows: install MSYS2 and use the UCRT64 shell, or install WSL2 and work inside Ubuntu. Do not use MinGW builds that come bundled with random IDEs — they are often outdated.

2.  Pick an Editor

You need an editor that shows line numbers, highlights C syntax, and uses spaces instead of tabs. Any of these work:

    - VS Code with the C/C++ extension
    - Vim or Neovim with syntax on
    - Nano for very small edits

Do not edit .c files in a word processor. Smart quotes will destroy your code.

3.  Project Layout

Keep one directory per assignment. A typical layout looks like this:

    lab03/
      main.c
      helpers.c
      helpers.h
      README.txt

Source files end in .c, header files in .h. Do not commit compiled binaries to your submission.

4.  Compile

The standard compile command for this course is:

    gcc -Wall -Wextra -std=c17 -g -o program main.c

Breakdown:
-Wall -Wextra turn on warnings. Warnings are real problems. Fix them.
-std=c17 use the C17 standard.
-g include debug symbols so gdb can show you line numbers.
-o program name the output file “program” instead of a.out.

If you have multiple .c files, list them all:

    gcc -Wall -Wextra -std=c17 -g -o program main.c helpers.c

5.  Run

On macOS and Linux:

    ./program

On Windows (WSL or MSYS2):

    ./program.exe

If the program reads input, either type it at the prompt or pipe a file:

    ./program < input.txt

6.  Common First-Time Errors

“command not found: gcc”
The compiler is not installed or not on your PATH. Re-check step 1.

“undefined reference to ‘function_name’”
You forgot to list all your .c files on the compile line, or you declared a function in a header but never wrote the body.

“Segmentation fault”
Your program dereferenced a bad pointer, wrote past an array boundary, or used an uninitialized variable. Compile with -g and run under a debugger.

“implicit declaration of function”
You called a standard library function without including its header. stdio.h for printf, stdlib.h for malloc, string.h for strlen.

7.  Submitting

Before you hand anything in:

    - Your code compiles with zero warnings using the flags above.
    - You tested at least three different inputs, including edge cases.
    - Your files are named exactly as the assignment specifies.
    - Your name and WSU ID are in a comment block at the top of every file.

8.  Getting Help

Office hours are the fastest path to answers. Post in the class forum for non-urgent questions. Include:

    - What you tried.
    - What you expected to happen.
    - What actually happened, including the exact error message.

Good luck this semester.
