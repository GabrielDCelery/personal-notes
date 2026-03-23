# Communication Protocols — Distilled

HTTP request/response is the default. Deviate only when the client needs updates it didn't ask for, or when the latency of a full round trip per message is unacceptable.

## The Core Decision

```
Client asks for data it requested              → REST
Server needs to push updates to the client     → SSE or WebSockets
Client also sends data back at high frequency  → WebSockets
Updates are infrequent, simplicity matters     → Polling
```

The question is: **who initiates the message, and how often?**

## REST / HTTP

The default. Client sends a request, server returns a response.

```
Client → GET /orders/123 → Server
Client ←    { order }    ← Server
```

- Stateless — every request carries full context
- Cacheable at every layer (CDN, browser, reverse proxy)
- Horizontally scalable — any server handles any request
- Simple to debug, monitor, and trace

**Default to REST.** Only reach for other protocols when REST genuinely can't do the job.

## Polling

Client asks the server on a timer: "anything new?"

```
Short polling (timer-based):
  Client → GET /notifications       every 5 seconds
  Server ← 200 { events: [] }       (empty most of the time)

Long polling (client waits):
  Client → GET /notifications
  Server holds the connection open (up to 30s)
  Server ← 200 { events: [...] }    (when something happens, or on timeout)
  Client → GET /notifications       (immediately reconnects)
```

|                    | Short polling          | Long polling      |
| ------------------ | ---------------------- | ----------------- |
| Latency            | Up to polling interval | Near real-time    |
| Wasted requests    | Many (mostly empty)    | Few               |
| Server connections | Brief, low             | Held open, higher |
| Complexity         | Minimal                | Low               |

Short polling at scale: 100K users polling every 5s = **20K req/sec of mostly empty responses.** It adds up fast.

**Use short polling only for low-frequency updates where real-time doesn't matter.** Long polling is a reasonable fallback when SSE or WebSockets aren't available.

## SSE (Server-Sent Events)

Server pushes a stream of events over a single persistent HTTP connection. Client connects once; server sends whenever it has something.

```
Client → GET /events
         Accept: text/event-stream

Server ← HTTP 200
         Content-Type: text/event-stream

         data: {"type":"order_update","id":"123"}\n\n
         data: {"type":"notification","msg":"..."}\n\n
         ...
```

- **Unidirectional** — server to client only
- Built on HTTP — works through proxies, load balancers, CDNs without configuration
- Browser reconnects automatically on disconnect (built into the spec)
- Text only (UTF-8)

**Use SSE for server-to-client push where the client doesn't need to send data back.** Dashboards, notifications, live feeds, progress bars, log streaming.

## WebSockets

Full-duplex connection over a single TCP socket. After an HTTP upgrade handshake, both sides send frames independently at any time.

```
HTTP upgrade:
  Client → GET /ws
           Upgrade: websocket
  Server ← 101 Switching Protocols

After upgrade — persistent, bidirectional:
  Client → { "action": "subscribe", "channel": "btc/usd" }
  Server → { "price": 62341.00 }
  Server → { "price": 62289.00 }
  Client → { "action": "trade", "qty": 0.1 }
```

- **Bidirectional** — both sides send and receive independently
- Low per-message overhead after upgrade (no HTTP headers per frame)
- Persistent and stateful — server must track each connected client
- Binary or text

**Use WebSockets when the client also sends messages at high frequency or when low latency in both directions matters.** Chat, collaborative editing, multiplayer games, trading platforms.

## Protocol Comparison

|                    | REST             | Short polling     | SSE                  | WebSockets                 |
| ------------------ | ---------------- | ----------------- | -------------------- | -------------------------- |
| **Direction**      | Request/response | Request/response  | Server → client      | Bidirectional              |
| **Latency**        | On request       | Up to interval    | ~1–10 ms             | ~1–5 ms                    |
| **Connection**     | Short-lived      | Repeated          | Persistent           | Persistent                 |
| **Overhead**       | HTTP headers/req | HTTP headers/req  | One handshake        | One handshake, tiny frames |
| **Scaling**        | Easy (stateless) | Easy              | Medium               | Hard (stateful)            |
| **Proxy / CDN**    | Full support     | Full support      | Generally works      | Needs config               |
| **Auto-reconnect** | N/A              | Client-side       | Built in             | Must implement             |
| **Best for**       | APIs, CRUD       | Infrequent checks | Notifications, feeds | Chat, collab, games        |

## Latency Reference

```
Scale: log10  |  1ms       10ms      100ms     1s        10s |
              0---+---------+---------+---------+---------+--+

WebSockets                  |███                 | ~1–5 ms
SSE                         |████                | ~1–10 ms
Long polling                |████████            | ~10–100 ms (server hold time)
Short polling (5s interval) |                    | up to 5,000 ms
```

## Scaling Persistent Connections

REST scales trivially — any server handles any request. Persistent connections (SSE, WebSockets) introduce state.

```
Problem:
  Client A connected to Server 1
  Event for Client A published by Server 2
  Server 2 can't reach Client A's socket

Solutions:
  1. Sticky sessions       — LB always routes the same client to the same server
                             (works, but uneven load distribution)
  2. Pub/sub backplane     — Redis pub/sub or Kafka; any server publishes,
                             all servers receive, each delivers to its own clients
  3. Dedicated gateway     — separate WebSocket/SSE service handles all connections;
                             app servers remain stateless, communicate via pub/sub
```

A server can typically hold 10K–100K concurrent WebSocket connections depending on message rate and payload size. The bottleneck is usually memory and file descriptors, not CPU.

**Don't put WebSocket connections on your API servers.** Run a dedicated connection gateway that bridges to stateless backends via pub/sub. This keeps your app servers simple and independently scalable.

## Decision Framework

```
Does the client need updates it didn't request?
  No  → REST

  Yes → Does the client also send data at high frequency?
    No  → SSE
    Yes → WebSockets

  How often are updates?
    Infrequent (> 30s, low stakes)  → Short polling
    Frequent but client is passive  → SSE
    Frequent and bidirectional      → WebSockets
```

## Key Mental Models

1. **Default to REST.** Only deviate when the client needs unprompted server updates.
2. **SSE for server push. WebSockets for bidirectional.** SSE is simpler, HTTP-native, and sufficient for most real-time use cases.
3. **Polling has a hidden cost at scale.** 100K users polling every 5s = 20K req/sec of empty responses.
4. **Persistent connections are stateful — plan for it.** Use a pub/sub backplane so any server can reach any client.
5. **Don't put WebSocket connections on your API servers.** Dedicated gateway + pub/sub keeps backends stateless.
6. **SSE reconnects automatically. WebSockets don't.** Implement reconnect logic with exponential backoff on the client.
7. **SSE is text-only. WebSockets support binary.** For binary protocols (audio, video, custom framing), WebSockets or WebTransport.
