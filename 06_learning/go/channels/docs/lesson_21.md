# Lesson 21: Tee Pattern - Split One Channel to Multiple Outputs

Concept: The opposite of merge - take one input channel and duplicate its values to multiple output channels. Like a T-junction in plumbing.

Your task:

1. tee(ctx, in) (out1, out2) - Duplicates input channel to two output channels
   - Reads from in channel
   - Sends each value to both out1 and out2
   - Respects context cancellation
   - Closes both outputs when input closes

2. In main():
   - Create one generator (sends 0-9)
   - Use tee() to split into two channels
   - Create two consumers that read at different speeds:
     - Consumer 1: prints "fast: X" (no delay)
     - Consumer 2: prints "slow: X" (100ms delay)
   - Use context with timeout (2 seconds)
   - Demonstrate that slower consumer blocks the tee

Expected behavior:
fast: 0
slow: 0
fast: 1
slow: 1
...

Challenge question: What happens when one consumer is slower? Does it block the other?
