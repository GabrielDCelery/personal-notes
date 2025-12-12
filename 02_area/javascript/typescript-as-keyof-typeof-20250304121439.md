---
title: Typescript as keyof typeof
author: GaborZeller
date: 2025-03-04T12-14-39Z
tags:
draft: true
---

# Typescript as keyof typeof

## What does as keyof typeof do?

```typescript
const ModelHelper =
  speedEventModelHelperMap[
    speedEventSeverity as keyof typeof speedEventModelHelperMap
  ];
```

`as keyof typeof speedEventModelHelperMap` is a type assertion that helps TypeScript understand exactly what keys are valid for the object. Let's break it down step by step:

1. `typeof speedEventModelHelperMap` gets the type of the object, which in this case is:

```typescript
{
    readonly Cancellation: typeof DrivingSpeedingCancellationDDBFactModelHelper;
    readonly Scored: typeof DrivingSpeedingScoredDDBFactModelHelper;
}
```

2. `keyof` gets all possible key types from that object type, which would be:

```typescript
type Keys = "Cancellation" | "Scored";
```

3. So `speedEventSeverity as keyof typeof speedEventModelHelperMap` tells TypeScript that `speedEventSeverity` should be treated as one of these specific string literals: either "Cancellation" or "Scored"

Without this type assertion:

```typescript
// This would give a type error because speedEventSeverity is just a string
// and could be any string value
const ModelHelper = speedEventModelHelperMap[speedEventSeverity];
```

With the type assertion:

```typescript
// TypeScript now knows speedEventSeverity can only be "Cancellation" or "Scored"
const ModelHelper =
  speedEventModelHelperMap[
    speedEventSeverity as keyof typeof speedEventModelHelperMap
  ];
```

This is particularly useful because: 2. It helps catch errors if you try to access a key that doesn't exist 3. You get better IDE support with autocompletion 4. If you add new entries to `speedEventModelHelperMap`, TypeScript will automatically include them in the valid keys

It's worth noting that in your code, this type assertion is necessary because `speedEventSeverity` comes in as a general `string` type, but we need to tell TypeScript that it's specifically one of the keys in our map.
