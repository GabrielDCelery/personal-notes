---
title: Fix logical volume interfering with Ubuntu installation
author: GaborZeller
tags:
  - linux
  - ubuntu
  - proxmox
---

# The problem

Wanted to remove Proxmox from my homelab machine and just install Ubuntu server. The installation kept crashing at th slection of which hard disk I wanted to install the operating system on.

# Identifying the root cause

Booted into the machine using `USB drive (Ventoy)` and during installation after the network card setup but before selecting the SSD pressed `Ctrl+Alt+F2` to get to shell. Ran `lsbk` to see what was going on. Showed us the disk structure and identified nvme0n1 with all the Proxmox partitions.

```sh
# lsblk

NAME                 MAJ:MIN RM   SIZE RO TYPE MOUNTPOINTS
sda                    8:0    0 953.9G  0 disk
sdb                    8:16   0 953.9G  0 disk
nvme0n1              259:0    0 465.8G  0 disk
├─nvme0n1p1          259:1    0  1007K  0 part
├─nvme0n1p2          259:2    0     1G  0 part /boot/efi
└─nvme0n1p3          259:3    0   464G  0 part
├─pve-swap         252:0    0     8G  0 lvm  [SWAP]
├─pve-root         252:1    0    96G  0 lvm  /
├─pve-data_tmeta   252:2    0   3.4G  0 lvm
│ └─pve-data-tpool 252:4    0 337.1G  0 lvm
│   └─pve-data     252:5    0 337.1G  1 lvm
└─pve-data_tdata   252:3    0 337.1G  0 lvm
└─pve-data-tpool 252:4    0 337.1G  0 lvm
└─pve-data     252:5    0 337.1G  1 lvm
```

Basically I had a logical volume setup that had to be removed (like a layered onion)

```sh
# Theory
Logical Volumes (LV)     ← Top layer - actual usable volumes
    ↓ (exist within)
Volume Groups (VG)       ← Middle layer - pool of storage
    ↓ (created from)
Physical Volumes (PV)    ← Bottom layer - actual disks/partitions
```

```sh
# What I had
/dev/nvme0n1p3           ← Physical partition
    ↓
pve (Physical Volume)    ← LVM metadata added to partition
    ↓
pve (Volume Group)       ← Pool named "pve"
    ↓
swap, root, data, etc.   ← Logical volumes carved from the pool
```

Removal Order

```sh
# 1. Remove logical volumes (top layer)
sudo lvremove -f /dev/pve/swap
sudo lvremove -f /dev/pve/root
# ... or let vgremove do it which is what I did but disabled it with lvchange -an pve

# 2. Remove volume group (middle layer)
# Removes the "pve" group and all volumes
sudo vgremove -f pve

# 3. Remove physical volume (bottom layer)
# Removes LVM markers from the partition itself
# Without pvremove, the partition would still have LVM metadata on it, which could confuse the installer or prevent proper reformatting.
sudo pvremove -f /dev/nvme0n1p3
```

Key Differences

| Command  | Target          | What it removes                      | Analogy                                                       |
| -------- | --------------- | ------------------------------------ | ------------------------------------------------------------- |
| lvremove | Logical volume  | Individual volumes (like swap, root) | Delete individual virtual driver carved from the pool         |
| vgremove | Volume Group    | Entire group + all LVs inside it     | Delete the storage pool that groups drives together           |
| pvremove | Physical volume | LVM metadata from disk/partition     | Remove "This is an LVM disk" sticker from the physical driver |

# The actual execution / solution

3.  Turn off swap (if it shows as active)

```sh
sudo swapoff -a
```

Turns off the active swap partition so it can be removed (can't remove if in use)

4.  Deactivate all LVM logical volumes

```sh
sudo lvchange -an pve
```

Makes logical volumes inactive so they can be deleted. The -a means "activate" and n means "no" (deactivate).

5. Remove the volume group

```sh
sudo vgremove -f pve
```

This deletes all the logical volumes (swap, root, data, VM disks) in one go. The -f forces it without asking for confirmation.

6. Remove the physical volume

```sh
sudo pvremove -f /dev/nvme0n1p3
```

This removes the LVM metadata from partition 3, telling the system it's no longer part of LVM.

7. Wipe all filesystem signatures from the drive

```sh
sudo wipefs -a /dev/nvme0n1
```

Removes any trace of filesystems, partition tables, LVM markers, etc. The -a means "all signatures".

8. Destroy the partition table completely

```sh
sudo sgdisk --zap-all /dev/nvme0n1
```

Completely erases the partition structure. This is like wiping the table of contents, making the disk appear completely blank.

9. Verify the drive is clean (should show nvme0n1 with no partitions)

```sh
lsblk
```

10. switch back to installation `Ctrl+Alt+F1` (though I just rebooted the machine)
