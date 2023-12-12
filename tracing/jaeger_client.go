package tracing

import (
	"context"
	"io"
	"time"

	"github.com/jaegertracing/jaeger/model"
	proto "github.com/jaegertracing/jaeger/proto-gen/api_v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// JaegerClient is a client for query Jaeger
type JaegerClient struct {
	conn        *grpc.ClientConn
	queryClient proto.QueryServiceClient
}

// NewJaegerClient create a new JaegerClient
//
//goland:noinspection GoUnusedExportedFunction
func NewJaegerClient(jaegerQueryAddr string) (*JaegerClient, error) {
	conn, err := grpc.DialContext(context.Background(), jaegerQueryAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             100 * time.Millisecond,
			PermitWithoutStream: true}),
	)
	if err != nil {
		return nil, err
	}
	queryClient := proto.NewQueryServiceClient(conn)
	return &JaegerClient{
		conn:        conn,
		queryClient: queryClient,
	}, nil
}

// QuerySpans query spans from Jaeger
func (c *JaegerClient) QuerySpans(ctx context.Context, query *proto.TraceQueryParameters) ([]model.Span, error) {
	findTraceClient, err := c.queryClient.FindTraces(ctx, &proto.FindTracesRequest{
		Query: query,
	})
	if err != nil {
		return nil, err
	}
	var spans []model.Span
	for {
		spanChunk, err := findTraceClient.Recv()
		if err == io.EOF { // EOF if finish
			break
		} else if err != nil {
			return nil, err
		}
		spans = append(spans, spanChunk.Spans...)
	}
	return spans, nil
}
