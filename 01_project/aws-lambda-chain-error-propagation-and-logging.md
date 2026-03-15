---
title: AWS Lambda chain error propagation and logging
author: GaborZeller
date: 2026-03-15
tags:
---

# AWS Lambda chain error propagation and logging

When Lambdas are daisy-chained synchronously, each one catching and re-throwing an error will emit its own log entry, causing duplicate/cascading error noise in CloudWatch.

## Core principle

**Log only at the origin. Propagate without logging.**

Treat the Lambda chain like a call stack — you wouldn't log an exception at every stack frame.

## Pattern

```typescript
// Lambda C — origin of failure
export const handler = async (event) => {
  try {
    await doSomething();
  } catch (err) {
    logger.error("Failed in service C", { err, event }); // log ONCE here
    throw new ServiceError("C_FAILED", err.message, { retryable: false });
  }
};

// Lambda B — intermediate
export const handler = async (event) => {
  try {
    await invokeLambdaC(event);
  } catch (err) {
    // do NOT log — just re-throw
    throw err;
  }
};
```

## Structured error envelope

Use a typed error with an `origin` and `logged` flag so intermediate Lambdas know not to log:

```typescript
interface ChainError {
  code: string;
  message: string;
  origin: string;    // which Lambda first threw
  logged: boolean;   // has this been logged at origin?
  retryable: boolean;
}

// Intermediate Lambda
} catch (err) {
  const chainErr = parseChainError(err);
  if (!chainErr.logged) {
    logger.error('Unexpected unstructured error', { err });
  }
  throw chainErr;
}
```

## Correlation IDs

Propagate a `traceId` through every invocation payload so you can filter CloudWatch Logs Insights across the full chain:

```typescript
const payload = {
  ...event,
  _meta: {
    traceId: event._meta?.traceId ?? ulid(), // generate at entry, propagate downstream
    callDepth: (event._meta?.callDepth ?? 0) + 1,
  },
};
```

Include `traceId` in every log line.

## X-Ray (observability without code changes)

Enable active tracing on all Lambdas and wrap the Lambda client:

```typescript
import { captureAWSv3Client } from "aws-xray-sdk";
const lambdaClient = captureAWSv3Client(new LambdaClient({}));
```

X-Ray service map will show exactly which function in the chain failed.

## Summary

| Concern              | Approach                                                        |
| -------------------- | --------------------------------------------------------------- |
| Error origin logging | Log fully (with stack trace) only at the source                 |
| Intermediate logging | Don't log caught errors — only log unexpected/unstructured ones |
| Correlation          | Propagate a `traceId` through all payloads and log lines        |
| Observability        | Enable X-Ray tracing across all Lambdas in the chain            |
| Error structure      | Use a typed error envelope with `origin` and `logged` fields    |
