---
title: "NPM and GIT registry formats"
date: 2025-10-12
tags: ["npm", "git"]
---

# The problem

Trying to get a better understanding of the URL format of NPM and GIT package registries.

## The format

```txt
//registry.npmjs.org/:_authToken=npm_*
//npm.pkg.github.com/:_authToken=ghp_*
```

The // format is how .npmrc files specify registry-specific configuration.

Format is: `//REGISTRY/:CONFIG_KEY=value`

## Why // instead of https://?

- .npmrc uses a protocol-agnostic format
- The // means "use whatever protocol npm is configured to use" (typically HTTPS)
- Keeps the config simpler and avoids hardcoding HTTP vs HTTPS

## The trailing :

The `:` separates the registry URL from the configuration key:

`//registry.npmjs.org/:_authToken=` means "for this registry, set the `_authToken` config"

Other examples:

```txt
//registry.npmjs.org/:_authToken=token123
//registry.npmjs.org/:always-auth=true
//npm.pkg.github.com/:_authToken=ghp_abc
```

It's essentially scoping the configuration to a specific registry. Without the registry prefix, the config would apply globally to all registries.
