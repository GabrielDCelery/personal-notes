---
title: Troubleshooting Linux memory usage
author: GaborZeller
date: 2024-09-30T20-19-33Z
tags:
draft: true
---

# Troubleshooting Linux memory usage

## Get a high level overview

```sh
free -h
```

```sh
               total        used        free      shared  buff/cache   available
Mem:           7.7Gi       446Mi       3.8Gi       2.0Mi       3.4Gi       6.9Gi
Swap:          2.0Gi          0B       2.0Gi
```
- `used` - the amount of memory used, just because it is maxed out is not necessarily an issue since the operating system uses memory for caching
- `free` - free is the amount of free memory, if it is close to zero it is usually not an issue because operaing system cache things in memory for efficiency
- `available` - if this is close to zero it indicates an issue
- `swap` - if this is fluctuating the memory needs to use the space allocated on disc to avoid programs crashing

## Check if swap is being used

```sh
vmstat -w 1
```

```sh
--procs-- -----------------------memory---------------------- ---swap-- -----io---- -system-- --------cpu--------
   r    b         swpd         free         buff        cache   si   so    bi    bo   in   cs  us  sy  id  wa  st
   1    0            0      3967924       451212      3162776    0    0    17    18    8   45   0   0 100   0   0
   0    0            0      3967924       451212      3162816    0    0     0     4   12  342   0   0 100   0   0
   0    0            0      3967924       451212      3162816    0    0     0     0   19  351   0   0 100   0   0
   0    0            0      3967924       451212      3162816    0    0     0     0    6  319   0   0 100   0   0
   0    0            0      3969976       451212      3162816    0    0     0     0   18  363   0   0 100   0   0
   0    0            0      3969976       451212      3162816    0    0     0     0    6  324   0   0 100   0   0
```

Check the `si` and `so` columns. They should stay at 0.

## Identify processes that use too much memory

```sh
pidstat -r
pidstat -r 5 60 # get an update every 5 seconds for the next 60 seconds
```

```sh
Linux 5.15.133.1-microsoft-standard-WSL2 (LAPTOP-PREDATOR-GABRIEL)      09/30/24        _x86_64_        (12 CPU)

21:37:50      UID       PID  minflt/s  majflt/s     VSZ     RSS   %MEM  Command
21:37:50        0         1      0.01      0.00    2460    1608   0.02  init(Ubuntu)
21:37:50        0         5      0.03      0.00    2772     464   0.01  init
21:37:50        0        14      0.00      0.00    2476     128   0.00  Relay(15)
21:37:50     1000       385      0.01      0.00    5292     252   0.00  dbus-daemon
...
```
- `%MEM` - total memory usage of the process, gives an indication how much memory this process uses compared to other processes
- `RSS` - the actual physical memory allocated to the process, if growing constatnly you probably got a memory leak (does not include swap)
- `VSZ` - total amount of virtual memory allocated including swap space


