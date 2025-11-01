---
title: "Linux std redirects"
date: 2025-11-01
tags: ["linux", "std"]
---

# How std redirect works

This example redirects both stderr and stdin to output.txt.

```sh
node app.js 1> output.txt 2>&1
```

## Part 1

`1> output.txt`

- `1` = file descriptor for stdout
- `>` = redirect operator

Meaning: Redirect stdout (fd 1) to output.txt

## Part 2

`2>&1`

- `2` = file descriptor for stderr
- `>&` = redirect to another file descriptor
- `1` = stdout's file descriptor

Meaning: Redirect stderr (fd 2) to wherever stdout (fd 1) is going

# Practical Examples

```sh
# Capture everything (stdout + stderr)
node app.js > output.txt 2>&1

# Discard everything
node app.js > /dev/null 2>&1

# Separate files for stdout and stderr
node app.js 1> output.txt 2> errors.txt

# Only capture errors
node app.js 2> errors.txt

# Discard errors, keep output
node app.js 2> /dev/null
```

# Modern syntax

```sh
# These are equivalent
node app.js > output.txt 2>&1
node app.js &> output.txt # Shorthand - both to same file

# So are these
node app.js > /dev/null 2>&1
node app.js &> /dev/null # Shorthand - discard both
```

# why is this wrong why does the order matter

```sh
# CORRECT
node app.js 1> output.txt 2>&1

# WRONG
node app.js 2>&1 1> output.txt
```

Why `2>&1 1> output.txt` is wrong?

Step-by-step execution:

1. Initial state:

```txt
stdout (1) → terminal
stderr (2) → terminal
```

2. After `2>&1` (first operation):

```txt
stdout (1) → terminal
stderr (2) → terminal (copied from where stdout currently points)
```

At this moment, stderr makes a copy of where stdout is pointing (the terminal).

3. After `1> output.txt` (second operation):

```txt
stdout (1) → output.txt (changed!)
stderr (2) → terminal (still pointing to old location!)
```

Now stdout changes to the file, but stderr is already "locked in" to the terminal.

The Key Concept: It's a Copy, Not a Live Link

When you do 2>&1, you're not saying `stderr always follows stdout`. You're saying `right now, copy stdout's current destination to stderr`.
