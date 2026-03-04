# gRPC & Protocol Buffers

```sh
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

> High-performance RPC framework. Used for service-to-service communication.

## Setup

### 1. Install protoc compiler and Go plugins

```sh
# protoc compiler
# Ubuntu/Debian
apt install -y protobuf-compiler

# macOS
brew install protobuf

# Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 2. Define a proto file

```protobuf
// proto/user/v1/user.proto
syntax = "proto3";

package user.v1;

option go_package = "github.com/yourorg/myapp/gen/user/v1;userv1";

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

### 3. Generate Go code

```sh
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/user/v1/user.proto
```

Or with `//go:generate`:

```go
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/user/v1/user.proto
```

## Server

### 4. Implement the service

```go
type userServer struct {
    userv1.UnimplementedUserServiceServer
    store UserStore
}

func (s *userServer) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
    user, err := s.store.Find(ctx, req.GetId())
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
    }

    return &userv1.GetUserResponse{
        User: &userv1.User{
            Id:    user.ID,
            Name:  user.Name,
            Email: user.Email,
        },
    }, nil
}
```

### 5. Start the server

```go
func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatal(err)
    }

    srv := grpc.NewServer()
    userv1.RegisterUserServiceServer(srv, &userServer{store: newStore()})

    // Enable reflection for debugging with grpcurl
    reflection.Register(srv)

    log.Println("gRPC server listening on :50051")
    if err := srv.Serve(lis); err != nil {
        log.Fatal(err)
    }
}
```

## Client

### 6. Create a client and call RPCs

```go
conn, err := grpc.NewClient("localhost:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := userv1.NewUserServiceClient(conn)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.GetUser(ctx, &userv1.GetUserRequest{Id: "123"})
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        log.Printf("gRPC error: code=%s msg=%s", st.Code(), st.Message())
    }
    return
}
fmt.Println(resp.GetUser().GetName())
```

## Error Handling

### 7. gRPC status codes

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// Return errors with status codes
return nil, status.Errorf(codes.NotFound, "user %s not found", id)
return nil, status.Errorf(codes.InvalidArgument, "name is required")
return nil, status.Errorf(codes.Internal, "database error: %v", err)
return nil, status.Errorf(codes.Unauthenticated, "invalid token")
return nil, status.Errorf(codes.PermissionDenied, "not allowed")
```

| Code                 | When to use                              |
| -------------------- | ---------------------------------------- |
| `NotFound`           | Resource doesn't exist                   |
| `InvalidArgument`    | Bad request / validation failure         |
| `Internal`           | Unexpected server error                  |
| `Unauthenticated`    | Missing or invalid credentials           |
| `PermissionDenied`   | Authenticated but not authorized         |
| `AlreadyExists`      | Duplicate resource                       |
| `DeadlineExceeded`   | Timeout                                  |
| `Unavailable`        | Service temporarily down (client retries)|

## Interceptors (Middleware)

### 8. Unary interceptor

```go
func loggingInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    start := time.Now()
    resp, err := handler(ctx, req)
    log.Printf("method=%s duration=%s err=%v", info.FullMethod, time.Since(start), err)
    return resp, err
}

srv := grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor),
)
```

### 9. Chain multiple interceptors

```go
srv := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        recoveryInterceptor,
        loggingInterceptor,
        authInterceptor,
    ),
)
```

## Streaming

### 10. Server-side streaming

```protobuf
service UserService {
  rpc WatchUsers(WatchUsersRequest) returns (stream User);
}
```

```go
// Server
func (s *userServer) WatchUsers(req *userv1.WatchUsersRequest, stream userv1.UserService_WatchUsersServer) error {
    for user := range s.userUpdates {
        if err := stream.Send(user); err != nil {
            return err
        }
    }
    return nil
}

// Client
stream, err := client.WatchUsers(ctx, &userv1.WatchUsersRequest{})
if err != nil {
    return err
}
for {
    user, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        return err
    }
    fmt.Println(user.GetName())
}
```

## Testing

### 11. Test with bufconn (in-memory)

```go
func startTestServer(t *testing.T) userv1.UserServiceClient {
    lis := bufconn.Listen(1024 * 1024)
    srv := grpc.NewServer()
    userv1.RegisterUserServiceServer(srv, &userServer{store: newMockStore()})

    go srv.Serve(lis)
    t.Cleanup(func() { srv.Stop() })

    conn, err := grpc.NewClient("passthrough:///bufnet",
        grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
            return lis.DialContext(ctx)
        }),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    require.NoError(t, err)
    t.Cleanup(func() { conn.Close() })

    return userv1.NewUserServiceClient(conn)
}

func TestGetUser(t *testing.T) {
    client := startTestServer(t)

    resp, err := client.GetUser(context.Background(), &userv1.GetUserRequest{Id: "123"})
    require.NoError(t, err)
    assert.Equal(t, "Alice", resp.GetUser().GetName())
}
```

## Debugging

### 12. grpcurl

```sh
# List services
grpcurl -plaintext localhost:50051 list

# Describe a service
grpcurl -plaintext localhost:50051 describe user.v1.UserService

# Call an RPC
grpcurl -plaintext -d '{"id": "123"}' localhost:50051 user.v1.UserService/GetUser
```

Requires reflection to be enabled on the server.

## Project Layout

```
myapp/
├── proto/
│   └── user/v1/user.proto
├── gen/
│   └── user/v1/
│       ├── user.pb.go
│       └── user_grpc.pb.go
├── internal/
│   └── server/
│       └── user.go        # service implementation
├── cmd/
│   └── server/main.go
└── buf.yaml               # optional — buf.build for proto management
```
