# CPU and Program Execution

How does a CPU actually run your code? Every language — Go, Java, TypeScript, Python, Rust — eventually becomes instructions that a CPU executes one (or more) at a time. Understanding this layer changes how you think about performance.

## What a CPU Actually Does

A CPU is a machine that:

1. **Fetches** an instruction from memory
2. **Decodes** it (figures out what to do)
3. **Executes** it (does the thing)
4. **Writes back** the result
5. Moves to the next instruction

That's it. Every program you've ever written reduces to this loop running billions of times per second.

## Registers — The Fastest Storage That Exists

Before we talk about memory (RAM, caches), there's something even faster: **registers**. These are tiny storage slots built directly into the CPU.

| Storage  | Size                              | Access time           |
| -------- | --------------------------------- | --------------------- |
| Register | 64 bits each, ~16 general purpose | ~0.3 ns (1 cycle)     |
| L1 cache | 32-64 KB                          | ~1 ns (3-4 cycles)    |
| RAM      | 8-64 GB                           | ~100 ns (300+ cycles) |

A register access is **300x faster than RAM**. This matters.

When you write `x := a + b` in Go, the compiled code does something like:

```
LOAD  a  → register R1      # get a from memory into a register
LOAD  b  → register R2      # get b from memory into a register
ADD   R1, R2 → R3           # add them (operates ONLY on registers)
STORE R3 → x                # write result back to memory
```

The CPU **cannot** add two numbers that are in RAM. It must load them into registers first, operate, then store back. Every operation works this way.

### Why this matters for programming

The compiler's job is to keep frequently-used values in registers and avoid unnecessary loads/stores to memory. This is called **register allocation** and it's one of the biggest reasons compiled languages (Go, Rust, C) are faster than interpreted ones (Python, Ruby) — a good compiler keeps hot values in registers, while an interpreter is constantly bouncing values between memory and registers.

### Registers are scarce — only ~16 per core

A typical x86-64 CPU has **16 general-purpose registers**. That's it. The actual names are historical baggage:

```
RAX, RBX, RCX, RDX, RSI, RDI, RBP, RSP, R8, R9, R10, R11, R12, R13, R14, R15
```

| Register                   | Conventional role                                               |
| -------------------------- | --------------------------------------------------------------- |
| RSP                        | Stack pointer — points to the top of the current stack          |
| RBP                        | Base pointer — sometimes marks the start of a stack frame       |
| RAX                        | Often holds return values                                       |
| RDI, RSI, RDX, RCX, R8, R9 | Function arguments (Linux — first 6 args go in these registers) |
| The rest                   | Truly general purpose — compiler assigns freely                 |

ARM CPUs (Apple Silicon, AWS Graviton) have a cleaner naming: `X0` through `X30` — 31 general-purpose registers, no weird historical names. Advantage of designing an instruction set from scratch.

When the compiler runs out of registers, it **spills** values to the stack (RAM). But in practice this rarely matters because most variables aren't alive at the same instant:

```go
x := compute()    // x is live from here...
y := other()
result := x + y   // ...to here. x is dead after this line.
z := more()       // x's register can now hold z
```

You can see the real assembly for any Go function:

```sh
go tool compile -S main.go
# or
go build -gcflags="-S" ./...
```

### Multiple cores and hyperthreading

Each core has its **own complete set of 16 registers**. Four cores = four threads truly executing simultaneously, each with their own registers. They don't share.

**Hyperthreading** (Intel) / **SMT** (AMD) makes a single physical core appear as two logical cores. Registers are **duplicated**, but execution units and caches are **shared**.

| Component             | Shared or duplicated?                           |
| --------------------- | ----------------------------------------------- |
| Registers             | Duplicated — each logical core gets its own set |
| Execution units (ALU) | Shared                                          |
| L1/L2 cache           | Shared                                          |
| Pipeline              | Shared                                          |

So a 4-core hyperthreaded CPU has 8 logical cores, each with their own 16 registers = 128 registers total, 8 threads truly "in flight."

The trick: one thread rarely uses all execution units at once. When thread A is waiting on a memory load, thread B can use the ALU. They fill gaps in each other's pipeline:

```
              Cycle 1    Cycle 2    Cycle 3    Cycle 4
Thread A:     [ADD]      [waiting for memory...]
Thread B:     [idle]     [ADD]      [MUL]      [waiting...]
Thread A:                                      [LOAD done, ADD]
```

This is why hyperthreading gives ~15-30% more throughput, not 100%. It's not a second core — it's better utilisation of one core's idle time.

This is also why Go's `runtime.GOMAXPROCS` defaults to the number of logical cores. On a 4-core hyperthreaded machine, `runtime.NumCPU()` returns 8. Go creates 8 P's (processors), running up to 8 goroutines truly in parallel. Your million goroutines get multiplexed onto those 8 — but at any instant, only 8 are executing on hardware.

### Context switches — how the OS juggles threads

When the OS switches between threads, it doesn't understand your code. It saves **all registers blindly**:

1. Timer interrupt fires (hardware)
2. CPU traps into the OS kernel
3. Kernel saves ALL 16 registers from CPU → Thread A's context struct in RAM
4. Kernel loads ALL 16 registers from Thread B's context struct in RAM → CPU
5. Kernel returns to user space
6. Thread B's code resumes — its registers are exactly as it left them

The OS treats registers as an opaque blob. Save 16 values, restore 16 values. The **compiled code** gives those registers meaning — Thread A's compiler decided R3 holds `total`, Thread B's compiler decided R3 holds `connectionCount`. The OS doesn't know or care.

### Where variables actually live

When your compiled program is loaded, the **code** (instructions) is mapped into memory. But variables end up in one of three places:

| Location      | What goes there                                   | Lifetime                 |
| ------------- | ------------------------------------------------- | ------------------------ |
| **Registers** | Small, hot variables the compiler optimises for   | Duration of the function |
| **Stack**     | Local variables, function arguments               | Duration of the function |
| **Heap**      | Data that escapes the function — pointers, shared | Until garbage collected  |

```go
func add() int {
    x := 3
    y := 4
    return x + y
}
// x and y probably never touch RAM — they live in registers.
// The compiler might even constant-fold this to: return 7
```

```go
func newUser() *User {
    u := User{Name: "Alice"}
    return &u  // pointer escapes the function
}
// u MUST go on the heap — the caller needs it after this function's
// stack frame is gone. Go's compiler decides this via escape analysis.
```

You can see these decisions:

```sh
go build -gcflags="-m" ./...
# output: "moved to heap: u"
```

The compiler's goal: keep as much as possible in registers and on the stack. Heap allocation is the most expensive option — it requires runtime allocation and creates work for the garbage collector.

### Registers vs caches

Registers and caches are both fast storage on the CPU, but they serve different roles:

|                        | Registers                                               | Caches                                                         |
| ---------------------- | ------------------------------------------------------- | -------------------------------------------------------------- |
| Addressed by           | **Name** — `RAX`, `RBX` (baked into the instruction)    | **Memory address** — cache is invisible, same addresses as RAM |
| Who controls placement | **Compiler** — picks which register in the machine code | **Hardware** — automatically caches recently-used memory       |
| Visible to code        | Yes — instructions explicitly name registers            | No — transparent, looks like fast RAM                          |
| Size                   | 16 slots x 8 bytes = 128 bytes                          | L1: ~64 KB, L2: ~256 KB, L3: ~16 MB                            |
| Can compute here       | **Yes** — ALU works directly on registers               | **No** — must load into a register first                       |

Caches don't replace registers. They make `LOAD` and `STORE` faster. The CPU still can't add two values in cache — it loads them into registers (hopefully from cache instead of slow RAM), computes, then stores back.

```
                      ┌──────────┐
                      │ Registers│  ← CPU works here
                      │ 128 B    │
                      └────┬─────┘
                    load ↑ │ store
                      ┌────┴─────┐
                      │ L1 Cache │  ← hardware keeps hot data here
                      │ 64 KB    │
                      └────┬─────┘
                      ┌────┴─────┐
                      │ L2 Cache │
                      │ 256 KB   │
                      └────┬─────┘
                      ┌────┴─────┐
                      │ L3 Cache │  ← shared across cores
                      │ 16 MB    │
                      └────┬─────┘
                      ┌────┴─────┐
                      │   RAM    │
                      │ 16+ GB   │
                      └──────────┘

Faster, smaller ↑     ↓ Slower, bigger
```

## Instructions — What the CPU Understands

Your code gets translated to **machine instructions**. Different languages get there differently:

```
Your Code
   │
   ├── Go/Rust/C ──→ compiler ──→ machine code (runs directly on CPU)
   │
   ├── Java ──→ compiler ──→ bytecode ──→ JVM ──→ JIT compiler ──→ machine code
   │
   ├── TypeScript ──→ tsc ──→ JavaScript ──→ V8 ──→ JIT compiler ──→ machine code
   │
   └── Python ──→ interpreter ──→ bytecode ──→ CPython VM (interprets each instruction)
```

**Ahead-of-time compiled** (Go, Rust, C) — the translation happens once at build time. The CPU runs native instructions directly. No overhead at runtime.

**JIT compiled** (Java, JavaScript/TypeScript) — code starts interpreted, then the runtime watches which functions run frequently ("hot paths") and compiles those to native machine code on the fly. First run is slow, subsequent runs approach native speed. Java and V8 (Node.js) are surprisingly fast because of this — their JIT compilers are extremely sophisticated.

**Interpreted** (Python, Ruby) — the VM reads each bytecode instruction and executes a corresponding chunk of C code. Every operation has overhead. A Python `for` loop is ~100x slower than the same loop in Go, not because the algorithm is different, but because each iteration goes through the interpreter.

### Real cost comparison

Adding two integers, roughly:

| Language               | What happens                                                          | Time    |
| ---------------------- | --------------------------------------------------------------------- | ------- |
| Go/Rust/C              | Single `ADD` instruction                                              | ~0.3 ns |
| Java (warmed)          | Same — JIT compiled to `ADD`                                          | ~0.3 ns |
| JavaScript (V8 warmed) | Same — JIT compiled to `ADD`                                          | ~0.3 ns |
| Python                 | Unwrap object, check type, unbox int, add, box result, wrap as object | ~50 ns  |

Java and JS achieve native speed on hot paths. Python pays interpreter + object overhead on every operation. This is why "Python is slow" — it's not the language design, it's the execution model.

### How the compiler controls the CPU

The compiler doesn't move anything into registers at runtime. It writes machine code that **tells the CPU** to do it. The instruction `MOV RAX, [address]` is just bytes in a file — nothing happens until the CPU executes it.

```
1. Compiler writes:    48 01 D8              ← bytes in binary file (means ADD RAX, RBX)
2. OS loads binary into memory                ← still just bytes in RAM
3. CPU fetches this instruction               ← CPU reads the bytes
4. CPU executes it                            ← NOW the registers change
```

The compiler's "choice" of registers is: which register names it encodes into the instruction bytes. `MOV RAX, [addr]` and `MOV R10, [addr]` are different bytes. The compiler picks at build time. The CPU blindly does what the instruction says.

Compiler quality matters because two compilers given the same source produce different instructions:

```
Good compiler:          Poor compiler:
  MOV RAX, [a]            MOV RAX, [a]
  ADD RAX, [b]            MOV [temp], RAX    ← unnecessary store
  MOV [x], RAX            MOV RBX, [b]
                           MOV RAX, [temp]   ← unnecessary load
                           ADD RAX, RBX
                           MOV [x], RAX
```

Same result, but the poor compiler emits extra memory operations. The CPU faithfully executes both — it doesn't optimise your instructions, it just runs them.

### How an interpreted language differs

In a compiled language, the compiler produces CPU instructions and gets out of the way. In an interpreted language, there's a **middleman that never leaves**. The interpreter is itself a compiled C program that reads your bytecode at runtime, instruction by instruction.

`x = a + b` in Python under the hood:

```
Your Python:     x = a + b
                     ↓
Python compiler (at import time) produces bytecode:
    LOAD_FAST   a
    LOAD_FAST   b
    BINARY_ADD
    STORE_FAST  x
                     ↓
CPython interpreter (a C program) runs a loop:

while (has_more_bytecodes) {
    opcode = next_bytecode()

    switch (opcode) {
        case LOAD_FAST:
            // ~10 machine instructions to look up variable,
            // push onto Python stack
        case BINARY_ADD:
            // ~50 machine instructions to:
            //   pop two objects from Python stack
            //   check: int? float? string? custom __add__?
            //   unbox the actual values
            //   do the add
            //   box result into a new Python object
            //   push result onto Python stack
        case STORE_FAST:
            // ~10 machine instructions to pop and store
    }
}
```

The CPU is still executing machine instructions — but they're the **interpreter's** instructions, not yours. Your `a + b` becomes ~70 machine instructions because the interpreter manages its own stack in RAM, looks up types at runtime, and boxes/unboxes values on every operation.

### Where TypeScript/JavaScript fits in

TypeScript types **disappear completely** before execution:

```
TypeScript:   const x: number = a + b
                  ↓
tsc compiler: strips types → const x = a + b    (plain JavaScript)
                  ↓
V8 engine:    runs the JavaScript
```

V8 never sees TypeScript types. But JavaScript isn't as slow as Python because V8 uses JIT compilation in phases:

```
Phase 1 — Interpreter (first run):
  Similar to Python. Interprets bytecode, slow.
  But V8 is profiling — watching which functions run often,
  what types the variables actually are.

Phase 2 — JIT compile (after many calls):
  V8: "This function has been called 1000 times.
       'a' was always a number. 'b' was always a number.
       I'll compile to native code assuming they're numbers."
  Produces:    ADD RAX, RBX    ← same as Go

Phase 3 — Deoptimise (if assumption breaks):
  Someone passes a string where a number was expected.
  V8: "My assumption was wrong. Throw away compiled code,
       fall back to interpreter, re-profile."
```

V8's internal type tracking is more granular than TypeScript's. TypeScript has `number`. V8 internally distinguishes between SMI (Small Integer — no heap allocation, fastest) and HeapNumber (boxed double, heap allocated, slower). It makes these decisions regardless of whether you wrote TypeScript or plain JavaScript.

### Execution model comparison

|               | Build time                           | First run           | Warmed up    | Type changes at runtime     |
| ------------- | ------------------------------------ | ------------------- | ------------ | --------------------------- |
| Go/Rust/C     | Compiles to native code, types known | Fast immediately    | Fast         | N/A — won't compile         |
| Java          | Compiles to bytecode                 | Slow (interpreting) | Fast (JIT'd) | Limited — types are checked |
| TypeScript/JS | Types erased (TS), interpreted       | Slow (interpreting) | Fast (JIT'd) | Deoptimise, recompile       |
| Python        | Compiled to bytecode                 | Slow (interpreting) | Still slow   | No cost — always dynamic    |

The real cost of JavaScript/TypeScript isn't steady-state performance on hot paths — V8 gets close to native. It's startup time (JIT hasn't kicked in), deoptimisation (types change, compiled code thrown away), memory overhead (profiling data, multiple compiled versions), and unpredictable latency (JIT compilation causes pauses). Go avoids all of this by paying the cost once at build time.

### Why JIT languages aren't all equal

Java and JavaScript both use JIT compilation, but Java is consistently faster. The difference is types.

Java knows types at compile time. JavaScript discovers them at runtime:

```
Java:
  int a = 5;     // The JVM *knows* this is an int. Always. Forever.
  int b = 10;    // No need to check. Compile directly to ADD.

JavaScript:
  let a = 5;     // V8 *observes* this has been a number so far...
  let b = 10;    // ...and *speculates* it will stay a number.
                  // Compiles to ADD, but inserts a type guard just in case.
```

Java's JIT compiles with **certainty**. JavaScript's JIT compiles with **assumptions** that can break:

```javascript
function add(a, b) {
  return a + b;
}

add(1, 2); // V8: "a and b are numbers, I'll compile for numbers"
add(3, 4); // fast — JIT compiled path
add(5, 6); // fast
add("oh", "no"); // V8: "my assumption was wrong"
//   → throw away compiled code
//   → fall back to interpreter
//   → re-profile
//   → recompile with broader type assumptions
```

This deoptimisation can't happen in Java because `int a` is always an `int`. The compiler never needs to guard against it suddenly being a string.

The other difference is memory layout. Java gives you `int[]`, `double[]` — guaranteed contiguous C-style arrays for primitives. JavaScript has `Float64Array` for that, but regular arrays depend on V8's internal heuristics.

Rough real-world speed hierarchy:

```
Go/Rust/C          ████████████████████████  fastest (native, no runtime)
Java (warmed)       ███████████████████████  ~1-2x slower than Go
JavaScript (warmed) ████████████████████     ~2-5x slower than Go
Python              ████                     ~50-100x slower than Go
```

Java sits between Go and JavaScript because it has JIT like JavaScript but with the advantage of a static type system — fewer guards, no deoptimisation surprises, and better memory layout for primitives.

### Why Python can't just do the same thing

Python also creates bytecode at import time — so why doesn't it JIT compile like Java or JavaScript? It could, but Python's extreme dynamism makes it much harder.

**Java's bytecode is rich with type information:**

```
iadd    // integer add — the JIT knows both operands are ints
dadd    // double add — the JIT knows both operands are doubles
```

Java's bytecode has separate instructions per type. The JIT doesn't need to guess.

**Python's bytecode is type-blind:**

```
LOAD_FAST   a          # push whatever 'a' is
LOAD_FAST   b          # push whatever 'b' is
BINARY_ADD             # add... somehow? ints? strings? custom objects?
STORE_FAST  x
```

`BINARY_ADD` has no idea what it's adding. The interpreter figures it out every single time.

A JIT could observe and speculate like V8 does — "a has always been an int, compile for ints." But Python lets you change almost anything at any time from anywhere:

```python
# You can change what + means after creation
class Vector:
    def __add__(self, other):
        return Vector(self.x + other.x)

Vector.__add__ = lambda self, other: "surprise"  # monkey-patch

# You can modify a class from any module
def some_unrelated_function():
    Vector.__add__ = something_else

# You can even shadow builtins
import builtins
builtins.int = str
```

A Python JIT would have to insert guards for all of this. Every `a + b` needs: "is `a` still an int? Is `int.__add__` still the original? Has anyone monkey-patched it?" The guards eat most of the performance gain.

Compare the level of dynamism:

```
Easiest to JIT                              Hardest to JIT
──────────────────────────────────────────────────────────
Java          JavaScript        Python
(types fixed) (shapes change    (everything can
               but operators     change at any time
               are stable)       from anywhere)
```

- **Java** — `int a` is always an int. `+` can't be overridden. No guards needed. Ever.
- **JavaScript** — objects can change shape, but operators are stable. V8 guards on type only.
- **Python** — types, operators, methods, even builtins can be replaced at any time from anywhere. Guards needed for everything.

Python JIT implementations do exist — **PyPy** achieves 2-10x speedups over CPython using trace-based JIT compilation, and **CPython 3.13+** is experimenting with a copy-and-patch JIT. But neither can match Java/JS speeds because the guard overhead from Python's dynamism is a fundamental cost.

## The Instruction Pipeline

Modern CPUs don't actually do fetch-decode-execute one at a time. They **pipeline** it — like an assembly line in a factory.

```
Time →    1    2    3    4    5    6    7
         ┌────┬────┬────┬────┐
Instr 1  │ F  │ D  │ E  │ W  │
         └────┼────┼────┼────┼────┐
Instr 2       │ F  │ D  │ E  │ W  │
              └────┼────┼────┼────┼────┐
Instr 3            │ F  │ D  │ E  │ W  │
                   └────┼────┼────┼────┘
Instr 4                 │ F  │ D  │ E  │...

F = Fetch, D = Decode, E = Execute, W = Write back
```

Without pipelining: 4 instructions take 16 cycles.
With pipelining: 4 instructions take 7 cycles. Each new instruction starts one cycle after the last.

Real CPUs have 15-20 pipeline stages, not 4. This means many instructions are "in flight" simultaneously.

### Branch Prediction — Guessing the Future

Pipelines break when the CPU hits a conditional branch:

```go
if x > 0 {
    doA()
} else {
    doB()
}
```

The CPU doesn't know which path to take until the comparison finishes (deep in the pipeline). But it needs to keep feeding instructions into the pipeline NOW. So it **guesses** — this is branch prediction.

Modern CPUs guess correctly ~95-99% of the time by tracking patterns. When they guess wrong, they throw away all the speculatively executed work and restart from the correct branch. This **pipeline flush** costs ~15-20 cycles.

### When this matters in real code

Sorted vs unsorted data:

```go
// Branch predictor loves this — pattern is: false false false... true true true
sort.Ints(data)
for _, v := range data {
    if v > 128 {
        sum += v
    }
}

// Branch predictor struggles — random true/false pattern
// (unsorted data with random values)
for _, v := range data {
    if v > 128 {
        sum += v
    }
}
```

The sorted version can be 2-5x faster on the same data because the branch predictor sees a clean pattern. This is true in Go, Java, C, any language — it's a hardware effect.

This is the same in any language. This famous Stack Overflow question ("Why is processing a sorted array faster than processing an unsorted array?") has been viewed 1.5M+ times — it surprised a lot of experienced developers.

## Superscalar Execution — Multiple Instructions Per Cycle

Modern CPUs can execute **multiple instructions simultaneously** if they don't depend on each other.

```go
// These are independent — CPU can execute all three in parallel
a := x + 1
b := y + 2
c := z + 3

// These are dependent — must execute sequentially
a := x + 1
b := a + 2    // needs a
c := b + 3    // needs b
```

A modern CPU has multiple execution units (ALUs, load/store units, etc.) and can issue 4-6 instructions per cycle if there are no dependencies. The CPU figures out these dependencies automatically — this is called **out-of-order execution**. You write sequential code, the hardware parallelises what it can.

### Why this matters for programming

You mostly don't need to think about this — the CPU and compiler handle it. But it explains why:

- **Simple data transformations are fast** — operations on independent array elements parallelise naturally
- **Linked list traversal is slow** — each `node.next` depends on the previous load, the CPU can't look ahead
- **Tight dependency chains are bottlenecks** — `a = f(a)` repeated 1000 times can't be parallelised

## Putting It All Together — What Happens When You Run a Go Program

```go
func sum(nums []int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}
```

### Step 1 — Compilation

`go build` turns this into machine instructions. The compiler assigns variables to registers and produces a tight loop:

```asm
# RAX = pointer to array, RBX = length, RCX = total, RDX = i

    XOR  RCX, RCX          # total = 0
    XOR  RDX, RDX          # i = 0
loop:
    CMP  RDX, RBX          # i >= length?
    JGE  done              # yes → exit
    MOV  R8, [RAX + RDX*8] # load nums[i]
    ADD  RCX, R8           # total += nums[i]
    INC  RDX               # i++
    JMP  loop
done:
    MOV  RAX, RCX          # return total
    RET
```

These are just bytes in a file. Nothing runs until the CPU executes them.

### Step 2 — Loading and instruction fetch

The OS loads the binary into memory. The entire loop is ~23 bytes — fits in a single L1 cache line (64 bytes). After the first iteration, every instruction fetch costs ~1 ns.

### Step 3 — Branch prediction

The loop has one conditional branch: `JGE done` (are we finished?). The CPU guesses it will be `false` every iteration. It's right 999,999 out of 1,000,000 times — the one miss at the end costs ~15 cycles and happens once total. The pipeline stays full for virtually the entire loop.

### Step 4 — Memory access

Each iteration does `MOV R8, [RAX + RDX*8]` to load `nums[i]`:

- **First access**: cache miss — fetches 64 bytes (8 ints) from RAM into L1. Costs ~100 ns.
- **Next 7 iterations**: same cache line, ~1 ns each.
- **After that**: the hardware prefetcher detects the sequential pattern and fetches ahead. Remaining million loads cost ~1 ns each.

Out of a million elements, you pay the ~100 ns RAM penalty exactly once.

### Step 5 — Register operation and return

`total` lives in register RCX for the entire loop — it never touches RAM. Each `ADD RCX, R8` costs 1 cycle (~0.3 ns). On return, the result moves to RAX (the return value register) and the function exits. No memory involved.

**Effective throughput**: ~2-3 cycles per iteration. A million elements takes roughly 1 ms. The same code in Python goes through the interpreter on every addition, unwrapping and re-wrapping integer objects — same algorithm, ~100x slower.

## Key Takeaways

- **Everything becomes CPU instructions** — the question is how directly your language maps to them
- **Registers are king** — the compiler's ability to keep values in registers is a huge performance factor
- **Compiled > JIT > interpreted** for raw compute, but JIT closes the gap on hot paths
- **Pipelines want predictable code** — sorted data, consistent branches, no surprises
- **Independence enables parallelism** — the CPU can execute multiple independent operations per cycle without you doing anything
- **Sequential memory access is fast** — this is a preview of the next lesson on memory hierarchy, but the CPU is optimised for accessing memory in order, not randomly

## What's Next

The CPU is fast, but it spends most of its time **waiting for memory**. The next lesson covers the memory hierarchy — caches, cache lines, and why the way you lay out data matters more than the algorithm you choose.
