---
title: Enable dual booting on my Thinkpad
author: GaborZeller
date: 2026-02-19
tags: dual boot, thinkpad
---

# The problem

Got a Lenovo Thinkpad T14 Gen 1, 14" Core I7-10610U 1.8GHz, 32GB RAM, 512GB M.2 with Windows 10 and wanted to endable dual boot on it.

# Steps

## Checked if the machine needs backing up or if the licensing was intact

```sh
# command prompt
slmgr /dli
# this opened a window with license info
```

## Shrank the Windows partition to make space for Linux

- `Win + X` -> Disk management (Create and format hard disk partitions)
  - 260 MB EFI system partition
  - ~480GB NTFS
  - 1.95GB Recovery Partition
- Right click C drive -> Shrink volume
  - decided to shrank by `358400 MB` which left `~160GB` on Windows and left `~350GB` for Linux
- Verified that I saw ~350GB `Unallocated` black bar

## Disable Fast Startup

Fast startup hybernates windows instead of fully shutting down, this can locak the NTFS file system and cause corruption when Linux tries to access it.

Control Panel -> Power Options -> Choose what the power buttons do (left menu) -> Change settings that are currently unavailable (admin perm) -> Uncheck `Turn on fast startup (recommended)` -> Hybernate should also be unchecked (probably is by default)-> Save changes

## Disabled Bitlocker

Bitlocked had to be disabled otherwise the Ubuntu installer can't access the UEFI partition.

## Configure BIOS

Shut down machine proper (not reboot) -> press F1 -> BIOS

a. Secure boot

I was using Ventoy for installing Ubuntu so had to disable secure boot because without it I was getting a `0x1A security violation error`.
Verify is secure boot is on - I decided to use Ubuntu so it was fine because they use a Microsoft-signed bootloader. So I got better security for free.

Security -> Secure Boot (toggle)

b. Boot mode

Startup -> UEFI/Legacy Boot -> Should be on `UEFI Only`

c. USB Boot

Startup -> Boot device list F12 Option should be ON

## Create Bootable USB

Tried using `Ventoy` first but the installation got stuck completely so went with `Rufus`.

- Open Rufus
- Device: Select your USB drive
- Boot selection: Click SELECT, then choose your Ubuntu ISO file
- Partition scheme: GPT (for UEFI) or MBR (for legacy BIOS)
- Target system: Matches your partition scheme
- File system: Leave as default (usually FAT32)
- Volume label: Whatever you want (e.g., "Ubuntu")

## Install Ubuntu

Pressed F12 when the Lenovo logo was shown and selected by USB. Went with the `normal mode` (grub2 is only needed for legacy compatiblity). Selected my Ubuntu 24.04 and chose "Try or Install Ubuntu".

- Normal installation
- Default selection
- Install Ubuntu alonside Windows Boot
