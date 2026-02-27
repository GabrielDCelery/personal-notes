# Lesson 17: Advanced Select Patterns

Concept: Select has powerful patterns beyond basic channel operations: non-blocking operations with default, priority handling, and disabling cases with nil channels.

Your task:

1. Create a program that processes messages from two channels with priority:
   - urgent channel (high priority)
   - normal channel (low priority)
   - Always process urgent messages first if available

2. Create a producer goroutine that:
   - Sends 3 messages to urgent channel (every 300ms)
   - Sends 10 messages to normal channel (every 100ms)
   - Closes both channels when done

3. Create a consumer function that:
   - Checks urgent channel first (non-blocking)
   - Only checks normal if no urgent messages
   - Prints which type of message received
   - Exits when both channels are closed

4. Track counts: print how many urgent vs normal processed at the end

Expected behavior:

- Urgent messages should be processed before normal, even if normal arrives first
- Output shows urgent messages "jump the queue"

Hints:

```go
  Non-blocking receive:
  select {
  case msg := <-ch:
      // got message
  default:
      // channel empty
  }

  Priority pattern:
  for {
      select {
      case msg, ok := <-urgent:
          if ok {
              // handle urgent
              continue  // check urgent again
          }
          urgent = nil  // disable this case
      default:
          // no urgent, check both
          select {
          case msg, ok := <-normal:
              // handle normal
          case msg, ok := <-urgent:
              // handle urgent
          }
      }
  }
```

Nil channel trick: urgent = nil disables that select case (nil channels block forever).

Key learning: These patterns let you build sophisticated message handling - priority queues, optional channels, timeouts with fallbacks, etc.

> what is the point of having a default and nested select case, couldnt solve this problem with a single select?

● Great question. A single select does not guarantee priority - it chooses randomly when multiple cases are ready:

// Single select - NO priority
select {
case msg := <-urgent:
// handle
case msg := <-normal:
// handle
}

If both channels have messages, Go picks one at random (50/50 chance). You might process a normal message while urgent ones are waiting.

The nested pattern forces priority:

select {
case msg := <-urgent:
// handle urgent
default:
// Only reaches here if urgent is EMPTY
select {
case msg := <-urgent:
// check again (might have arrived)
case msg := <-normal:
// only if still no urgent
}
}

1. First select: try urgent (non-blocking via default)
2. If urgent is empty → fall into default
3. Second select: now wait for either, but we already know urgent was empty

Demonstration:

urgent := make(chan int, 5)
normal := make(chan int, 5)

// Both have messages
urgent <- 1
normal <- 1

// Single select: might pick normal (random)
// Nested select: guarantees urgent first

So yes, you could use a single select, but you'd lose the priority guarantee - which is the whole point of this pattern.

> id we know urgent is empty why check again for urgent, why not just have case msg:= <-normal

● Because time passes between the outer default and inner select. During that tiny gap, a new urgent message could arrive.

If inner select only had <-normal:

select {
case msg := <-urgent:
// handle
default:
// urgent was empty HERE

      // ... nanoseconds pass ...
      // ... new urgent message arrives ...

      select {
      case msg := <-normal:  // BLOCKS waiting for normal
          // Meanwhile urgent is waiting, ignored!
      }

}

You'd be stuck waiting for normal while urgent messages pile up.

With both:

default:
select {
case msg := <-urgent: // catches any that arrived
case msg := <-normal:
}
}

Now if urgent arrives while waiting, you'll catch it.

Think of it as:

- Outer select: "Is urgent ready RIGHT NOW?"
- Inner select: "Wait for the next message from either, but we know urgent wasn't ready a moment ago"

It's not perfect priority (there's still a race), but it's much better than random 50/50.

> does this mean a select will block until it receives something?

● Yes, a select without default blocks until one case is ready:

// Blocks until ch1 OR ch2 has a message
select {
case msg := <-ch1:
case msg := <-ch2:
}

With default, it becomes non-blocking:

// Never blocks - immediately takes default if nothing ready
select {
case msg := <-ch1:
case msg := <-ch2:
default:
// instant fallthrough if both empty
}

| Select type  | Behavior                                |
| ------------ | --------------------------------------- |
| No default   | Blocks until a case is ready            |
| With default | Never blocks, falls through immediately |

That's why the nested pattern works:

- Outer select (with default): non-blocking check for urgent
- Inner select (no default): blocks waiting for next message from either
