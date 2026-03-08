# gRPC & Protocol Buffers

```sh
npm install @grpc/grpc-js @grpc/proto-loader
# Or for generated code:
npm install @grpc/grpc-js google-protobuf
npm install -D grpc-tools grpc_tools_node_protoc_ts
```

> High-performance RPC framework. Used for service-to-service communication.

## Proto Definition

```protobuf
// proto/user/v1/user.proto
syntax = "proto3";

package user.v1;

message User {
  string id = 1;
  string name = 2;
  string email = 3;
}

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  User user = 1;
}

message ListUsersRequest {
  int32 page_size = 1;
  string page_token = 2;
}

message ListUsersResponse {
  repeated User users = 1;
  string next_page_token = 2;
}

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}
```

## Dynamic Loading (no code generation)

```typescript
import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";

const packageDef = protoLoader.loadSync("proto/user/v1/user.proto", {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

const proto = grpc.loadPackageDefinition(packageDef) as any;
const userProto = proto.user.v1;
```

## Server

```typescript
function getUser(
  call: grpc.ServerUnaryCall<any, any>,
  callback: grpc.sendUnaryData<any>,
) {
  const id = call.request.id;
  const user = store.find(id);

  if (!user) {
    callback({
      code: grpc.status.NOT_FOUND,
      message: `User ${id} not found`,
    });
    return;
  }

  callback(null, { user });
}

function listUsers(
  call: grpc.ServerUnaryCall<any, any>,
  callback: grpc.sendUnaryData<any>,
) {
  const users = store.findAll();
  callback(null, { users, next_page_token: "" });
}

const server = new grpc.Server();
server.addService(userProto.UserService.service, { getUser, listUsers });

server.bindAsync(
  "0.0.0.0:50051",
  grpc.ServerCredentials.createInsecure(),
  (err, port) => {
    if (err) throw err;
    console.log(`gRPC server listening on :${port}`);
  },
);
```

## Client

```typescript
const client = new userProto.UserService(
  "localhost:50051",
  grpc.credentials.createInsecure(),
);

// Unary call — callback
client.getUser(
  { id: "123" },
  (err: grpc.ServiceError | null, response: any) => {
    if (err) {
      console.error(`gRPC error: code=${err.code} message=${err.message}`);
      return;
    }
    console.log(response.user.name);
  },
);

// Promisified
function getUser(id: string): Promise<any> {
  return new Promise((resolve, reject) => {
    client.getUser({ id }, (err: grpc.ServiceError | null, response: any) => {
      if (err) reject(err);
      else resolve(response);
    });
  });
}

const response = await getUser("123");
```

## Status Codes

```typescript
import * as grpc from "@grpc/grpc-js";

callback({ code: grpc.status.NOT_FOUND, message: "user not found" });
callback({ code: grpc.status.INVALID_ARGUMENT, message: "name is required" });
callback({ code: grpc.status.INTERNAL, message: "database error" });
callback({ code: grpc.status.UNAUTHENTICATED, message: "invalid token" });
callback({ code: grpc.status.PERMISSION_DENIED, message: "not allowed" });
```

| Code                | When to use                      |
| ------------------- | -------------------------------- |
| `NOT_FOUND`         | Resource doesn't exist           |
| `INVALID_ARGUMENT`  | Bad request / validation failure |
| `INTERNAL`          | Unexpected server error          |
| `UNAUTHENTICATED`   | Missing or invalid credentials   |
| `PERMISSION_DENIED` | Authenticated but not authorized |
| `ALREADY_EXISTS`    | Duplicate resource               |
| `DEADLINE_EXCEEDED` | Timeout                          |
| `UNAVAILABLE`       | Service temporarily down         |

## Interceptors (Middleware)

```typescript
// Server interceptor
function loggingInterceptor(methodDescriptor: any, call: any) {
  const start = Date.now();
  const original = call.handler;

  call.handler = (call: any, callback: any) => {
    original(call, (err: any, response: any) => {
      console.log(`${methodDescriptor.path} took ${Date.now() - start}ms`);
      callback(err, response);
    });
  };
}
```

## Metadata (headers)

```typescript
// Client — send metadata
const metadata = new grpc.Metadata();
metadata.add("authorization", `Bearer ${token}`);

client.getUser({ id: "123" }, metadata, (err, response) => {
  // ...
});

// Server — read metadata
function getUser(
  call: grpc.ServerUnaryCall<any, any>,
  callback: grpc.sendUnaryData<any>,
) {
  const token = call.metadata.get("authorization")[0] as string;
  // ...
}
```

## Debugging

```sh
# Install grpcurl
brew install grpcurl

# List services (requires reflection)
grpcurl -plaintext localhost:50051 list

# Call an RPC
grpcurl -plaintext -d '{"id": "123"}' localhost:50051 user.v1.UserService/GetUser
```

## Code Generation (alternative — buf + ts-proto)

```sh
npm install -D @bufbuild/buf ts-proto
```

```yaml
# buf.gen.yaml
version: v2
plugins:
  - remote: buf.build/community/timostamm-protobuf-ts
    out: gen
```

```sh
npx buf generate
```

Generates fully typed TypeScript client and server interfaces.
