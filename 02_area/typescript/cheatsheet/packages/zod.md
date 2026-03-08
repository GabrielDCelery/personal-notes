# Zod

```sh
npm install zod
```

> TypeScript-first schema validation. Define a schema once, get runtime validation and static types.

## Basics

```typescript
import { z } from "zod";

const UserSchema = z.object({
  name: z.string().min(1),
  email: z.string().email(),
  age: z.number().int().positive().optional(),
});

// Infer TypeScript type from schema
type User = z.infer<typeof UserSchema>;
// { name: string; email: string; age?: number }

// Validate
const user = UserSchema.parse(data); // throws ZodError on failure
const result = UserSchema.safeParse(data); // returns { success, data?, error? }

if (result.success) {
  console.log(result.data.name); // typed
} else {
  console.log(result.error.issues);
}
```

## Primitive Types

```typescript
z.string();
z.number();
z.boolean();
z.bigint();
z.date();
z.undefined();
z.null();
z.any();
z.unknown();
z.literal("admin"); // exact value
z.enum(["admin", "user", "guest"]);
z.nativeEnum(MyEnum); // from TS enum
```

## String Validations

```typescript
z.string().min(1); // non-empty
z.string().max(255);
z.string().email();
z.string().url();
z.string().uuid();
z.string().regex(/^[a-z]+$/);
z.string().startsWith("https://");
z.string().trim(); // transform: strip whitespace
z.string().toLowerCase(); // transform: to lowercase
```

## Number Validations

```typescript
z.number().int();
z.number().positive();
z.number().nonnegative();
z.number().min(1).max(100);
z.number().finite();
z.coerce.number(); // coerce string "42" → 42
```

## Objects

```typescript
const UserSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  email: z.string().email(),
  role: z.enum(["admin", "user"]),
  metadata: z.record(z.string()), // { [key: string]: string }
});

// Partial (all optional)
const UpdateSchema = UserSchema.partial();

// Pick / Omit
const CreateSchema = UserSchema.omit({ id: true });
const SummarySchema = UserSchema.pick({ id: true, name: true });

// Extend
const AdminSchema = UserSchema.extend({
  permissions: z.array(z.string()),
});

// Merge
const FullSchema = BaseSchema.merge(ExtraSchema);

// Strict — reject unknown keys
const StrictSchema = UserSchema.strict();

// Passthrough — allow unknown keys
const LooseSchema = UserSchema.passthrough();
```

## Arrays & Tuples

```typescript
z.array(z.string()); // string[]
z.array(z.string()).nonempty(); // [string, ...string[]]
z.array(z.number()).min(1).max(10);

z.tuple([z.string(), z.number()]); // [string, number]
```

## Unions & Discriminated Unions

```typescript
// Simple union
z.union([z.string(), z.number()]);
// shorthand: z.string().or(z.number())

// Discriminated union (faster, better errors)
const EventSchema = z.discriminatedUnion("type", [
  z.object({ type: z.literal("click"), x: z.number(), y: z.number() }),
  z.object({ type: z.literal("scroll"), offset: z.number() }),
  z.object({ type: z.literal("keypress"), key: z.string() }),
]);
```

## Transforms

```typescript
// Transform after validation
const schema = z.string().transform((val) => val.toUpperCase());
schema.parse("hello"); // "HELLO"

// Coerce from string
const PortSchema = z.coerce.number().int().min(1).max(65535);
PortSchema.parse("3000"); // 3000

// Default values
z.string().default("anonymous");
z.number().default(0);

// Preprocess (before validation)
z.preprocess((val) => String(val), z.string());
```

## Custom Validation

```typescript
const PasswordSchema = z
  .string()
  .min(8)
  .refine((val) => /[A-Z]/.test(val), "Must contain uppercase")
  .refine((val) => /[0-9]/.test(val), "Must contain a number");

// superRefine for complex logic
const SignupSchema = z
  .object({
    password: z.string().min(8),
    confirmPassword: z.string(),
  })
  .superRefine((data, ctx) => {
    if (data.password !== data.confirmPassword) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Passwords don't match",
        path: ["confirmPassword"],
      });
    }
  });
```

## Error Handling

```typescript
try {
  UserSchema.parse(data);
} catch (err) {
  if (err instanceof z.ZodError) {
    console.log(err.issues);
    // [{ code: "too_small", minimum: 1, path: ["name"], message: "..." }]

    console.log(err.flatten());
    // { formErrors: [], fieldErrors: { name: ["..."], email: ["..."] } }
  }
}
```

## Common Patterns

### API request validation

```typescript
const CreateUserSchema = z.object({
  name: z.string().min(1).max(100),
  email: z.string().email(),
  role: z.enum(["admin", "user"]).default("user"),
});

type CreateUserInput = z.infer<typeof CreateUserSchema>;

// In route handler
const input = CreateUserSchema.parse(req.body); // throws 400-worthy error
```

### Environment variables

```typescript
const EnvSchema = z.object({
  PORT: z.coerce.number().default(3000),
  DATABASE_URL: z.string().url(),
  NODE_ENV: z
    .enum(["development", "production", "test"])
    .default("development"),
});

export const env = EnvSchema.parse(process.env);
```

### API response typing

```typescript
const ApiResponseSchema = z.object({
  data: z.array(UserSchema),
  total: z.number(),
  page: z.number(),
});

const response = await fetch("/api/users");
const body = ApiResponseSchema.parse(await response.json());
// body is fully typed
```
