package grpcutils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewGrpcConnection(ctx context.Context, host string) (*grpc.ClientConn, error) {

	var opts []grpc.DialOption

	if strings.HasSuffix(host, ":443") {
		opts = append(opts, grpc.WithAuthority(host))
		systemRoots, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}

		cred := credentials.NewTLS(&tls.Config{
			RootCAs: systemRoots,
		})
		opts = append(opts, grpc.WithTransportCredentials(cred))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	opts = append(opts, grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))

	grpcConn, err := grpc.DialContext(ctx, host, opts...)
	if err != nil {
		return nil, err
	}

	return grpcConn, nil
}
