---
title: AWS X-Ray
author: Gabor Zeller
date: 2024-09-01T16:46:13Z
tags: ['aws']
draft: true
---

# AWS X-Ray

AWS X-Ray gives you the ability to have distributed tracing accross your services by tracking sessions through an applications. For example if you have a an application that is built of components like `API Gateway`, `Lambda`, `DynamoDB`, `Elastic Beanstalk` etc... then X-Ray has the ability to collect the session data from all of these services into one session flow for analysis.

#### How does it work?

At the start of a session the first service (usually the `API Gateway`) creates a unique `trace ID` that gets attached to the request in form of a `tracing header`.

All the services that are part of the session and have tracing enabled will send their data to `X-Ray` in the form of `segments`. These detail things like the `host`, `request`, `response`, `issues` that were found etc...

A segment can be made up of several `sub-segments` to provide more granularity.

All of this data is being used to create a `Service Graph` that is being used to determine how the application is structured.

All of this gets turned into a `Service Map` which is a visualization of the `Service Graph` with all the `trace details`.

#### How to enable it?

Tracing has to be enabled either at the service or application level. For applicatins that generally involves installing libraries and enabling at the `sdk` level.

Here are some examples of how to do it at a service level:

- EC2 - there is already an X-Ray agent installed
- ECS - an agent is already installed with every task
- Lambda - you have to enable it at the lambda's monitoring section
- Beanstalk - agent is already pre-installed
- API Gateway - has to be enabled at a per-stage basis
- SNS and SQS - can also be configured to forward data to X-Ray

It is important to note that for every service that you want to interact with X-Ray they need the appropriate `IAM persmissions`.
