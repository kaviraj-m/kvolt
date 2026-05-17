package grpc

import "google.golang.org/grpc"

// config holds internal configuration built from functional options.
type config struct {
	serverOptions      []grpc.ServerOption
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	reflection         bool
	health             bool
}

// Option configures a gRPC Server.
type Option func(*config)

// WithServerOptions appends native gRPC server options.
// Use this for advanced configuration like credentials, keepalive, etc.
//
//	grpc.WithServerOptions(grpc.Creds(creds), grpc.MaxRecvMsgSize(4<<20))
func WithServerOptions(opts ...grpc.ServerOption) Option {
	return func(c *config) {
		c.serverOptions = append(c.serverOptions, opts...)
	}
}

// WithUnaryInterceptors appends unary server interceptors.
// Interceptors execute in the order they are added.
//
//	grpc.WithUnaryInterceptors(grpc.LoggingInterceptor(), grpc.RecoveryInterceptor())
func WithUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(c *config) {
		c.unaryInterceptors = append(c.unaryInterceptors, interceptors...)
	}
}

// WithStreamInterceptors appends stream server interceptors.
// Interceptors execute in the order they are added.
func WithStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) Option {
	return func(c *config) {
		c.streamInterceptors = append(c.streamInterceptors, interceptors...)
	}
}

// WithReflection enables gRPC server reflection.
// This allows tools like grpcurl and grpcui to discover services dynamically.
func WithReflection(enable bool) Option {
	return func(c *config) {
		c.reflection = enable
	}
}

// WithHealthCheck enables the gRPC health checking protocol (grpc.health.v1).
// This is useful for load balancers and orchestration systems like Kubernetes.
func WithHealthCheck(enable bool) Option {
	return func(c *config) {
		c.health = enable
	}
}
