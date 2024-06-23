### Chapter 12

#### Understanding Disk Storge

When installing an operating system generally the `disk` is divided into one or more `partitions`. Each `partition` is then formatted with a `filesystem`.

In Linux partitions can also be formatted to be `swap areas` or `LVM (Logial Volume Manager) physical volumes`.

Each partition is then connected to the file system by `mounting` it to a point in the filesystem.

Historically the `MBR (Master Boot Record)` has been used to store size and layout information about the partitions, but later an other standard called `Globally Unique Identifier (GUID) partition tables` (in short `gpt`) emerged as part of the `UEFI` architecture.

#### Viewing Disk Partitions

There are lot of tools to handle Linux partitions, examples are:

- `fdisk` - for MBR
- `gdisk` - GPT in a CLI
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

#### Partitioning disks

To add new storage to a Linux machine you generally need to follow four steps.

1. Install the new hard drive
2. Partition the hard drive
3. Create file system on the partitioned hard drive
4. Mount the file system

Generally for partitioning there are two tools that can be used on Linux. One is `parted` the other one is `fdisk`.

The former applies the individual changes immediately while with the latter you have to specifiy all the changes before comitting them.

An other useful tool before partitioning a disk is `df` (display filesystem) that can be used to check if the device already has a file system on it.

```sh
$ df -hT /dev/sdc
Filesystem     Type  Size  Used Avail Use% Mounted on
/dev/sdc       ext4  251G   36G  204G  15% /
```

##### Example using `parted`

1. run `parted /dev/<device>` - start partitioning
2. help - show cavailable commands
3. print - show details of disk
4. rm - will prompt for a partition number you want to delete (e.g. 1)
5. mklabel - will prompt for partition label type (e.g. gpt)
6. mkpart - create new partition, will prompt for the name of the partition, file system type and start/end og the partition (this has to be ran multiple times to create all partitions)

Once done create a file system on each partition

```sh
mkfs -t ext4 /dev/sdb1
mkfs -t ext4 /dev/sdb2
mkswap /dev/sfb3 # If you want to create swap space
```

Once done create a folder where you want to mount the new file system and mount it.

```sh
mkdir /mnt/test
mount /dev/sdb1 /mnt/test
```

If you want to un-mount run

```sh
umount /dev/sdb1
```

##### Example using `fdisk`

...

#### Extending partitions

On AWS if you run into the issue of an EBS being too small you have the opportunity to extend them.

First you have to increase the size of the volume through the console. Once the process is finished and list the block devices you will see the EBS volume's size being bigger.

In the below case it was the `xvde` device that got increased.

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
xvde     202:64   0    5G  0 disk
└─xvde1  202:65   0 1022M  0 part /mnt/test
```

Once the EBS's size has been increased the next step is to

```sh
$ growpart /dev/xvde 1
CHANGED: partition=1 start=2048 old: size=2093056 end=2095103 new: size=10483679 end=10485726
```

```sh
resize2fs /dev/xvde1
```

#### Using Logical Volumes

Logical volumes are a better way to solve the problem of running out of disk space without downtime. Logical volumes allow you to create pools of spaces called `volume groups` and assign physical disk partitions to those groups. Once done you can assign space from these volume groups to logical volumes.

Some benefits of volume groups:

1. Add more space to logical volume from volume group while the logical volume is still in use
2. Add more physical space to the volume group if you start running out of space
3. Move data from smaller physical volumes to larger ones for more efficient disk use

To create a volume group have a physical volume ready.

```sh
vgcreate myvg0 /dev/xvde
```

To see the new volume group.

```sh
$ vgdisplay myvg0

--- Volume group ---
VG Name               myvg0
System ID
Format                lvm2
Metadata Areas        1
Metadata Sequence No  1
VG Access             read/write
VG Status             resizable
MAX LV                0
Cur LV                0
Open LV               0
Max PV                0
Cur PV                1
Act PV                1
VG Size               <5.00 GiB
PE Size               4.00 MiB
Total PE              1279
Alloc PE / Size       0 / 0
Free  PE / Size       1279 / <5.00 GiB
VG UUID               Ceqc5j-4SAO-grFV-HyZa-zFkl-VjS1-p5OWBl
```

After having a logical volume group you can create a logical volume from the space.

```sh
sudo lvcreate -n music -L 1G myvg0
```

The code will create a new device that can be found under `/dev/mapper`

```sh
$ ls /dev/mapper/myvg0*

/dev/mapper/myvg0-music
```

After that a file system can be created on the logical volume then mounted on the file system.

```sh
$ lsblk

NAME          MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
loop0           7:0    0 25.2M  1 loop /snap/amazon-ssm-agent/7983
loop1           7:1    0 55.7M  1 loop /snap/core18/2812
loop2           7:2    0 38.7M  1 loop /snap/snapd/21465
loop3           7:3    0 38.8M  1 loop /snap/snapd/21759
loop4           7:4    0 55.7M  1 loop /snap/core18/2823
xvda          202:0    0    8G  0 disk
├─xvda1       202:1    0    7G  0 part /
├─xvda14      202:14   0    4M  0 part
├─xvda15      202:15   0  106M  0 part /boot/efi
└─xvda16      259:0    0  913M  0 part /boot
xvde          202:64   0    5G  0 disk
└─myvg0-music 252:0    0    1G  0 lvm # The new logical volume
```

```sh
$ mkfs -t ext4 /dev/mapper/myvg0-music
$ mkdir /mnt/music
$ mount /dev/mapper/myvg0-music /mnt/music
$ df -h /mnt/music/

Filesystem               Size  Used Avail Use% Mounted on
/dev/mapper/myvg0-music  974M   24K  907M   1% /mnt/music
```

#### Growing Logical Volumes

Let's take a look at how our logical volume looks like after creating the logical volume.

```sh
$ vgdisplay myvg0

--- Volume group ---
VG Name               myvg0
System ID
Format                lvm2
Metadata Areas        1
Metadata Sequence No  2
VG Access             read/write
VG Status             resizable
MAX LV                0
Cur LV                1
Open LV               1
Max PV                0
Cur PV                1
Act PV                1
VG Size               <5.00 GiB
PE Size               4.00 MiB
Total PE              1279
Alloc PE / Size       256 / 1.00 GiB
Free  PE / Size       1023 / <4.00 GiB
VG UUID               Ceqc5j-4SAO-grFV-HyZa-zFkl-VjS1-p5OWBl
```

In order to expand the logical volume we use the `lvextend` command.

```sh
lvextend -L +1G /dev/mapper/myvg0-music
```

The above successfullt extends the size of the volume but the file system also needs to be adjusted accordingly.

```sh
resize2fs /dev/mapper/myvg0-music
```

And verify that the file system was adjusted.

```sh
$ df -h /mnt/music/

Filesystem               Size  Used Avail Use% Mounted on
/dev/mapper/myvg0-music  2.0G   24K  1.9G   1% /mnt/musi
```

#### Checking for existing LVM

```sh
$ pvdisplay /dev/xvde

--- Physical volume ---
PV Name               /dev/xvde
VG Name               myvg0
PV Size               5.00 GiB / not usable 4.00 MiB
Allocatable           yes
PE Size               4.00 MiB
Total PE              1279
Free PE               767
Allocated PE          512
PV UUID               dslBQU-8lgn-K7In-pDgH-CPPC-SoJD-VN4hLc
```

```sh
$ vgdisplay myvg0

--- Volume group ---
VG Name               myvg0
System ID
Format                lvm2
Metadata Areas        1
Metadata Sequence No  3
VG Access             read/write
VG Status             resizable
MAX LV                0
Cur LV                1
Open LV               1
Max PV                0
Cur PV                1
Act PV                1
VG Size               <5.00 GiB
PE Size               4.00 MiB
Total PE              1279
Alloc PE / Size       512 / 2.00 GiB
Free  PE / Size       767 / <3.00 GiB
VG UUID               Ceqc5j-4SAO-grFV-HyZa-zFkl-VjS1-p5OWBl
```

```sh
$ lvdisplay myvg0

--- Logical volume ---
LV Path                /dev/myvg0/music
LV Name                music
VG Name                myvg0
LV UUID                G40qgf-ICVQ-2blb-YUsB-O0BW-UzvR-c49vOn
LV Write Access        read/write
LV Creation host, time ip-172-31-21-38, 2024-06-23 15:41:27 +0000
LV Status              available
# open                 1
LV Size                2.00 GiB
Current LE             512
Segments               1
Allocation             inherit
Read ahead sectors     auto
- currently set to     256
Block device           252:0
```
