# OpenTelemetry (otel)

```sh
npm install @opentelemetry/sdk-node \
  @opentelemetry/api \
  @opentelemetry/auto-instrumentations-node \
  @opentelemetry/exporter-trace-otlp-grpc
```

> Observability framework for traces, metrics, and logs. Vendor-neutral — works with Jaeger, Datadog, Grafana, etc.

## Concepts

| Term            | Meaning                                           |
| --------------- | ------------------------------------------------- |
| Trace           | Full journey of a request across services         |
| Span            | Single unit of work within a trace                |
| Context         | Carries the active span across function calls     |
| Exporter        | Sends traces to a backend (Jaeger, Datadog, etc.) |
| Instrumentation | Auto or manual creation of spans                  |

## Setup (auto-instrumentation)

```typescript
// tracing.ts — import BEFORE anything else
import { NodeSDK } from "@opentelemetry/sdk-node";
import { getNodeAutoInstrumentations } from "@opentelemetry/auto-instrumentations-node";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-grpc";
import { Resource } from "@opentelemetry/resources";
import { ATTR_SERVICE_NAME } from "@opentelemetry/semantic-conventions";

const sdk = new NodeSDK({
  resource: new Resource({
    [ATTR_SERVICE_NAME]: "my-service",
  }),
  traceExporter: new OTLPTraceExporter({
    url: "http://localhost:4317",
  }),
  instrumentations: [getNodeAutoInstrumentations()],
});

sdk.start();

process.on("SIGTERM", async () => {
  await sdk.shutdown();
});
```

```sh
# Run with tracing loaded first
node --require ./dist/tracing.js ./dist/server.js
# Or with --import for ESM
node --import ./dist/tracing.js ./dist/server.js
```

Auto-instrumentation creates spans for HTTP, Express, Fastify, pg, mysql, Redis, gRPC, etc.

## Manual Spans

```typescript
import { trace } from "@opentelemetry/api";

const tracer = trace.getTracer("my-service");

async function processOrder(orderId: string) {
  return tracer.startActiveSpan("processOrder", async (span) => {
    try {
      span.setAttribute("order.id", orderId);

      await validateOrder(orderId);
      await chargePayment(orderId);
      await sendConfirmation(orderId);

      return { success: true };
    } catch (err) {
      span.recordException(err as Error);
      span.setStatus({ code: 2, message: (err as Error).message }); // ERROR
      throw err;
    } finally {
      span.end();
    }
  });
}
```

## Add Attributes

```typescript
tracer.startActiveSpan("getUser", async (span) => {
  span.setAttribute("user.id", userId);
  span.setAttribute("db.system", "postgresql");
  span.setAttribute("db.operation", "SELECT");

  const user = await db.users.findUnique({ where: { id: userId } });

  span.setAttribute("user.found", !!user);
  span.end();
  return user;
});
```

## Span Events

```typescript
tracer.startActiveSpan("handleRequest", async (span) => {
  span.addEvent("cache.miss", { "cache.key": cacheKey });

  const data = await fetchFromDb();

  span.addEvent("cache.set", { "cache.key": cacheKey, "cache.ttl": 3600 });
  await cache.set(cacheKey, data, 3600);

  span.end();
  return data;
});
```

## Context Propagation

```typescript
import { context, propagation } from "@opentelemetry/api";

// Inject trace context into outgoing HTTP headers (auto for fetch/http)
const headers: Record<string, string> = {};
propagation.inject(context.active(), headers);
// headers now contains traceparent, tracestate

// Extract from incoming headers
const ctx = propagation.extract(context.active(), req.headers);
context.with(ctx, () => {
  // spans created here are part of the incoming trace
});
```

## Express Middleware (manual span per route)

```typescript
import { trace, SpanStatusCode } from "@opentelemetry/api";

const tracer = trace.getTracer("my-service");

function traceMiddleware(operationName: string) {
  return (req: Request, res: Response, next: NextFunction) => {
    tracer.startActiveSpan(operationName, (span) => {
      span.setAttribute("http.route", req.path);
      span.setAttribute("http.method", req.method);

      res.on("finish", () => {
        span.setAttribute("http.status_code", res.statusCode);
        if (res.statusCode >= 500) {
          span.setStatus({ code: SpanStatusCode.ERROR });
        }
        span.end();
      });

      next();
    });
  };
}
```

## Environment Variables

```sh
# Common OTEL env vars (no code changes needed)
OTEL_SERVICE_NAME=my-service
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
OTEL_TRACES_SAMPLER=parentbased_traceidratio
OTEL_TRACES_SAMPLER_ARG=0.1        # sample 10% of traces
OTEL_LOG_LEVEL=info
```
