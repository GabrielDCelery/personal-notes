---
title: Typescript as const
author: GaborZeller
date: 2025-03-04T12-13-34Z
tags:
draft: true
---

# Typescript as const

## What does as const do?

```typescript
const journeyHelperMap = {
  [JourneySource.APP]: JourneyAppDDBFactModelHelper,
  [JourneySource.DEVICE]: JourneyDeviceDDBFactModelHelper,
  [JourneySource.VENDOR]: JourneyVendorDDBFactModelHelper,
} as const;
```

`as const` is a type assertion in TypeScript that makes the object literal become a deeply immutable (read-only) type with its most specific literal type possible. Here's what it does specifically:

1. Makes all properties read-only (cannot be modified after creation)
2. Converts the values to literal types instead of wider types
3. Creates a deeply immutable object type

Here is an example:

WITHOUT `as const`:

```typescript
const journeyModelHelperMap = {
    [JourneySource.APP]: JourneyAppDDBFactModelHelper
};

// TypeScript infers this as:
{
    [key: string]: typeof JourneyAppDDBFactModelHelper
}
// The properties are mutable
// The values are inferred as their general type
```

WITH `as const`:

```typescript
const journeyModelHelperMap = {
    [JourneySource.APP]: JourneyAppDDBFactModelHelper
} as const;

// TypeScript infers this as:
{
    readonly [JourneySource.APP]: typeof JourneyAppDDBFactModelHelper
}
// The properties are read-only
// The values are treated as specific literal types
```

In your specific code, `as const` ensures:

1. The mapping between journey sources and their helpers cannot be modified after creation
2. TypeScript knows exactly which journey sources are valid keys
3. Better type checking when accessing the helpers through the map
4. The compiler can make better type inferences when using this object

This is particularly useful in this case because it provides better type safety when accessing the helper classes based on the journey source, and prevents accidental modifications to the mapping.
