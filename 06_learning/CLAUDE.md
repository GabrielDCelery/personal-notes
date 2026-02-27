# AI-Assisted Learning - Lesson Generation Guide

This repository uses AI as a personal tutor to create focused, interview-ready lesson series on various technical topics.

## Project Philosophy

Unlike traditional courses, these lessons provide:

- Focused refreshers for experienced developers
- On-demand clarification and feedback
- Interview-critical concepts and gotchas
- Hands-on exercises with solutions

## Target Audience

**Senior developers (2+ years experience)** preparing for technical interviews who need:

- Quick refreshers on concepts they may have forgotten
- Deep dives into interview-critical topics
- Understanding of common edge cases and gotchas
- Practical challenges to test understanding

## Lesson Structure Template

Each lesson follows this structure:

### 1. Title + One-line Description

```markdown
# Lesson XX: [Topic Name]

[One sentence describing what this lesson covers]
```

### 2. Concept Sections

Each major concept includes:

#### a. Context/Pain Point (WHY it matters)

Start with a paragraph explaining the real-world problem or scenario where this concept matters. Help learners understand WHY they need to know this, not just WHAT it is.

**Example:**

```markdown
You find yourself writing the same type transformations over and over: "take this
interface but make all properties optional," or "create a type with these exact keys."
TypeScript's utility types are the standard library for type-level operations -
pre-built, battle-tested, and universally understood.
```

#### b. Quick Reference Tables

Use tables for comparisons, options, or quick lookups:

```markdown
| Type              | When to Use           | Installed   | Bundled            |
| ----------------- | --------------------- | ----------- | ------------------ |
| `dependencies`    | Runtime requirements  | Always      | Yes (if published) |
| `devDependencies` | Build/test tools only | Only in dev | No                 |
```

#### c. Code Examples with Annotations

- Use ✓ for correct examples
- Use ❌ for incorrect/problematic examples
- Show both good and bad patterns

````markdown
````json
{
  "dependencies": {
    "typescript": "^5.0.0" // ❌ Wrong - TypeScript is a build tool
  },
  "devDependencies": {
    "typescript": "^5.0.0" // ✓ Correct
  }
}
\```
````
````

#### d. Implementation Details (when relevant)

For TypeScript lessons, show how built-in types are implemented:

````markdown
**Implementation**:
\```typescript
type Partial<T> = {
[P in keyof T]?: T[P];
};
\```
````

#### e. Common Mistakes & Gotchas

Highlight edge cases and common pitfalls:

```markdown
### Common Mistakes

**Gotcha**: For `0.x.y` versions, `^` behaves like `~` because any minor change can be breaking.
```

### 3. Hands-On Exercises (2 per lesson)

Include 2 practical exercises with collapsible solutions:

````markdown
## Hands-On Exercise 1: [Exercise Name]

[Clear description of what to build/solve]

\```typescript
// Starter code or requirements
\```

<details>
<summary>Solution</summary>

**Issues** (if debugging exercise):

1. ❌ Problem 1
2. ⚠️ Warning 2

**Fixed** (if applicable):
\```typescript
// Solution code
\```

</details>
````

### 4. Interview Questions (4 per lesson)

Include 4 interview questions with:

- **Question text** as heading
- **Context paragraph** explaining WHY this question is asked (what it tests)
- **Collapsible answer**

```markdown
### Q1: What's the difference between ^ and ~ in package.json?

Interviewers ask this to see if you understand package architecture and have published
libraries (not just apps). It reveals whether you know the difference between building
something standalone versus building something that plugs into an ecosystem.

<details>
<summary>Answer</summary>

- `^1.4.2`: Allows updates that don't change the leftmost non-zero digit
  - Allows: 1.4.3, 1.5.0, 1.99.0
  - Blocks: 2.0.0

**Exception**: For 0.x.y versions...

</details>
```

### 5. Key Takeaways

Numbered list (7-10 items) summarizing the most important points:

```markdown
## Key Takeaways

1. **Dependencies**: Runtime vs dev vs peer - know the difference
2. **Semver**: `^` for minor updates, `~` for patches, exact for critical deps
3. **Exports**: Modern way to define package entry points (supports dual ESM/CJS)
   ...
```

### 6. Next Steps

Preview what's coming in the next lesson:

```markdown
## Next Steps

In [Lesson 02: tsconfig.json Mastery](lesson-02-tsconfig-mastery.md), you'll learn:

- Critical compiler options and their implications
- Module resolution strategies
- Project references for monorepos
```

## Formatting Conventions

### Markdown Features

- Use `<details>` tags for collapsible content (solutions, answers)
- Use fenced code blocks with language identifiers
- Use tables for comparisons and reference data
- Use bold for emphasis on key terms
- Use inline code for technical terms, file names, commands

### Code Annotations

- `✓` - Correct approach
- `❌` - Wrong/problematic approach
- `⚠️` - Warning/edge case

### Emphasis Style

- **Bold** for key concepts and field names
- `Code font` for technical terms, values, types
- _Italic_ sparingly (prefer bold for emphasis)

## Directory Structure

```
[topic]/
├── README.md              # Overview and lesson index
├── lesson-01-[name].md    # First lesson
├── lesson-02-[name].md    # Second lesson
├── ...
└── practice/              # Optional: practice projects/exercises
```

## Lesson Series Organization

### Typical Flow (6-8 lessons)

1. **Fundamentals/Setup** - Core concepts, configuration
2. **Core Concepts** - Primary patterns and principles
3. **Advanced Topics** - Deep dives into complex areas
4. **Practical Application** - Real-world usage patterns
5. **Performance/Optimization** - Scaling and best practices
6. **Publishing/Distribution** - Sharing and deployment (if applicable)

### Example: TypeScript Series

1. package.json Deep Dive (setup/tooling)
2. tsconfig.json Mastery (configuration)
3. Type System Internals (fundamentals)
4. Utility Types & Manipulation (advanced)
5. Generics Deep Dive (advanced)
6. Module Systems (practical)
7. Declaration Files (practical)
8. Publishing npm Packages (distribution)

## Content Guidelines

### Tone

- Professional and direct
- Focus on practical application
- Assume reader has experience but may need refreshers
- Explain WHY, not just WHAT

### Depth Level

- Skip absolute basics (readers are senior devs)
- Focus on intermediate-to-advanced concepts
- Cover edge cases and gotchas
- Include real-world scenarios

### Interview Focus

- Highlight commonly asked questions
- Explain what each question tests
- Show thought process, not just answers
- Cover both theory and practical application

## Examples from Existing Lessons

### Good Context Paragraph (explains WHY)

> "Here's the nightmare: Node.js uses CommonJS. Browsers use ESM. Bundlers want ESM for tree-shaking. TypeScript needs `.d.ts` files. Your library needs to work everywhere. The old solution was to pick one format and let tools figure it out. The modern solution is to provide EVERYTHING..."

### Good Table Structure

Clear, scannable, covers the essential comparisons:

```markdown
| Type         | .js files are | .mjs | .cjs     |
| ------------ | ------------- | ---- | -------- |
| `"module"`   | ESM           | ESM  | CommonJS |
| `"commonjs"` | CommonJS      | ESM  | CommonJS |
```

### Good Exercise Format

Clear requirements, starter code, complete solution with explanations.

## AI Instructions for Lesson Generation

When generating lessons:

1. **Read existing lessons first** to match tone and structure
2. **Use the template** consistently across all lessons
3. **Start with WHY** before explaining HOW
4. **Include practical examples** from real-world scenarios
5. **Test understanding** through exercises that require application, not just recall
6. **Connect lessons** with "Next Steps" previews
7. **Focus on interview relevance** - what gets asked and why

## Maintenance Notes

- Keep lessons focused (each should be completable in 30-60 minutes)
- Update examples when technologies change
- Add new lessons as topics evolve
- Maintain consistent numbering and cross-references
