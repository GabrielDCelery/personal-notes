---
title: "OpenSSL encoding"
date: 2025-11-04
tags: ["openssl"]
---

# The issue

There are a lot of different ways one can look at a key/cert/pem so decided to delve a bit deeper into the various data structures.

The encoding layers:

1. Raw key data → Binary format (DER - Distinguished Encoding Rules)
2. DER → Base64 encoded → PEM format
3. Base64 decode → Gets you back to binary DER, which looks like gibberish

## Base64 version of key

```sh
grep -v "BEGIN\|END" fd-public.key

# output
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsAyEFvWPaX5rktS9SC5M
cS2w6s+IyUKsWrc39x6rRg1cg6caFOTPFfK/K6juzAU7OPHB1sb+Cb/5tV3ZA1+g
wQR5tqb642/Ma2i+UjKQFgfG66i+w/P0fV1ggHop7Y7MjL9+0PUiBbLu+TFqnWOe
7wKtn4Fs7E9Hk2QSyb1m40/5qjtNoqDN4HnxA/j1R2BFe/gIGSmQeJFcZrieW6lL
wFwLwLLBEP7wmsK5Prlhji45TTVHwG0a6GVV0fAiNiHWxlaybxEo3M66J6xPCd0l
yoYwwL5Nyvk/qEPCl5UalRVfBCUSHHZNcV+TeFADf/Rpvxm5qMavdQlU/944GG5x
swIDAQAB
```

## Hex version of key

```sh
grep -v "BEGIN\|END" fd-public.key | base64 -d | xxd

# output
00000000: 3082 0122 300d 0609 2a86 4886 f70d 0101  0.."0...*.H.....
00000010: 0105 0003 8201 0f00 3082 010a 0282 0101  ........0.......
00000020: 00b0 0c84 16f5 8f69 7e6b 92d4 bd48 2e4c  .......i~k...H.L
00000030: 712d b0ea cf88 c942 ac5a b737 f71e ab46  q-.....B.Z.7...F
00000040: 0d5c 83a7 1a14 e4cf 15f2 bf2b a8ee cc05  .\.........+....
00000050: 3b38 f1c1 d6c6 fe09 bff9 b55d d903 5fa0  ;8.........].._.
00000060: c104 79b6 a6fa e36f cc6b 68be 5232 9016  ..y....o.kh.R2..
00000070: 07c6 eba8 bec3 f3f4 7d5d 6080 7a29 ed8e  ........}]`.z)..
00000080: cc8c bf7e d0f5 2205 b2ee f931 6a9d 639e  ...~.."....1j.c.
00000090: ef02 ad9f 816c ec4f 4793 6412 c9bd 66e3  .....l.OG.d...f.
000000a0: 4ff9 aa3b 4da2 a0cd e079 f103 f8f5 4760  O..;M....y....G`
000000b0: 457b f808 1929 9078 915c 66b8 9e5b a94b  E{...).x.\f..[.K
000000c0: c05c 0bc0 b2c1 10fe f09a c2b9 3eb9 618e  .\..........>.a.
000000d0: 2e39 4d35 47c0 6d1a e865 55d1 f022 3621  .9M5G.m..eU.."6!
000000e0: d6c6 56b2 6f11 28dc ceba 27ac 4f09 dd25  ..V.o.(...'.O..%
000000f0: ca86 30c0 be4d caf9 3fa8 43c2 9795 1a95  ..0..M..?.C.....
00000100: 155f 0425 121c 764d 715f 9378 5003 7ff4  ._.%..vMq_.xP...
00000110: 69bf 19b9 a8c6 af75 0954 ffde 3818 6e71  i......u.T..8.nq
00000120: b302 0301 0001                           ......
```

Format breakdown

```sh
00000000: 3082 0122 300d 0609 2a86 4886 f70d 0101 0.."0...\*.H.....
│ │ │
│ │ └─ ASCII representation
│ └─ Hexadecimal bytes (the actual data)
└─ Offset (position in the file)
```

1. Left column (00000000, 00000010):

- Byte offset from the start of the file
- In hexadecimal
- 00000000 = byte 0, 00000010 = byte 16 (hex 10 = decimal 16)

2. Middle section (3082 0122 300d...):

- The actual data in hexadecimal
- Each pair is one byte (e.g., 30, 82, 01, 22)
- This is your DER-encoded public key structure

3. Right column (0.."0...\*.H.....):

- ASCII representation of the hex bytes
- Printable characters are shown as-is
- Non-printable bytes shown as .
- This is mostly gibberish for binary data, but can be helpful for finding text strings

## DER version of key

```sh
openssl asn1parse -inform DER -in <(grep -v "BEGIN\|END" fd-public.key | base64 -d)

# output
 0:d=0  hl=4 l= 290 cons: SEQUENCE
 4:d=1  hl=2 l=  13 cons: SEQUENCE
 6:d=2  hl=2 l=   9 prim: OBJECT            :rsaEncryption
17:d=2  hl=2 l=   0 prim: NULL
19:d=1  hl=4 l= 271 prim: BIT STRING
```

DER is a binary encoding format for structured data, specifically a strict subset of ASN.1 (Abstract Syntax Notation One). It's used extensively in cryptography for certificates, keys, and other PKI data.

DER uses Tag-Length-Value (TLV) encoding. [Tag - what type of data] [Length - the length of the data] [Value - the actual data]

Example breakdown:

```sh
02 03 01 00 01
│ │ └──────── Value: 0x01 0x00 0x01 (65537 in decimal)
│ └─────────── Length: 3 bytes
└────────────── Tag: 02 = INTEGER
```

## PEM stands for Privacy Enhanced Mail.

It's a bit of a historical quirk - PEM was originally developed in the 1990s as a standard for securing email, but it never really took off for that purpose. However, the encoding format it defined became extremely popular and is now the de facto standard for encoding certificates, keys, and other cryptographic data.

### What PEM actually is:

Base64-encoded data wrapped in header/footer lines:

```sh
-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKJ5pq... (base64 encoded data)
...
-----END CERTIFICATE-----
```

### Why use PEM?

- Text-based - can be easily copied/pasted, viewed in text editors
- Email-safe - can be sent in emails without corruption
- Widely supported - most tools and systems understand PEM format
- Human-readable headers - you can see what type of data it contains

### Other formats you might see

- .der - Binary encoded version (same data, just binary instead of base64)
- .crt/.cer - Usually PEM or DER encoded certificates
- .key - Usually PEM encoded private keys
- .p12/.pfx - PKCS#12 format (can bundle cert + key together)
