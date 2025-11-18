---
title: "Manage multiple environments using mise"
date: 2025-10-21
tags: ["mise"]
---

# 1. Use dedicated .env files

Create separate .env files for each environment and use a `MISE_ENV` variable to switch between environments

```txt
project/
├── .mise.toml
├── .env.dev
├── .env.test
├── .env.prod
```

In the `mise.toml` file load the env file using the `MISE_ENV` variable.

```mise.toml
[env]
 # Load based on MISE_ENV variable
 _.file = [".env.{{ env.MISE_ENV | default(value='dev') }}"]
```

Have different `.env.*` files

```.env.dev
AWS_PROFILE="dev-profile"
```

```.env.test
AWS_PROFILE="test-profile"
```

Example usage with the CLI

```sh
# Development (default)
mise exec -- terraform plan

# Test
MISE_ENV=test mise exec -- terraform plan

# Production
MISE_ENV=prod mise exec -- terraform plan
```

# 2. Use different mise profiles

Load the different environments using mise profiles.

```txt
project/
├── .mise.toml              # Base config
├── .mise.dev.toml          # Dev overrides
├── .mise.test.toml         # Test overrides
├── .mise.prod.toml         # Prod overrides
```

Have a base `mise.toml` file.

```mise.toml
[tools]
terraform = "1.9.0"
node = "22"

[env]
TF_LOG = "INFO"
```

Have environment specific mise configurations

```mise.dev.toml
[env]
AWS_PROFILE = "dev-profile"
TF_VAR_environment = "dev"
AWS_REGION = "us-east-1"
```

```mise.prod.toml
[env]
AWS_PROFILE = "prod-profile"
TF_VAR_environment = "prod"
AWS_REGION = "us-west-2"
```

Example usage with the CLI

```sh
export MISE_PROFILE=dev
mise exec -- terraform plan
```

# 3. Use tasks

Have tasks that load the appropriate profiles

```mise.toml
[tools]
terraform = "1.9.0"

[tasks.dev]
run = "terraform plan"
env = { AWS_PROFILE = "dev-profile", TF_VAR_environment = "dev" }

[tasks.test]
run = "terraform plan"
env = { AWS_PROFILE = "test-profile", TF_VAR_environment = "test" }

[tasks.prod]
run = "terraform plan"
env = { AWS_PROFILE = "prod-profile", TF_VAR_environment = "prod" }

# More complex tasks
[tasks."deploy:dev"]
run = '''
#!/bin/bash
terraform init
terraform plan
terraform apply -auto-approve
'''
env = { AWS_PROFILE = "dev-profile", TF_VAR_environment = "dev" }

[tasks."deploy:prod"]
run = '''
#!/bin/bash
terraform init
terraform plan
read -p "Apply to PRODUCTION? (yes/no) " confirm
[[ "$confirm" == "yes" ]] && terraform apply
'''
env = { AWS_PROFILE = "prod-profile", TF_VAR_environment = "prod" }
```

Example usage with the CLI

```sh
mise run dev
mise run prod
mise run deploy:dev
mise run deploy:prod
```

# Reference

[^1] [Mise Config environments](https://mise.jdx.dev/configuration/environments.html)
[^2] [Mise Environments](https://mise.jdx.dev/environments/#env-file)
