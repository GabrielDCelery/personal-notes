---
title: Git CLI read repo access for organization
author: GaborZeller
date: 2025-12-11
tags: git
---

# The problem

How to give programmatic read access to applications to the organization's repo.

# GitHub App

## Create the GitHub app

1. Go to your organization settings:
   - Navigate to https://github.com/organizations/YOUR_ORG/settings/apps
   - Or: Organization → Settings → Developer settings → GitHub Apps → New GitHub App

2. Fill out the form:
   - GitHub App name: e.g., "Repository Reader Bot"
   - Homepage URL: Any valid URL (can be your org URL)
   - Webhook: Uncheck "Active" (not needed for read-only access)
   - Permissions → Repository permissions:
     - Contents: Read-only (allows reading code)
     - Metadata: Read-only (automatically selected)
   - Where can this GitHub App be installed?: Choose "Only on this account"

3. Click Create GitHub App

## Generate private key

1. After creation, scroll to Private keys section
2. Click Generate a private key
3. Save the downloaded .pem file securely

## Install the App

1. On the app's settings page, click Install App (left sidebar)
2. Click Install next to your organization
3. Choose:
   - All repositories or
   - Only select repositories (choose which ones)
4. Click Install
5. Note the Installation ID from the URL after install:
   - URL will be: https://github.com/organizations/YOUR_ORG/settings/installations/INSTALLATION_ID

## Generate tokens

We need the following environment variables:

Go to Org -> Settings -> GitHub Apps -> Configure -> App settings

```sh
APP_ID="1234567" # Go to Org -> Settings -> GitHub Apps -> Configure -> App settings
CLIENT_ID="Ab51pa1eao44saze6APA" # Go to Org -> Settings -> GitHub Apps -> Configure -> App settings
INSTALLATION_ID="12345678" # Go to Org -> Settings -> GitHub Apps -> Configure -> (you will see the ID in the URL)
PRIVATE_KEY_PATH="./thekeywegeneratedanddownloaded.pem" # The key we generated earlier
```

Use the following script to generate a token

```sh
#!/usr/bin/env bash

client_id=$CLIENT_ID # Client ID as first argument

pem=$( cat $PRIVATE_KEY_PATH ) # file path of the private key as second argument

now=$(date +%s)
iat=$((${now} - 60)) # Issues 60 seconds in the past
exp=$((${now} + 600)) # Expires 10 minutes in the future

b64enc() { openssl base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n'; }

header_json='{
    "typ":"JWT",
    "alg":"RS256"
}'
# Header encode
header=$( echo -n "${header_json}" | b64enc )

payload_json="{
    \"iat\":${iat},
    \"exp\":${exp},
    \"iss\":\"${client_id}\"
}"
# Payload encode
payload=$( echo -n "${payload_json}" | b64enc )

# Signature
header_payload="${header}"."${payload}"
signature=$(
    openssl dgst -sha256 -sign <(echo -n "${pem}") \
    <(echo -n "${header_payload}") | b64enc
)

# Create JWT
JWT="${header_payload}"."${signature}"
# printf '%s\n' "JWT: $JWT"

TOKEN=$(curl -s -X POST \
    -H "Authorization: Bearer $JWT" \
    -H "Accept: application/vnd.github+json" \
    "https://api.github.com/app/installations/$INSTALLATION_ID/access_tokens" \
 | jq -r '.token')

printf '%s\n' "TOKEN: $TOKEN"
```

## Use the token

Once we have the token just use it

```sh
git clone https://x-access-token:thetokenwegenerated@github.com/nameoftheorg/nameoftherepo.git
```

# Footnotes

[^1] [Authenticate as GitHub App](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app)
