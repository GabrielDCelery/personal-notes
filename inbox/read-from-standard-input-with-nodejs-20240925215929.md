---
title: Read from standard input with nodejs
author: GaborZeller
date: 2024-09-25T21-59-29Z
tags:
draft: true
---

# Read from standard input with nodejs

```javascript
const readline = require('readline');

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

rl.on('line', (input) => {
  console.log(input);
});

rl.on('close', () => {
  console.log('EOF reached');
});
```
