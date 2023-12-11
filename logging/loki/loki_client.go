package loki

import (
	"context"
	io "io"
	"time"

	"github.com/grafana/loki/pkg/logproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Client is a client of loki
type Client struct {
	conn         *grpc.ClientConn
	pusherClient logproto.PusherClient
	queryClient  logproto.QuerierClient
}

// NewLokiClient create a new LokiClient
func NewLokiClient(lokiAddr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	conn, err := grpc.DialContext(ctx, lokiAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true}),
	)
	if err != nil {
		return nil, err
	}
	pusherClient := logproto.NewPusherClient(conn)
	queryClient := logproto.NewQuerierClient(conn)
	return &Client{
		conn:         conn,
		pusherClient: pusherClient,
		queryClient:  queryClient,
	}, nil
}

// SendLogs send logs to loki
func (c Client) SendLogs(ctx context.Context, request *logproto.PushRequest) error {
	_, err := c.pusherClient.Push(ctx, request)
	return err
}

// GetLogs get logs from loki
func (c Client) GetLogs(ctx context.Context, selector string, start, end time.Time, limit uint32) ([]StreamWithLables, error) {
	client, err := c.queryClient.Query(ctx, &logproto.QueryRequest{
		Selector: selector,
		Start:    start,
		End:      end,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}
	streams := []StreamWithLables{}
	for {
		resp, err := client.Recv()
		if err == io.EOF { // EOF if finish
			break
		} else if err != nil {
			return nil, err
		}
		for _, v := range resp.Streams {
			streams = append(streams, getStreamWithLables(v))
		}
	}
	return streams, nil
}
