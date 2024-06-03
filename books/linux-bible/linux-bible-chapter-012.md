### Chapter 12

#### Understanding Disk Storge

When installing an operating system generally the `disk` is divided into one or more `partitions`. Each `partition` is then formatted with a `filesystem`.

In Linux partitions can also be formatted to be `swap areas` or `LVM (Logial Volume Manager) physical volumes`.

Each partition is then connected to the file system by `mounting` it to a point in the filesystem.

Historically the `MBR (Master Boot Record)` has been used to store size and layout information about the partitions, but later an other standard called `Globally Unique Identifier (GUID) partition tables` (in short `gpt`) emerged as part of the `UEFI` architecture.

#### Viewing Disk Partitions

There are lot of tools to handle Linux partitions, examples are:

- `fdisk` - for MBR
- `gdisk` - gdisk GPT in a CLI
- `parted` - MBR and GPT in a CLI
- `gparted` - MBR and GPT in a GUI

Below is an example partitioning of an AWS gp3 SSD using parted.

```sh
$ sudo parted -l /dev/xvda

Model: Xen Virtual Block Device (xvd)
Disk /dev/xvda: 8590MB
Sector size (logical/physical): 512B/512B
Partition Table: gpt
Disk Flags:

Number  Start   End     Size    File system  Name  Flags
14      1049kB  5243kB  4194kB                     bios_grub
15      5243kB  116MB   111MB   fat32              boot, esp
16      116MB   1074MB  957MB   ext4               bls_boot
 1      1075MB  8590MB  7515MB  ext4
```

And using fdisk.

```sh
$ sudo fdisk -l /dev/xvda

Disk /dev/xvda: 8 GiB, 8589934592 bytes, 16777216 sectors
Units: sectors of 1 * 512 = 512 bytes
Sector size (logical/physical): 512 bytes / 512 bytes
I/O size (minimum/optimal): 512 bytes / 512 bytes
Disklabel type: gpt
Disk identifier: 8F27EC90-814B-448C-88E6-B35D88412452

Device        Start      End  Sectors  Size Type
/dev/xvda1  2099200 16777182 14677983    7G Linux filesystem
/dev/xvda14    2048    10239     8192    4M BIOS boot
/dev/xvda15   10240   227327   217088  106M EFI System
/dev/xvda16  227328  2097152  1869825  913M Linux extended boot
```

#### Partition naming conventions

Devices that use the `SCSI (Small Computer System Interface)` protocol like `SATA (Serial AT Attachment)` or `USB (Universal Serial Bus)` devices are generally represented by an `sd?` (for example `/dev/sda` or `/dev/sdb`) and can have a maximum of 16 minor devices.

Devices that connect using `NVMe (Non-Volatile Memory Express)` tend to be SSD devices that connect directly to the PCIe bus of the motherboard unlike SATA drives, which use the older SATA interface. These interfaces can be divided into namespaces and partitions, for example `/dev/nvme0n1p1`.

For x86 computers disks can have up to 4 primary partitions. If you want to have more one of the partitions have to be an `extended partition` and further parititions have to be logical partitions that use the space on this extended partition.

#### Boot partitions

If we use the `lsblk` command we can see not just the devices and partitions but also where they are mounted on the filesystem.

```sh
$ lsblk
NAME     MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
loop0      7:0    0 25.2M  1 loop /snap/amazon-ssm-agent/7983
loop1      7:1    0 55.7M  1 loop /snap/core18/2812
loop2      7:2    0 38.7M  1 loop /snap/snapd/21465
xvda     202:0    0    8G  0 disk
├─xvda1  202:1    0    7G  0 part /
├─xvda14 202:14   0    4M  0 part
├─xvda15 202:15   0  106M  0 part /boot/efi
└─xvda16 259:0    0  913M  0 part /boot
```

For newer systems the first boot partition is mounted on `/boot/efi` and the remainder of the boot process is located on `/boot`. For older systems that do not use UEFI but MBR there is only a single `/boot` partition.

#### Creating a single-partition disk
