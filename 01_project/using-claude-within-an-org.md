---
title: Using Claude within an org
author: GaborZeller
tags:
  - claude
---

# The problem

Wanted to figure out how Claude is worth using within an organization and how the actual distribution of it would like withing the org.

# The answers

## Org level restrictions

Using [Claude Teams](https://code.claude.com/docs/en/iam#claude-for-teams-or-enterprise) we can invite members via email. There is also Claude Console but that is the API based billing.

## How to organize rules and restrictions

[They got a pretty good when to use section](https://code.claude.com/docs/en/settings#when-to-use-each-scope). For system level controls there is the `managed-settings.json` file. (Though everyone got sudo on their Macbooks so we need to check if NinjaOne can restrict that)

## Sandbox mode

Sandbox mode is not designed for enforcing restrictions, there is no default "launch in sandbox mode" or "you can't run except in sandbox mode" feature.

Even when it comes to permissions `Filesystem and network restrictions are configured via Read, Edit, and WebFetch permission rules, not via these sandbox settings` which means besides the few flags like `excludedCommands`, `network.allowedDomains` or `network.allowUnixSockets` there is no such things as setting allow/deny rules outside of sandbox and setting allow/deny inside of sandbox.

```json
{
  "sandbox": {
    "enabled": true,
    "network": {
      "allowedDomains": ["github.com", "*.npmjs.org"]
    }
  },
  "allowedMcpServers": [{ "serverName": "github" }],
  "strictKnownMarketplaces": [
    {
      "source": "github",
      "repo": "your-org/approved-plugins"
    }
  ]
}
```

## MCP servers

We can enforce company-wide rules using `managed-mcp.json` that is pretty good. [How to control MCP access](https://code.claude.com/docs/en/mcp#managed-mcp-configuration)

## Permission gatchas

It is hard to figure out what exactly you need to deny permissions because of pattern matching

1.

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

So yeah the below approach worked but who knows if I could have further messed with it and bypass it

```sh
"Bash(**/.ssh*)",
"Bash(*ssh*)",
"Glob(**/.ssh/**)",
"Read(**/.ssh/**)",
"Edit(**/.ssh/**)",
"Write(**/.ssh/**)"
```

2.

These two are not the same (spaces matter)

```sh
"Bash(curl *)",
"Bash(curl*)",
```

## Remote AI agent

One approach we can take is to only allow developers to interact with a "remote agent" and deny local agent actions. This is going to need some testing to see how it works in practice because Claude Code does not natively support this.

The potential tools we have found so far are:

A. Custom MCP server (e.g. `claude mcp serve --port 8080 --tools "Bash,Read"`)
B. Claude Code on Web (though you can't limit claude code to ONLY use this and nothing else)

1. Use a `managed-settings.json` to limit local use
2. Prevent people messing with the managed settings (NinjaOne?)
3. Figure out how to set up permissions and connection
