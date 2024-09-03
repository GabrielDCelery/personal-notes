---
title: EBS
author: Gabor Zeller
date: 2024-09-01T16:46:13Z
tags: ['aws']
draft: true
---

# EBS

#### What kind of offerings does EBS has?

- SSD gp2 and gp3 for standard application usage
- SSD io1/io2/io2 block express generally for very fast computation processes
- HDD throughput optimized
- HDD cols storage

#### How does the GP2 EBS work?

AWS offers a General Purpose SSD called `gp2` and `gp3`. In terms of size these van vary between `1GB` and `16TB` size.

The data is transferred in and out from EBS in `chunks`. One chunk is `16Kb` in size. Transferring a single chunk is `1 IO` operation. So `1 IOPS = 1 IO in 1 second`.

For example if you transfer a `160Kb` file within a second that is `10 IOPS`.

When creating an EBS volume it gets a bucket that can hav a max capacity of `5.4 million` IO Credits and gets refilled at the base rate of `100IOPS` plus based on the size of the volume which is `3 IO credit per second per GB`.

In terms of limitations

- You cannot do data transfers if your bucket credit is empty
- The burst limit on an EBS is `3000 IOPS` which means you cannot go beyond that

For volumes that are `1TB` or greater than that in size their bucket refill rate is `3 * 1000 = 3000` which means their refill rate is greater than the baseline burst performance which means they will refill faster than you could write/read data from them. This also applies to the baseline performance, so if you provision a larger disk you can achieve up to `16000 IOPS`.

#### How does the GP3 EBS work?

`GP3` replaces the old credit system and comes with a much more simplified solution where the baseline performance offers `3000 IOPS` and `125 MiB/s` data transfer. This is burtsable to `16000 IOPS` and `1000 MiB/s` data transfer. For GP3 these don't come automatically they have to be added.

#### How does provisioned IOPS SSD work?

There are currently three variations of this block storage, `io1`, `io2` and `io2 block express`. The unique feature they all have is that their IO can be independently adjusted from the volume size.

The io1 version can have up to `64000 IOP` which is `4 times` greater than that `gp2 and gp3` can achieve and can have a data transfer rate of `1000 MB/s`. The block express version can have up to `256000 IOPS` and `4000 MBs`.

I terms of size it can have `4GB-16TB` for io1/io2 and `4GB-64TB` for the block express variation.

#### IOPS considerations for EC2 instances

Generally an EC2 instance's IOPS speed is much higher than the EBS volme that you are provisioning. If you want to max out the IOPS on an EC2 instance and reach `260,000 IOPS` and `7500 MB/s` transfer speeds you will either need multiple EBS volumes (for example 4 for io1 EBS volumes) or use a single io2 block express device which was specifically designed fo maxing out the IOPS capabilities on an EC2 instance.

#### How does the HDD based storage work?

AWS has two offerings for HDD devices, one of them is a `throughput optimized` volme, the other is `cold storage`. They are not good for random access, generally used for continous writes. In terms of size they vary between `125GB` and `16TB` and their write throughut is limited to `500 IOPS`. It is important to note that unlike SSD where one IO block is 16Kb for HDD one block is `1 MB`, which means the write throughput is `500 MB/s`.

#### Exam tips

The IO Credit Bucket starts full, which means for `30 minutes` you can max out the `3000 IOPS` usage.


