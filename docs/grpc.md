# gRPC Support

KVolt provides first-class gRPC support through the `grpc` package. You get a clean server wrapper, built-in interceptors (logging, recovery, auth), and seamless integration with the HTTP engine.

## Quick Start

```go
package main

import (
    "github.com/go-kvolt/kvolt"
    kvctx "github.com/go-kvolt/kvolt/context"
    kvgrpc "github.com/go-kvolt/kvolt/grpc"
    pb "your-app/proto/gen" // your protoc-generated package
)

func main() {
    app := kvolt.New()

    // HTTP route
    app.GET("/health", func(c *kvctx.Context) error {
        return c.JSON(200, map[string]string{"status": "ok"})
    })

    // gRPC server with interceptors
    grpcSrv := app.GRPCServer(
        kvgrpc.WithReflection(true),
        kvgrpc.WithHealthCheck(true),
        kvgrpc.WithUnaryInterceptors(
            kvgrpc.LoggingInterceptor(),
            kvgrpc.RecoveryInterceptor(),
        ),
    )

    // Register your protobuf service
    pb.RegisterMyServiceServer(grpcSrv.Raw(), &myServiceImpl{})

    // Run HTTP on :8080 and gRPC on :9090
    app.RunAll(":8080", ":9090")
}
```

## Server Configuration

### Creating the gRPC Server

Use `app.GRPCServer(opts...)` to get the gRPC server. It's lazily initialized on first call:

```go
grpcSrv := app.GRPCServer(
    kvgrpc.WithReflection(true),       // Enable grpcurl/grpcui discovery
    kvgrpc.WithHealthCheck(true),      // Enable gRPC health checking protocol
)
```

### Options

| Option | Description |
|--------|-------------|
| `WithReflection(bool)` | Enable gRPC server reflection for tool discovery |
| `WithHealthCheck(bool)` | Enable `grpc.health.v1` health checking protocol |
| `WithUnaryInterceptors(...)` | Add unary interceptors (execute in order) |
| `WithStreamInterceptors(...)` | Add stream interceptors (execute in order) |
| `WithServerOptions(...)` | Pass-through native `grpc.ServerOption` values |

### Registering Services

Use the `Raw()` method to access the underlying `*grpc.Server` for protoc-generated registration:

```go
pb.RegisterGreeterServer(grpcSrv.Raw(), &greeterImpl{})
pb.RegisterUserServer(grpcSrv.Raw(), &userImpl{})
```

## Built-in Interceptors

KVolt gRPC interceptors mirror the HTTP middleware:

### Logging

```go
kvgrpc.LoggingInterceptor()       // Unary
kvgrpc.StreamLoggingInterceptor() // Stream
```

Output: `[gRPC] /hello.Greeter/SayHello | OK | 1.234ms`

### Recovery

```go
kvgrpc.RecoveryInterceptor()       // Unary
kvgrpc.StreamRecoveryInterceptor() // Stream
```

Catches panics and returns `codes.Internal` with a safe error message.

### Auth

```go
kvgrpc.AuthInterceptor(func(ctx context.Context, token string) error {
    // Validate token — return nil if valid, error if not
    if token != "my-secret" {
        return fmt.Errorf("invalid token")
    }
    return nil
})

kvgrpc.StreamAuthInterceptor(validateFunc)
```

Extracts the bearer token from the `authorization` metadata key.

## Running Modes

### Standalone gRPC

```go
grpcSrv := app.GRPCServer(...)
pb.RegisterMyServiceServer(grpcSrv.Raw(), &impl{})
app.RunGRPC(":9090")
```

### HTTP + gRPC (RunAll)

```go
app.RunAll(":8080", ":9090")
```

Both servers shut down gracefully on `SIGINT`/`SIGTERM`.

### Direct Server Control

For advanced use, you can bypass the Engine and use the gRPC server directly:

```go
srv := kvgrpc.NewServer(
    kvgrpc.WithReflection(true),
    kvgrpc.WithUnaryInterceptors(kvgrpc.LoggingInterceptor()),
)
pb.RegisterMyServiceServer(srv.Raw(), &impl{})
srv.ListenAndServe(":9090")
```

## Health Checking

When `WithHealthCheck(true)` is set, the server registers the standard `grpc.health.v1.Health` service. Update service status:

```go
import healthpb "google.golang.org/grpc/health/grpc_health_v1"

grpcSrv.HealthServer().SetServingStatus("myService", healthpb.HealthCheckResponse_SERVING)
```

This integrates with Kubernetes gRPC health probes and load balancers.
