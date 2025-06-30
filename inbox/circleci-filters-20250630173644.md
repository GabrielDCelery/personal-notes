---
title: CircleCI filters
author: GaborZeller
date: 2025-06-30T17-36-44Z
tags:
draft: true
---

# CircleCI filters

## Using anchor to create reusable filters

In yaml the `&` symbol is an anchor that can be used to reference reusable config later with `*`.

```yaml
workflows:
  deploy-dev-and-test:
    jobs:
      - test:
          filters: &build-filters
            branches:
              only:
                - master
            tags:
              ignore: /.*/
      - build:
          filters: *build-filters
      - publish:
          filters: *build-filters
          requires:
            - build
            - test
```
