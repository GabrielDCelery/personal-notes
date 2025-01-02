---
title: Comptia A Plus CPU
author: GaborZeller
date: 2025-01-02T20-02-44Z
tags:
draft: true
---

# Comptia A Plus CPU

## CPU structure

The most important parts of the CPU:

- EDB (external data bus)
	- transfers data in and out of a CPU
	- come in 8, 16, 32, 64 bits
- Register
	- stores data in the CPU
	- a CPU is made up of multiple registers and registers vary in size (can be 8, 16, 32 etc... bits)
	- there are general purpose registers like AX, BX, CX etc...
	- you can have special registers for special tasks (e.g. register for floating point calculations)
- External Data Bus Codebook
	- contains instructions on what actions should the register do (e.g. 10000000 - the next line is a number put it on the AX register, followed by 00000010 which is the number 2 going on the register)
- Clock
	- the clock wire is a special wire that tells the CPU to act on that an other piece of information needs to be processed
	- CPU usually requires two or more clock cycles to act on a command
	- the maximum number of clock cycles that the CPU can handle in a given period is the `clock speed` (e.g. 4.77 MHz = 4.77 million cycles per second)

## CPU and motherboard

- System crystal
	- its a quartz oscillator on the motherboard that acts as a conductor for the devices connected to the computer
	- the signal it sends gets adjusted by a `clock chip` so it can speed up the signal to accomodate `multiple CPUs`

## CPU and RAM

The RAM is like a `spreadsheet` that stores data that the CPU actively uses. Each line in that spreadsheet is a `byte`.

Since the CPU needs to be able to access different parts of that spreadsheet we need an extra chip called the `MCC (memory controller chip)` which retrieves data from the memory for the EDB and puts back data from the EDB to memory. The MCC and the EDB are connected via the `address bus`.

Each wire in the address bus represents a bit, so for example a 20 line address bus can handle 2^20 (1 million) combinations of addresses, so it can handle 1 Megabyte of memory.

## CPU manufacturing

To manufacture a CPU yuou need three things.

1. An `instruction set` which determines how the CPU will interact with other components (this is the `ISA` - industry standard architecture)
2. A `floorplan` which is the physical layout of the chip
3. A `fabrication company` that assembles the thing

`Intel` does all three steps and `AMD` only does the first two and outsources the third to other companies. `ARM` does the first step and licenses their designs to companies like Apple or Samsung, who design the floorplan which get further outsourced to other companies.

[!TIP] The CompTIA exam spells out the acronyms, so ARM is going to be Advanced RISC Machine (ARM)

[!TIP] CompTIA reverses the numbers for x86-64 so it will be shown as x64/x86

## CPU models






