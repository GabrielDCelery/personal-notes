---
title: Go workspace management
author: GaborZeller
---

```sh
touch go.work # Creates the workspace file
find . -name go.mod -exec dirname {} \; | xargs -I {} go work use {} # populates the go.work file with the workspace directory locations
find . -name go.mod -exec dirname {} \; | xargs -n1 -I {} sh -c 'cd {} && go get ./...' # install all the dependencies that are used across the project
```
