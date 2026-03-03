---
title: CircleCI useful orbs
author: GaborZeller
---

# continuation: circleci/continuation@x.y.z

Useful when have a monorepo and want to run pipelines conditionally

```yml
orbs:
  continuation: circleci/continuation@1.1.0

workflows:
  # when creating a tag `release/*` run the release workflow
  release:
    jobs:
      - continuation/continue:
          configuration_path: .circleci/release.yml
          filters:
            branches:
              only:
                - /^rc\/developers\/.*$/
                - /^rc\/datascience\/.*$/
            tags:
              only:
                - /^release\/developers\/.*$/
                - /^release\/datascience\/.*$/
```

# path-filtering: circleci/path-filtering@x.y.z

Useful when want to set variables based on which part of the code changed

```yml
orbs:
  path-filtering: circleci/path-filtering@1.3.0

workflows:
  #---------------------------------------------------------------------------------------------------------
  # when pushing to `main` dependent on which content has changed run different workflows specified in development_test.yaml
  main-deployment:
    when:
      equal: [main, << pipeline.git.branch >>]
    jobs:
      - path-filtering/filter:
          base-revision: main
          config-path: .circleci/somethingsomethingpipeline.yml
          mapping: |
            services/developers/.* has_dev_code_changed true
            services/datascience/.* has_datascience_code_changed true
```
