# Does TCP Only See a Single Stream of Data?

## Question
Does this mean the TCP connection itself only sees a single stream of data and does not know what each chunk represents?

## Answer

Yes, exactly! That's a great way to understand it.

### TCP's Perspective

From TCP's point of view:
- It only sees a **continuous byte stream** - just raw binary data
- TCP has **no concept** of HTTP messages, frames, requests, responses, or application-level boundaries
- It just ensures bytes are delivered reliably and in order from sender to receiver
- TCP breaks the byte stream into **segments** for transmission, but these segments are purely for transport purposes and have no relationship to HTTP/2 frames

### HTTP/2's Perspective

HTTP/2 operates on top of this byte stream:
- It writes binary frames into the TCP byte stream
- On the receiving end, it reads from the TCP byte stream and parses out the frame boundaries
- It uses the frame headers (which include length fields and stream IDs) to figure out where one frame ends and another begins
- Then it reassembles frames belonging to the same stream back into complete HTTP messages

### Analogy

Think of it like a postal service:
- **TCP** = the delivery truck that just carries boxes without knowing what's inside
- **HTTP/2** = the warehouse workers who packed items into specific boxes and labeled them, and the workers at the destination who unpack and sort them by label

This layered architecture is fundamental to how the internet works - each layer does its job without needing to understand the layers above it.
