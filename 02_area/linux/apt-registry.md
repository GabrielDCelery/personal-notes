# The problem

Wanted to have a better understanding of the apt system.

# The details

There are two main parts to this

- where apt looks for packages
- how it verifies if the package is legit

## Where apt looks

There is the `/etc/sources.list.d` that has the files for configuring this e.g.

```sh
# ls -la /etc/sources.list.d

total 28
drwxr-xr-x 2 root root 4096 Jan 23 22:01 .
drwxr-xr-x 9 root root 4096 Jan 23 21:30 ..
-rw-r--r-- 1 root root  110 Jan 23 21:30 docker.list
-rw-r--r-- 1 root root  107 Jan 23 21:30 mise.list
-rw-r--r-- 1 root root  285 Jan 23 22:01 nvidia-container-toolkit.list
-rw-r--r-- 1 root root  386 Jan 18 12:35 ubuntu.sources
-rw-r--r-- 1 root root 2552 Aug  5 17:02 ubuntu.sources.curtin.orig

```

## How does apt know what to get and how it verifies

Then there is the APT repository source file that tells APT where to download Docker packages from:

```sh
# cat docker.list
deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu noble stable
```

| Part                                     | Meaning                                                                                                   |
| ---------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| deb                                      | the name of the package repository (not the source code!)                                                 |
| [arc=amd64]                              | Only download packages for the 64-bit x86 architecture                                                    |
| [signed-by=/etc/apt/keyrings/docker.asc] | use this GPG key verify package signature for security (the key is stored locally at `/etc/apt/keyrings`) |
| https://download.docker.com/linux/ubuntu | the repository url                                                                                        |
| noble                                    | the Ubuntu version                                                                                        |
| stable                                   | the build we want to install (not nightly or test builds)                                                 |

## Step-by-step

1. When you run apt update, APT fetches the package index from https://download.docker.com/linux/ubuntu/dists/noble/stable/
2. APT verifies the downloaded index is signed by the key at /etc/apt/keyrings/docker.asc (prevents tampering)
3. When you run apt install docker-ce, APT downloads the package from this repository and verifies its signature before installing

> [!NOTE] The /etc/apt/keyrings/ directory is the modern recommended location for storing repository signing keys (replacing the deprecated apt-key method).
