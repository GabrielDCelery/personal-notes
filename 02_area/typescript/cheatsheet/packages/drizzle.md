# Drizzle ORM

```sh
npm install drizzle-orm postgres    # or pg, mysql2, better-sqlite3
npm install -D drizzle-kit
```

> TypeScript ORM with SQL-like query builder. Schema defined in TypeScript, generates migrations. Thin layer over SQL — you stay close to the database.

## Setup (PostgreSQL)

```typescript
import { drizzle } from "drizzle-orm/postgres-js";
import postgres from "postgres";
import * as schema from "./schema";

const client = postgres(process.env.DATABASE_URL!);
export const db = drizzle(client, { schema });
```

## Schema Definition

```typescript
// schema.ts
import {
  pgTable,
  text,
  integer,
  timestamp,
  boolean,
  uuid,
} from "drizzle-orm/pg-core";

export const users = pgTable("users", {
  id: uuid("id").defaultRandom().primaryKey(),
  name: text("name").notNull(),
  email: text("email").notNull().unique(),
  role: text("role", { enum: ["admin", "user"] })
    .default("user")
    .notNull(),
  active: boolean("active").default(true).notNull(),
  createdAt: timestamp("created_at").defaultNow().notNull(),
});

export const posts = pgTable("posts", {
  id: uuid("id").defaultRandom().primaryKey(),
  title: text("title").notNull(),
  content: text("content"),
  authorId: uuid("author_id")
    .references(() => users.id)
    .notNull(),
  createdAt: timestamp("created_at").defaultNow().notNull(),
});
```

## Queries — Select

```typescript
import { eq, and, gt, like, desc, sql } from "drizzle-orm";

// Select all
const allUsers = await db.select().from(users);

// Where
const user = await db.select().from(users).where(eq(users.id, id));

// Multiple conditions
const activeAdmins = await db
  .select()
  .from(users)
  .where(and(eq(users.role, "admin"), eq(users.active, true)));

// Select specific columns
const names = await db
  .select({ name: users.name, email: users.email })
  .from(users);

// Order and limit
const recent = await db
  .select()
  .from(users)
  .orderBy(desc(users.createdAt))
  .limit(10)
  .offset(20);

// Like
const matched = await db
  .select()
  .from(users)
  .where(like(users.name, "%alice%"));
```

## Queries — Insert

```typescript
// Single insert
const [newUser] = await db
  .insert(users)
  .values({ name: "Alice", email: "alice@example.com" })
  .returning();

// Multiple insert
await db.insert(users).values([
  { name: "Alice", email: "alice@example.com" },
  { name: "Bob", email: "bob@example.com" },
]);

// On conflict (upsert)
await db
  .insert(users)
  .values({ name: "Alice", email: "alice@example.com" })
  .onConflictDoUpdate({
    target: users.email,
    set: { name: "Alice Updated" },
  });
```

## Queries — Update / Delete

```typescript
// Update
const [updated] = await db
  .update(users)
  .set({ name: "Alice Smith", active: false })
  .where(eq(users.id, id))
  .returning();

// Delete
await db.delete(users).where(eq(users.id, id));
```

## Joins

```typescript
// Inner join
const postsWithAuthors = await db
  .select({
    postTitle: posts.title,
    authorName: users.name,
  })
  .from(posts)
  .innerJoin(users, eq(posts.authorId, users.id));

// Left join
const usersWithPosts = await db
  .select()
  .from(users)
  .leftJoin(posts, eq(users.id, posts.authorId));
```

## Relational Queries (query API)

```typescript
// Define relations in schema
import { relations } from "drizzle-orm";

export const usersRelations = relations(users, ({ many }) => ({
  posts: many(posts),
}));

export const postsRelations = relations(posts, ({ one }) => ({
  author: one(users, { fields: [posts.authorId], references: [users.id] }),
}));

// Query with relations
const usersWithPosts = await db.query.users.findMany({
  with: { posts: true },
  where: eq(users.active, true),
  limit: 10,
});

const post = await db.query.posts.findFirst({
  with: { author: true },
  where: eq(posts.id, postId),
});
```

## Transactions

```typescript
const result = await db.transaction(async (tx) => {
  const [user] = await tx
    .insert(users)
    .values({ name: "Alice", email: "alice@example.com" })
    .returning();
  await tx.insert(posts).values({ title: "First Post", authorId: user.id });
  return user;
});
```

## Raw SQL

```typescript
import { sql } from "drizzle-orm";

const result = await db.execute(
  sql`SELECT count(*) FROM users WHERE active = true`,
);

// In a where clause
const users = await db
  .select()
  .from(users)
  .where(sql`${users.createdAt} > now() - interval '7 days'`);
```

## Migrations (drizzle-kit)

```typescript
// drizzle.config.ts
import { defineConfig } from "drizzle-kit";

export default defineConfig({
  schema: "./src/schema.ts",
  out: "./drizzle",
  dialect: "postgresql",
  dbCredentials: {
    url: process.env.DATABASE_URL!,
  },
});
```

```sh
npx drizzle-kit generate    # generate migration from schema changes
npx drizzle-kit migrate     # apply migrations
npx drizzle-kit push        # push schema directly (dev only)
npx drizzle-kit studio      # open visual editor
```

## Type Inference

```typescript
// Infer types from schema
type User = typeof users.$inferSelect; // select result type
type NewUser = typeof users.$inferInsert; // insert input type
```
