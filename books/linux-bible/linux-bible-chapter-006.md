### Chapter 6

#### Understanding processes on Linux

Linux is both a multiuser and multitasking system which means the same process can be ran by multiple users at the same time and even the same user can run the same process in multiple instances.

Each process gets allocated a unique process ID (`PUID`). No two processess can share the same ID, but once a process stops the ID gets freed up and an other process can claim it.

Each process also gets associated with a `user` and `group` which determines what system resources the process has access to.

#### Listing processes

The two most common commands to list running processes are `ps` and `top`. The `ps` command returns a snapshot, the `top` command is more of an interactive utility tool.

```sh
$ ps u

USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
ubuntu      1102  0.4  1.1   9060  5376 pts/0    Ss   20:51   0:00 -bash
ubuntu      1151  0.0  0.9  11320  4352 pts/0    R+   20:51   0:00 ps u
```

`USER` - the user who started the process
`CPU/MEM` - the amount of cpu and memory the process is using
`VSZ/RSS` - the amount of memory the process can use in theory and does actually use in kilobytes
`STAT` - represents the state of the process, in the example `R` stands for running, `S` for sleeping and `+` indicates foreground
`START` - when the process started

#### Working with ps

For easier readability the pipe the ourtput of ps to less.

```sh
$ ps u | less
```

The command also allows to list specific columns using the `-o` options flag.

```sh
$ ps -o pid,user,uid,vsz,rss,comm

    PID USER       UID    VSZ   RSS COMMAND
   1102 ubuntu    1000   9060  5248 bash
   1333 ubuntu    1000  11288  4352 ps
```

The command also allows for sorting using the `--sort=column_name` flag.

```sh
$ ps -o pid,user,uid,vsz,rss,comm --sort=vsz

    PID USER       UID    VSZ   RSS COMMAND
   1102 ubuntu    1000   9060  5248 bash
   1333 ubuntu    1000  11288  4352 ps
```

#### Working with top

```sh
ubuntu@ip-172-31-29-195:~$ top
top - 21:36:14 up 48 min,  1 user,  load average: 0.00, 0.00, 0.00
Tasks: 100 total,   1 running,  99 sleeping,   0 stopped,   0 zombie
%Cpu(s):  0.0 us,  0.0 sy,  0.0 ni,100.0 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st
MiB Mem :    454.2 total,    126.2 free,    170.6 used,    186.8 buff/cache
MiB Swap:      0.0 total,      0.0 free,      0.0 used.    283.6 avail Mem

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
      1 root      20   0   22508  13524   9556 S   0.0   2.9   0:04.46 systemd
      2 root      20   0       0      0      0 S   0.0   0.0   0:00.00 kthreadd
      3 root      20   0       0      0      0 S   0.0   0.0   0:00.00 pool_workqueue_relea+
      4 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 kworker/R-rcu_g
      5 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 kworker/R-rcu_p
      6 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 kworker/R-slub_
      7 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 kworker/R-netns
     10 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 kworker/0:0H-events_+
     12 root       0 -20       0      0      0 I   0.0   0.0   0:00.00 kworker/R-mm_pe
```

The top command also lists processes by relevant information and gets refreshed every 5 second.

Some useful hotkeys while using top.

`h` - help
`M/P` - sort processes by memory or processor
`R` - reverse sorting order
`u` - after pressing type in the username press enter and filter down to processed that belong to the user

`k` - allows for killing the process. Type in the process ID then the signal `15` for `graceful termination` and `9` for `killing outright`
`r` - allows for renicing the process. Type in the process ID then the renice value between `-20` and `19`

#### Working with System Monitor

When having access the GNOME desktop a Windows like task manager can be used for managing processes.

#### Working with foreground and background processes

When using a "dumb" terminal you can only see one active process at a time. The active process you can see in the terminal is in the foreground, if you want to work with multiple processes you can move them into the background.

When having an actively running process it can be sent to the background using `Ctrl-Z` (for example neovim).

The other way is to add the `&` symbol to the end of a running command.

```sh
$ find /usr > /tmp/alluserfiles &
[2] 1397
```

When putting the command to the background the job ID and process ID both get printed.

Jobs in the background can be listed using the `jobs` command.

```sh
$ jobs

[1]+  Stopped                 vi
[2]-  Running                 find /usr > /tmp/alluserfiles &
```

Commands can be brought to the foreground using the following command

```sh
$ fg %1
```

#### Sending signals to processes

In Linux the `kill` command can be used to send different types of signals to running processes.

The most common signals are:

`SIGTERM (15)`- try to gracefully kill the process
`SIGKILL (9)` - kill the process flat out
`SIGHUP (1)` - reloads a configuration for a process
`SIGSTOP` - pauses a process
`SIGCONT` - continues a process

Ways to send a signal to a process (the default signal is SIGTERM if it is not specified):

```sh
$ kill 10432
$ kill -15 10432
$ kill -SIGKILL 10432
```

An alternative way to send signals is to use `killall <processname>` to send a signal to multiple processes sharing the same name.

#### Setting process priority with niceness

When the operating system tries to make a decision which process to allocate CPU resources to it uses the process's niceness value. The less nice processes are the more likely they get access to the CPU. A process's niceness value can vary between -20 and 19 and by default each process starts at 0.

When it comes to setting a process's niceness value there are some rules.

- A regular user can only modify their own process
- A regular user cannot set negative niceness on a process
- A regular user can only increase niceness, they cannot reduce it
- A root user can modify any process to any niceness value

To start a process with a specific niceness value run:

```sh
$ nice -n +5 updatedb &
```

The above will start the updatedb command in the background with a niceness value 5.

If we wanted to change the niceness value of a running process we would run the below command as the root user.

```sh
$ renice -n -5 20845
```

#### Limiting processes with cgroups

While niceness is a simple way to manage the CPU usage of single processes owned by a user it cannot control the niceness of child processes that the process might start up and is not robust enough if we want to have a system-wide control of who can access what.

This is where cgroups come into play. They can be used to identify processes and tasks belonging to a particular control group (even child processes) and to determine what resources they have access to and at what capacity.

Examples:

- storage (blkio) - limiting IO access to the block device
- processor (cpu) - how much cpu the process has access to
- cpu assignment (cpuset) - when there are multiple cpu cores this can be used to assign a task to a particular core
- device access (devices) - allows tasks to open and create (mknod) selected device types
- suspend (freezer) - suspends and resumes tasks belonging to cgroup tasks
- memory (memory) - limits memory usage
- network bandwidth (net_cls) - limits network access by inspecting packets
- network priorities (net_prio) - sets priorities of network traffic

These things are generally set up by editing configuration files in

- `/etc/cgconfig.conf` - creating cgroups
- `/etc/cgrules.conf` - configuring what cgroups can do
