---
title: Troubleshooting Linux disk usage
author: GaborZeller
date: 2024-09-29T20-32-29Z
tags:
draft: true
---

# Troubleshooting Linux disk usage

## Check disk capacities

The first thing to check if our disk space is filling up:

```sh
df -h
```

```sh
Filesystem      Size  Used Avail Use% Mounted on
none            3.9G  4.0K  3.9G   1% /mnt/wsl
none            953G  856G   98G  90% /usr/lib/wsl/drivers
/dev/sdc        251G   35G  204G  15% /
none            3.9G  124K  3.9G   1% /mnt/wslg
none            3.9G     0  3.9G   0% /usr/lib/wsl/lib
rootfs          3.9G  2.1M  3.9G   1% /init
none            3.9G     0  3.9G   0% /run
none            3.9G     0  3.9G   0% /run/lock
none            3.9G     0  3.9G   0% /run/shm
none            3.9G     0  3.9G   0% /run/user
tmpfs           3.9G     0  3.9G   0% /sys/fs/cgroup
none            3.9G  504K  3.9G   1% /mnt/wslg/versions.txt
none            3.9G  504K  3.9G   1% /mnt/wslg/doc
C:\             953G  856G   98G  90% /mnt/c
```
Check the `Use%` column. Generally the `tmpfs` can be ignored because its Linux's in-memory file system used for very fast reads and writes.

[!TIP] - To solve the problem you need to increase the disc space or move/delete some of the data to an other partition 

## Check inode usage

Inodes store metadata about your files and directories and also where the data resides on the disk. It is possible to run out of inodes before running out of disc space.


```sh
df -ih
```

```sh
Filesystem     Inodes IUsed IFree IUse% Mounted on
none             983K     2  983K    1% /mnt/wsl
none              999 -976K  977K     - /usr/lib/wsl/drivers
/dev/sdc          16M  1.1M   15M    7% /
none             983K    34  983K    1% /mnt/wslg
none             983K     5  983K    1% /usr/lib/wsl/lib
rootfs           982K    11  982K    1% /init
none             983K     8  983K    1% /run
none             983K     1  983K    1% /run/lock
none             983K     1  983K    1% /run/shm
none             983K     1  983K    1% /run/user
tmpfs            983K    16  983K    1% /sys/fs/cgroup
none             983K    74  983K    1% /mnt/wslg/versions.txt
none             983K    74  983K    1% /mnt/wslg/doc
C:\               999 -976K  977K     - /mnt/c
```

Check the `IUse%` column to see how much of the inodes have been used.

[!TIP] - Unlike disc space inodes can not be increased only by re-formatting the filesystem or by moving/deleting files

## Check I/O wait

The issue might be that one or more processes are writing to/reading from the disc at a pace that it can not handle.

```sh
iostat -x
iostat -xd <device> # e.g iostat -xd sdc
iostat -xd <device> 1 # use the 1 flag to get per second updates
```

```sh
Linux 5.15.133.1-microsoft-standard-WSL2 (LAPTOP-PREDATOR-GABRIEL)      09/29/24        _x86_64_        (12 CPU)

avg-cpu:  %user   %nice %system %iowait  %steal   %idle
           0.15    0.00    0.07    0.01    0.00   99.77

Device            r/s     rkB/s   rrqm/s  %rrqm r_await rareq-sz     w/s     wkB/s   wrqm/s  %wrqm w_await wareq-sz     d/s     dkB/s   drqm/s  %drqm d_await dareq-sz     f/s f_await  aqu-sz  %util
sda              0.05      3.16     0.02  24.50    0.38    67.80    0.00      0.00     0.00   0.00    0.00     0.00    0.00      0.00     0.00   0.00    0.00     0.00    0.00    0.00    0.00   0.00
sdb              0.00      0.04     0.00   0.00    0.08    23.22    0.00      0.00     0.00   0.00    0.50     2.00    0.00      0.00     0.00   0.00    0.00     0.00    0.00    0.00    0.00   0.00
sdc              3.93     50.76     1.20  23.47    0.18    12.93    1.23     65.44     6.33  83.75    1.86    53.32    0.31     40.34     0.01   2.78    0.10   131.78    0.17    0.37    0.00   0.29
```

Look for the `r_await` and `w_await` to see if you have pending writes and reads that are throttling the disc. Looking at the `r/s` or `w/s` numbers is also not a bad idea to get an idea what is going on with the disc.

### Which processes have heave I/O usage

If the issue is related to I/O then you need to identify the processes that are causing this.

```sh
pidstat -d | head -20
```

```sh
Linux 5.15.133.1-microsoft-standard-WSL2 (LAPTOP-PREDATOR-GABRIEL)      09/29/24        _x86_64_        (12 CPU)

21:51:09      UID       PID   kB_rd/s   kB_wr/s kB_ccwr/s iodelay  Command
21:51:09     1000        15      2.34      0.00      0.00       0  zsh
21:51:09     1000       128      0.11      0.00      0.00       0  tmux: client
21:51:09     1000       130      4.59      5.81      4.99       0  tmux: server
21:51:09     1000       429     10.81      0.92      0.15       0  zsh
21:51:09     1000      6902     21.84     48.60     12.12       0  zsh
21:51:09     1000     12674      5.76      9.11      2.26       0  zsh
21:51:09     1000     28054      0.00      0.01      0.00       0  zsh
21:51:09     1000     29874      0.00      0.00      0.00       0  pnotes
21:51:09     1000     29886      0.00      0.10      0.00       0  nvim
21:51:09     1000     30863      0.00      0.00      0.00       0  zsh
```
Check the `kB_rd/s` and `kB_wr/s` columns to identify read and write heavy processes.

[!TIP] - If there are processes that are choking your disc then either you have to scale down your process(es) or consider using a disc with a better I/O performance.

## Check the file descriptor limits

File desriptors are unique identifiers for I/O resource or an abstract handler for an open file. You can run out of these like you can run out of inodes and you can get a `Too many open files` error when trying to create file descriptors beyond the limit.

```sh
cat /proc/<pid>/limits # pick any process id
```

```sh
Limit                     Soft Limit           Hard Limit           Units
Max cpu time              unlimited            unlimited            seconds
Max file size             unlimited            unlimited            bytes
Max data size             unlimited            unlimited            bytes
Max stack size            8388608              unlimited            bytes
Max core file size        0                    unlimited            bytes
Max resident set          unlimited            unlimited            bytes
Max processes             31410                31410                processes
Max open files            1024                 1048576              files
Max locked memory         67108864             67108864             bytes
Max address space         unlimited            unlimited            bytes
Max file locks            unlimited            unlimited            locks
Max pending signals       31410                31410                signals
Max msgqueue size         819200               819200               bytes
Max nice priority         0                    0
Max realtime priority     0                    0
Max realtime timeout      unlimited            unlimited            us
```

Check the `Max open files` row that will show you how many open files are allowed.

To see how many file descriptors a particular process has open:

```sh
ls -1 /proc/<pid>/fd | wc -l
```

[!TIP] - If you want to increase the file descriptor limit you can do that by editing the /etc/security/limits.conf file

```sh
# Add or modify the following lines
* soft nofile 4096
* hard nofile 8192

```
[!TIP] - You can also edit the /etc/sysctl.conf file

Add the following line:

```sh
fs.file-max = 65536
```

Apply the changes:

```sh
sudo sysctl -p 
```

Verify the changes:

```sh
ulimit -n
```
