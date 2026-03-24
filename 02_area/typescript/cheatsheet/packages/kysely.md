# Kysely

```sh
npm install kysely pg
npm install -D @types/pg
```

> Type-safe SQL query builder for TypeScript. No ORM, no magic — types are inferred from your database schema interface. Full SQL control with autocompletion.

## Setup (PostgreSQL)

```typescript
import { Kysely, PostgresDialect } from "kysely";
import { Pool } from "pg";
import type { Database } from "./types";

const db = new Kysely<Database>({
  dialect: new PostgresDialect({
    pool: new Pool({ connectionString: process.env.DATABASE_URL }),
  }),
});

export default db;
```

## Schema Types

```typescript
// types.ts — define your database interface manually or generate it
import type { Generated, Insertable, Selectable, Updateable } from "kysely";

export interface UserTable {
  id: Generated<string>;
  name: string;
  email: string;
  role: "admin" | "user";
  active: boolean;
  created_at: Generated<Date>;
}

export interface PostTable {
  id: Generated<string>;
  title: string;
  content: string | null;
  author_id: string;
  created_at: Generated<Date>;
}

export interface Database {
  users: UserTable;
  posts: PostTable;
}

// Convenience types per table
export type User = Selectable<UserTable>;
export type NewUser = Insertable<UserTable>;
export type UserUpdate = Updateable<UserTable>;
```

## Queries — Select

```typescript
// Select all
const users = await db.selectFrom("users").selectAll().execute();

// Select columns
const names = await db.selectFrom("users").select(["name", "email"]).execute();

// Where
const user = await db
  .selectFrom("users")
  .selectAll()
  .where("id", "=", id)
  .executeTakeFirst();

// Multiple conditions
const activeAdmins = await db
  .selectFrom("users")
  .selectAll()
  .where("role", "=", "admin")
  .where("active", "=", true)
  .execute();

// Order and limit
const recent = await db
  .selectFrom("users")
  .selectAll()
  .orderBy("created_at", "desc")
  .limit(10)
  .offset(20)
  .execute();

// Like
const matched = await db
  .selectFrom("users")
  .selectAll()
  .where("name", "like", "%alice%")
  .execute();
```

## Queries — Insert

```typescript
// Single insert, return inserted row
const user = await db
  .insertInto("users")
  .values({
    name: "Alice",
    email: "alice@example.com",
    role: "user",
    active: true,
  })
  .returningAll()
  .executeTakeFirstOrThrow();

// Multiple insert
await db
  .insertInto("users")
  .values([
    { name: "Alice", email: "alice@example.com", role: "user", active: true },
    { name: "Bob", email: "bob@example.com", role: "user", active: true },
  ])
  .execute();

// Upsert (on conflict)
await db
  .insertInto("users")
  .values({
    name: "Alice",
    email: "alice@example.com",
    role: "user",
    active: true,
  })
  .onConflict((oc) => oc.column("email").doUpdateSet({ name: "Alice Updated" }))
  .execute();
```

## Queries — Update / Delete

```typescript
// Update
const updated = await db
  .updateTable("users")
  .set({ name: "Alice Smith", active: false })
  .where("id", "=", id)
  .returningAll()
  .executeTakeFirst();

// Delete
await db.deleteFrom("users").where("id", "=", id).execute();
```

## Joins

```typescript
// Inner join
const postsWithAuthors = await db
  .selectFrom("posts")
  .innerJoin("users", "users.id", "posts.author_id")
  .select(["posts.title", "users.name as author_name"])
  .execute();

// Left join
const usersWithPosts = await db
  .selectFrom("users")
  .leftJoin("posts", "posts.author_id", "users.id")
  .selectAll("users")
  .select("posts.title")
  .execute();
```

## Transactions

```typescript
const result = await db.transaction().execute(async (trx) => {
  const user = await trx
    .insertInto("users")
    .values({
      name: "Alice",
      email: "alice@example.com",
      role: "user",
      active: true,
    })
    .returningAll()
    .executeTakeFirstOrThrow();

  await trx
    .insertInto("posts")
    .values({ title: "First Post", author_id: user.id, content: null })
    .execute();

  return user;
});
```

## Raw SQL

```typescript
import { sql } from "kysely";

// Standalone query
const result =
  await sql`SELECT count(*) FROM users WHERE active = true`.execute(db);

// Inside a query
const users = await db
  .selectFrom("users")
  .selectAll()
  .where(sql`created_at > now() - interval '7 days'`)
  .execute();
```

## Migrations

```typescript
// migrations/001_create_users.ts
import type { Kysely } from "kysely";

export async function up(db: Kysely<any>): Promise<void> {
  await db.schema
    .createTable("users")
    .addColumn("id", "uuid", (col) =>
      col.primaryKey().defaultTo(sql`gen_random_uuid()`),
    )
    .addColumn("name", "text", (col) => col.notNull())
    .addColumn("email", "text", (col) => col.notNull().unique())
    .addColumn("role", "text", (col) => col.notNull().defaultTo("user"))
    .addColumn("active", "boolean", (col) => col.notNull().defaultTo(true))
    .addColumn("created_at", "timestamptz", (col) =>
      col.notNull().defaultTo(sql`now()`),
    )
    .execute();
}

export async function down(db: Kysely<any>): Promise<void> {
  await db.schema.dropTable("users").execute();
}
```

```typescript
// migrate.ts — run migrations programmatically
import { Migrator, FileMigrationProvider } from "kysely";
import { promises as fs } from "fs";
import path from "path";

const migrator = new Migrator({
  db,
  provider: new FileMigrationProvider({
    fs,
    path,
    migrationFolder: path.join(__dirname, "migrations"),
  }),
});

// Run all pending migrations
const { error, results } = await migrator.migrateToLatest();
results?.forEach((r) => {
  if (r.status === "Success") console.log(`migrated: ${r.migrationName}`);
  if (r.status === "Error") console.error(`failed: ${r.migrationName}`);
});
if (error) throw error;

// Rollback one
await migrator.migrateDown();
```

## Type Generation (from existing DB)

```sh
# kysely-codegen — generate types from a live database
npm install -D kysely-codegen
npx kysely-codegen --url $DATABASE_URL --out-file src/types/db.d.ts
```
