package tracing

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	proto "github.com/grafana/tempo/pkg/tempopb"
	v1 "github.com/grafana/tempo/pkg/tempopb/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/kiga-hub/arc/logging/loki"
)

// TempoClient is a client for Tempo
type TempoClient struct {
	conn        *grpc.ClientConn
	queryClient proto.QuerierClient
}

// NewTempoClient create a new TempoClient
func NewTempoClient(tempoQueryAddr string) (*TempoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	conn, err := grpc.DialContext(ctx, tempoQueryAddr,
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
	queryClient := proto.NewQuerierClient(conn)
	return &TempoClient{
		conn:        conn,
		queryClient: queryClient,
	}, nil
}

// FindTraceByID find trace by traceID
func (c *TempoClient) FindTraceByID(ctx context.Context, traceID string) ([]*v1.ResourceSpans, error) {
	// similar to http://localhost:33100/api/traces/06460b6bef6c8dfe
	bs, err := c.hexStringToTraceID(traceID)
	if err != nil {
		return nil, err
	}
	fmt.Println(bs)
	resp, err := c.queryClient.FindTraceByID(ctx, &proto.TraceByIDRequest{
		TraceID: bs,
	})
	if err != nil {
		return nil, err
	}
	if resp.Trace == nil || len(resp.Trace.Batches) == 0 {
		return nil, errors.New("trace not found")
	}
	return resp.Trace.Batches, nil
}

// GetTraceIDFromLog get traceIDs from logs query by selector
func GetTraceIDFromLog(ctx context.Context, lokiClient *loki.Client, selector string) ([]string, error) {
	var result []string
	resp, err := lokiClient.GetLogs(ctx, selector, time.Unix(0, 0), time.Now().UTC(), 1)
	if err != nil {
		return result, err
	}
	if len(resp) == 0 {
		return result, nil
	}

	for _, v := range resp {
		traceID, ok := v.LabelSet[TraceIDKey]
		if !ok {
			continue
		}
		result = append(result, string(traceID))
	}
	return result, nil
}

func (c *TempoClient) hexStringToTraceID(id string) ([]byte, error) {
	// the encoding/hex package does not like odd length strings.
	// just append a bit here
	if len(id)%2 == 1 {
		id = "0" + id
	}

	byteID, err := hex.DecodeString(id)
	if err != nil {
		return nil, err
	}

	size := len(byteID)
	if size > 16 {
		return nil, errors.New("trace ids can't be larger than 128 bits")
	}
	if size < 16 {
		byteID = append(make([]byte, 16-size), byteID...)
	}

	return byteID, nil
}
