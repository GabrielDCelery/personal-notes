---
title: Using Claude within an org
author: GaborZeller
tags:
  - claude
---

# The problem

Wanted to figure out how Claude is worth using within an organization.

# The answers

## How to organize rules and restrictions

[They got a pretty good when to use section](https://code.claude.com/docs/en/settings#when-to-use-each-scope)

## Sandbox mode

Sandbox mode is not designed for enforcing restrictions, there is no default "launch in sandbox mode" or "you can't run except in sandbox mode" feature.

Even when it comes to permissions `Filesystem and network restrictions are configured via Read, Edit, and WebFetch permission rules, not via these sandbox settings.` which means besides the few flags like `excludedCommands`, `network.allowedDomains` or `network.allowUnixSockets` there is no such things as setting allow/deny rules outside of sandbox and setting allow/deny inside of sandbox.

## MCP servers

We can enforce company-wide rules using `managed-mcp.json` that is pretty good. [How to control MCP access](https://code.claude.com/docs/en/mcp#managed-mcp-configuration)

## Permission gatchas

You want to restrict access to `~/.ssh`? Get ready to do this

```sh
"Bash(*~/.ssh*)",
"Glob(*~/.ssh*)",
"Read(~/.ssh/*)",
"Edit(~/.ssh/*)",
"Write(~/.ssh/*)"
```

And even this gets bypassed using `Glob`

```sh
- pattern: **/*
- path: /home/gaze/.ssh
```

So yeah the below approach worked but who knows if I could have further messed with it

```sh
"Bash(**/.ssh*)",
"Bash(*ssh*)",
"Glob(**/.ssh/**)",
"Read(**/.ssh/**)",
"Edit(**/.ssh/**)",
"Write(**/.ssh/**)"
```
