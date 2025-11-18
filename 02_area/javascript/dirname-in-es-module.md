---
title: "Dirname in ES module"
date: 2025-10-28
tags: ["javascript", "es"]
---

# The issue

In CommonJS modules (require), `__dirname` is automatically available.

In ES modules (import/export), `__dirname` doesn't exist, so you need to construct it manually using this pattern.

```js
import { dirname } from "path";
import { fileURLToPath } from "url";

const __dirname = dirname(fileURLToPath(import.meta.url));
```

Step by step:

1. `import.meta.url` - Returns the file URL of the current module
   - Example: file:///Users/gabrielzeller/projects/homelab/custom-cdk/index.js

2. `fileURLToPath(import.meta.url)` - Converts the file URL to a file system path
   - Example: /Users/gabrielzeller/projects/homelab/custom-cdk/index.js
   - (from Node's url module)

3. `dirname(...)` - Extracts the directory path
   - Example: /Users/gabrielzeller/projects/homelab/custom-cdk
   - (from Node's path module)
