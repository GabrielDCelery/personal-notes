---
title: "Git packages"
date: 2025-09-28
tags: ["git"]
---

# Problem

Wanted to have a better understanding of Git package management.

## The quickstart project

First followed the steps on the official Github [Quickstart for GitHub Packages](https://docs.github.com/en/packages/quickstart) to set up a repo [demo-github-packages](https://github.com/GabrielDCelery/demo-github-packages).

After publishing the package created a classic `github token` with `read:packages` access and added it to the `.npmrc` file to be able to install the package `//npm.pkg.github.com/:_authToken=${mytoken}`
