---
title: "How to open browser in wsl"
date: 2025-10-31
tags: ["wsl", "windows", "browser"]
---

1. Install xdg-utils

Many Linux applications use the cli tool `xdg-open` to open URLs or files. `xdg-utils` installs that tool.

```sh
sudo apt install xdg-utils -y
```

2. Set BROWSER env variable

When `xdg-open` runs it checks the `BROWSER` env variable to determine how to open URLs. In WSL by pointing to `cmd.exe` we achieve that.

```sh
# Add to your shell rc file (e.g. zshrc)
export BROWSER="/mnt/c/Windows/System32/cmd.exe /c start"
```

So when `xdg-open https://example.com` runs it executes `/mnt/c/Windows/System32/cmd.exe /c start https://example.com`.

3. Edit /etc/wsl.conf and add:

```conf
[interop]
enabled=true
appendWindowsPath=true
```

- `enabled=true` - Enables WSL interoperability, which allows you to run Windows executables (like cmd.exe) directly from your WSL environment
- `appendWindowsPath=true` - Adds Windows PATH directories to your WSL PATH, making Windows executables more accessible (though the above solution uses the full path anyway)

4. Restart wsl

```powershell
wsl --shutdown
```
