---
title: Troubleshooting Linux cpu usage
author: GaborZeller
date: 2024-09-29T21-49-41Z
tags:
draft: true
---

# Troubleshooting Linux cpu usage

## Check CPU profile

To see details of your CPU run the following:

```sh
lscpu
```

```sh
Architecture:            x86_64
  CPU op-mode(s):        32-bit, 64-bit
  Address sizes:         39 bits physical, 48 bits virtual
  Byte Order:            Little Endian
CPU(s):                  12
  On-line CPU(s) list:   0-11
...
```
## Check load over time

One useful trick is to see how the CPU usage has been changing over time.

```sh
uptime
```

```sh
22:53:20 up  8:30,  0 users,  load average: 0.00, 0.00, 0.00 # The last three numbers indicate the load average for the last 1, 5, & 15 minutes.
```
## Get a high level overview of CPU consumption

`vmstat` is a good way to figure out what is hogging up the CPU.

```sh
vmstat -w 1
```

```sh
--procs-- -----------------------memory---------------------- ---swap-- -----io---- -system-- --------cpu--------
   r    b         swpd         free         buff        cache   si   so    bi    bo   in   cs  us  sy  id  wa  st
   1    0            0      5058912       330996      2204436    0    0     4     5    5   46   0   0 100   0   0
   0    0            0      5058912       330996      2204436    0    0     0     0    7  326   0   0 100   0   0
   0    0            0      5056300       331004      2204436    0    0     0    36   31  395   0   0 100   0   0
   0    0            0      5056300       331004      2204436    0    0     0     0    5  318   0   0 100   0   0
   0    0            0      5056300       331004      2204436    0    0     0     0    7  310   0   0 100   0   0
   0    0            0      5056336       331004      2204436    0    0     0     0   15  344   0   0 100   0   0
```
- `r` - number of processes waiting for runtime. If this is constantly higher than your core count then you have too much backlog
- `us` - percentage of time spent on executing user space instructions. If this number is high it means an application is over utilising the CPU
- `sy` - this is the time spent on executing system instructions. Generally should be around 20%, if higher it might mean you have issues with the kernel or a driver
- `wa` - time spent on waiting for I/O operations. If constantly high check the disc
- `st` - this is the percentage of time stolen from CPU where a virtual machine is stealing CPU cycles from an other. If this is the case and possible you probably want to move the noisy service to an other physical host

## Do a per CPU analysis

Having an overall view is fine but you might need a breakdown per cores.

```sh
mpstat -P ALL 1
```

```sh
Linux 5.15.133.1-microsoft-standard-WSL2 (LAPTOP-PREDATOR-GABRIEL)      09/29/24        _x86_64_        (12 CPU)

23:13:33     CPU    %usr   %nice    %sys %iowait    %irq   %soft  %steal  %guest  %gnice   %idle
23:13:34     all    0.08    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00   99.92
23:13:34       0    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00  100.00
23:13:34       1    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00  100.00
23:13:34       2    0.99    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00   99.01
23:13:34       3    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00  100.00
23:13:34       4    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00  100.00
23:13:34       5    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00  100.00
23:13:34       6    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00  100.00
23:13:34       7    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00  100.00
23:13:34       8    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00  100.00
23:13:34       9    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00  100.00
23:13:34      10    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00  100.00
23:13:34      11    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00    0.00  100.00
```

The above view is useful if you want to be able to tell if there is a single CPU struggling with the load. That is usually a good indication of a single threaded process choking one of the cores.

## Identify the process that is choking your CPU

To get a read and copy friendly view of your processess run the following:

```sh
pidstat
pidstat -u 5 60 # get an update every 5 seconds for the next 60 seconds
```

```sh
Linux 5.15.133.1-microsoft-standard-WSL2 (LAPTOP-PREDATOR-GABRIEL)      09/29/24        _x86_64_        (12 CPU)

23:19:03      UID       PID    %usr %system  %guest   %wait    %CPU   CPU  Command
23:19:03        0        14    0.00    0.01    0.00    0.00    0.01     8  Relay(15)
23:19:03     1000        15    0.00    0.00    0.00    0.00    0.00     3  zsh
23:19:03     1000       130    0.03    0.03    0.00    0.00    0.06     4  tmux: server
23:19:03     1000       429    0.00    0.00    0.00    0.00    0.00     1  zsh
23:19:03     1000       753    0.00    0.00    0.00    0.00    0.00     3  pnotes
23:19:03     1000       763    0.00    0.00    0.00    0.00    0.00     3  nvim
23:19:03     1000       765    0.03    0.00    0.00    0.00    0.03     5  nvim
23:19:03     1000      6902    0.01    0.00    0.00    0.00    0.02     2  zsh
23:19:03     1000     12674    0.02    0.01    0.00    0.00    0.03     4  zsh
23:19:03     1000     28054    0.00    0.00    0.00    0.00    0.00     3  zsh
23:19:03     1000     31561    0.00    0.00    0.00    0.00    0.00    10  nvim
23:19:03     1000     31563    0.01    0.00    0.00    0.00    0.01     9  nvim
```

- `%usr` - user space allocation. A high number is an indicator that a process needs scaling
- `%system` - system/kernel space time allocation. A high number indicates driver/kernel issues
- `%wait` - indicates I/O disc issues

