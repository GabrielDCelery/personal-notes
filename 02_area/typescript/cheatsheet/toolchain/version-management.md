# Node.js Version Management

## Quick Reference

| Method                     | Best for                            |
| -------------------------- | ----------------------------------- |
| `nvm`                      | Most popular, bash-based            |
| `fnm`                      | Fast (Rust-based), cross-platform   |
| `mise`                     | Polyglot (Node, Go, Python, etc.)   |
| `volta`                    | Auto-switches, pins in package.json |
| `.node-version` / `.nvmrc` | Per-project version pinning         |

## .node-version / .nvmrc

### 1. Pin Node version per project

```sh
# .node-version (supported by fnm, mise, nvm, volta)
20.11.0

# .nvmrc (nvm-specific, also supported by fnm)
20.11.0
```

Most tools auto-switch when you `cd` into a directory with these files.

### 2. engines in package.json

```json
{
  "engines": {
    "node": ">=20"
  }
}
```

```ini
# .npmrc — make engines a hard requirement
engine-strict=true
```

## nvm

### 3. Install and use

```sh
# Install nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash

# Install a version
nvm install 20
nvm install 22

# Switch
nvm use 20
nvm use 22

# Set default
nvm alias default 20

# Use project's .nvmrc
nvm use
```

### 4. Auto-switch on cd (add to .zshrc)

```sh
autoload -U add-zsh-hook
load-nvmrc() {
  if [[ -f .nvmrc ]]; then
    nvm use
  fi
}
add-zsh-hook chpwd load-nvmrc
load-nvmrc
```

## fnm (recommended — fast)

### 5. Install and use

```sh
# Install fnm
curl -fsSL https://fnm.vercel.app/install | bash
# or
brew install fnm

# Install a version
fnm install 20
fnm install 22

# Switch
fnm use 20

# Set default
fnm default 20

# Auto-switch (add to .zshrc)
eval "$(fnm env --use-on-cd)"
```

## mise

### 6. Install and use

```sh
# Install mise
curl https://mise.run | sh

# Install Node
mise use node@20
mise use node@22

# Pin per project
mise use --pin node@20.11.0   # creates .mise.toml

# .mise.toml
[tools]
node = "20.11.0"
```

mise also manages Go, Python, Terraform, etc. — single tool for everything.

## volta

### 7. Install and use

```sh
# Install volta
curl https://get.volta.sh | bash

# Install and pin to package.json
volta install node@20
volta pin node@20

# package.json gets:
# "volta": { "node": "20.11.0" }
```

Volta auto-switches based on package.json — no `.node-version` file needed.

## CI Version Pinning

### 8. GitHub Actions

```yaml
- uses: actions/setup-node@v4
  with:
    node-version-file: ".node-version" # or .nvmrc
    cache: "npm"

# Or pin explicitly
- uses: actions/setup-node@v4
  with:
    node-version: "20"
    cache: "npm"
```

### 9. GitLab CI

```yaml
image: node:20-slim

build:
  script:
    - npm ci
    - npm run build
```

### 10. Docker

```dockerfile
# Pin in Dockerfile
FROM node:20.11.0-slim
```

Always pin to a specific version in production Dockerfiles.
