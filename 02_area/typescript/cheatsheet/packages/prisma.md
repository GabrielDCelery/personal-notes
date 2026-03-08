# Prisma

```sh
npm install prisma -D
npm install @prisma/client
npx prisma init
```

> Type-safe ORM with declarative schema, auto-generated client, and migrations. Schema lives in `.prisma` files, not TypeScript.

## Setup

```prisma
// prisma/schema.prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id        String   @id @default(uuid())
  name      String
  email     String   @unique
  role      Role     @default(USER)
  active    Boolean  @default(true)
  posts     Post[]
  createdAt DateTime @default(now()) @map("created_at")

  @@map("users")
}

model Post {
  id        String   @id @default(uuid())
  title     String
  content   String?
  author    User     @relation(fields: [authorId], references: [id])
  authorId  String   @map("author_id")
  createdAt DateTime @default(now()) @map("created_at")

  @@map("posts")
}

enum Role {
  ADMIN
  USER
}
```

```typescript
// db.ts
import { PrismaClient } from "@prisma/client";

export const prisma = new PrismaClient();
```

## Queries — Find

```typescript
// Find many
const users = await prisma.user.findMany();

// Where
const admins = await prisma.user.findMany({
  where: { role: "ADMIN", active: true },
});

// Find unique (by unique field)
const user = await prisma.user.findUnique({
  where: { email: "alice@example.com" },
});

// Find first
const user = await prisma.user.findFirst({
  where: { name: { contains: "Alice" } },
});

// Select specific fields
const names = await prisma.user.findMany({
  select: { name: true, email: true },
});

// Order, limit, offset
const recent = await prisma.user.findMany({
  orderBy: { createdAt: "desc" },
  take: 10,
  skip: 20,
});
```

## Queries — Create

```typescript
// Single create
const user = await prisma.user.create({
  data: { name: "Alice", email: "alice@example.com" },
});

// Create many
await prisma.user.createMany({
  data: [
    { name: "Alice", email: "alice@example.com" },
    { name: "Bob", email: "bob@example.com" },
  ],
});

// Upsert
const user = await prisma.user.upsert({
  where: { email: "alice@example.com" },
  update: { name: "Alice Updated" },
  create: { name: "Alice", email: "alice@example.com" },
});
```

## Queries — Update / Delete

```typescript
// Update
const user = await prisma.user.update({
  where: { id },
  data: { name: "Alice Smith", active: false },
});

// Update many
await prisma.user.updateMany({
  where: { active: false },
  data: { role: "USER" },
});

// Delete
await prisma.user.delete({ where: { id } });
```

## Relations (include)

```typescript
// Include related records
const userWithPosts = await prisma.user.findUnique({
  where: { id },
  include: { posts: true },
});

// Nested include
const post = await prisma.post.findUnique({
  where: { id: postId },
  include: { author: { select: { name: true, email: true } } },
});

// Filter on relations
const usersWithPosts = await prisma.user.findMany({
  where: { posts: { some: { title: { contains: "TypeScript" } } } },
  include: { posts: true },
});
```

## Transactions

```typescript
// Interactive transaction
const result = await prisma.$transaction(async (tx) => {
  const user = await tx.user.create({
    data: { name: "Alice", email: "alice@example.com" },
  });
  await tx.post.create({
    data: { title: "First Post", authorId: user.id },
  });
  return user;
});

// Batch transaction (all or nothing)
await prisma.$transaction([
  prisma.user.create({ data: { name: "Alice", email: "alice@example.com" } }),
  prisma.user.create({ data: { name: "Bob", email: "bob@example.com" } }),
]);
```

## Raw SQL

```typescript
const users =
  await prisma.$queryRaw`SELECT * FROM users WHERE active = ${true}`;

await prisma.$executeRaw`UPDATE users SET active = false WHERE id = ${id}`;
```

## Migrations

```sh
npx prisma migrate dev --name init       # create + apply migration (dev)
npx prisma migrate deploy                # apply pending migrations (prod)
npx prisma migrate reset                 # reset database (dev)
npx prisma db push                       # push schema without migration (prototyping)
npx prisma generate                      # regenerate client after schema change
npx prisma studio                        # visual editor
```

## Prisma vs Drizzle

|                      | Prisma                | Drizzle               |
| -------------------- | --------------------- | --------------------- |
| Schema               | `.prisma` DSL         | TypeScript            |
| Query style          | object-based          | SQL-like builder      |
| Generated client     | yes (code generation) | no (schema = types)   |
| Migrations           | prisma migrate        | drizzle-kit           |
| Raw SQL escape hatch | `$queryRaw`           | `sql` template tag    |
| Performance          | good                  | thinner layer, faster |
| Learning curve       | low                   | low (if you know SQL) |
