package machinery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RichardKnop/machinery/v1/log"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/silenceper/pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kiga-hub/arc/logging"
)

const (
	// PoolCount pool count
	PoolCount int = 2
	// PoolTypeName pool type name
	PoolTypeName = 0
	// PoolTypeHealth pool type health
	PoolTypeHealth = 1
)

func _() {
	fmt.Println("Pool count: ", PoolCount)
	fmt.Println("Pool type name: ", PoolTypeName)
	fmt.Println("Pool type health: ", PoolTypeHealth)
}

/*
// If it returns false, the given request will not be traced.
func filter(ctx context.Context, fullMethodName string) bool {
	if fullMethodName == "/grpc.health.v1.Health/Check" ||
		fullMethodName == "/grpc.health.v1.Health/Watch" {
		return false
	}
	return true
}
*/

// GetPool get pool
//
//goland:noinspection GoUnusedExportedFunction
func GetPool(target string, min, max int) (pool.Pool, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(grpcPrometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpcPrometheus.StreamClientInterceptor),
	}

	//factory create new connection
	// factory := func() (interface{}, error) { return net.Dial("tcp", addr) }
	factory := func() (interface{}, error) { //*grpc.ClientConn
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		return grpc.DialContext(ctx, target, opts...)
	}

	//close close connection
	// close := func() error { return v.(net.Conn).Close() }
	clientConnClose := func(v interface{}) error { return v.(*grpc.ClientConn).Close() }

	poolConfig := &pool.Config{
		InitialCap:  min,
		MaxIdle:     max,
		MaxCap:      max,
		Factory:     factory,
		Close:       clientConnClose,
		IdleTimeout: 60 * time.Second,
	}
	//spew.Dump(poolConfig)
	return pool.NewChannelPool(poolConfig)
}

// GetConn get connection
//
//goland:noinspection GoUnusedExportedFunction
func GetConn(p pool.Pool, timeout int) (*grpc.ClientConn, error) {
	skipTime := time.Now().Add(time.Second * time.Duration(timeout))

	for {
		if time.Now().After(skipTime) {
			return nil, errors.New("get connection timeout")
		}
		iFace, err := p.Get()
		if err == pool.ErrMaxActiveConnReached {
			fmt.Println(200)
			time.Sleep(time.Millisecond * 50)
			continue
		}

		if err != nil {
			return nil, err
		}

		conn := iFace.(*grpc.ClientConn)
		return conn, nil
	}
}

// ReturnConn return connection
//
//goland:noinspection GoUnusedExportedFunction
func ReturnConn(p pool.Pool, conn *grpc.ClientConn, err error) error {
	if p == nil || conn == nil {
		return nil
	}
	if err == nil {
		//fmt.Println(999)
		return p.Put(conn)
	}
	//fmt.Println(666)
	return p.Close(conn)

}

// SetLoggerLevel set logger level
/* e.g.
if level.Level() > zapCore.DebugLevel {
	machinery.SetLoggerLevel(false)
}
*/
//goland:noinspection GoUnusedExportedFunction
func SetLoggerLevel(logger logging.ILogger) {
	l := NewMachineryLogger(logger)
	log.Set(l)
}
