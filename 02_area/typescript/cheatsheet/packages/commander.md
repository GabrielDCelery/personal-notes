# Commander

```sh
npm install commander
```

> The most widely used Node.js CLI framework. Used by webpack, eslint, and many others.

## Basic structure

```typescript
import { Command } from "commander";

const program = new Command();

program.name("myapp").description("A brief description").version("1.0.0");

program
  .command("serve")
  .description("Start the server")
  .option("-p, --port <number>", "port to listen on", "3000")
  .option("-v, --verbose", "enable verbose output")
  .action((options) => {
    const port = parseInt(options.port, 10);
    startServer(port, options.verbose);
  });

program.parse();
```

## Options (flags)

```typescript
program
  .command("deploy")
  .requiredOption("-e, --env <environment>", "target environment")
  .option("-d, --dry-run", "simulate without executing", false)
  .option("-t, --timeout <ms>", "timeout in milliseconds", "30000")
  .option("--no-cache", "disable caching")
  .action((options) => {
    // options.env     — string (required)
    // options.dryRun  — boolean
    // options.timeout — string
    // options.cache   — boolean (--no-cache sets to false)
  });
```

## Arguments (positional)

```typescript
program
  .command("get")
  .argument("<id>", "resource ID")
  .argument("[format]", "output format", "json")
  .action((id, format) => {
    console.log(`Getting ${id} as ${format}`);
  });

// Variadic arguments
program
  .command("delete")
  .argument("<ids...>", "IDs to delete")
  .action((ids: string[]) => {
    console.log(`Deleting: ${ids.join(", ")}`);
  });
```

## Subcommands

```typescript
const db = program.command("db").description("Database operations");

db.command("migrate")
  .description("Run database migrations")
  .action(() => runMigrations());

db.command("seed")
  .description("Seed database with test data")
  .option("--count <n>", "number of records", "100")
  .action((options) => seedDatabase(parseInt(options.count, 10)));

db.command("reset")
  .description("Reset database")
  .action(() => resetDatabase());

// Usage: myapp db migrate, myapp db seed --count 50
```

## Global options

```typescript
program
  .option("-c, --config <path>", "config file path", "./config.json")
  .option("--log-level <level>", "log level", "info");

// Access in subcommands via parent
program.command("serve").action((options, cmd) => {
  const globalOpts = cmd.parent?.opts();
  console.log(globalOpts?.config);
});
```

## Async actions

```typescript
program.command("deploy").action(async (options) => {
  try {
    await deploy(options);
  } catch (err) {
    console.error(err instanceof Error ? err.message : err);
    process.exit(1);
  }
});

// Commander handles async errors in Node 14+
program.parseAsync();
```

## Hooks

```typescript
program
  .command("serve")
  .hook("preAction", (thisCommand) => {
    console.log(`About to run: ${thisCommand.name()}`);
  })
  .action(() => startServer());
```

## Help customization

```typescript
program.addHelpText(
  "after",
  "\nExamples:\n  $ myapp serve --port 8080\n  $ myapp db migrate",
);

program.configureHelp({
  sortSubcommands: true,
  sortOptions: true,
});
```

## Typical project layout

```
myapp/
├── src/
│   ├── index.ts        # program.parse()
│   └── commands/
│       ├── serve.ts
│       ├── deploy.ts
│       └── db.ts
├── package.json        # "bin": { "myapp": "dist/index.js" }
└── tsconfig.json
```

```json
// package.json
{
  "bin": {
    "myapp": "dist/index.js"
  }
}
```

```typescript
// src/index.ts — first line
#!/usr/bin/env node
```
