package machinery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RichardKnop/machinery/v1/log"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/silenceper/pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kiga-hub/common/logging"
)

const (
	//POOLCOUNT 连接池数量
	POOLCOUNT = 2
	//POOLTYPEBIZZ 连接池类型昵称
	POOLTYPEBIZZ = 0
	//POOLTYPEHEALTH 连接池类型健康状态
	POOLTYPEHEALTH = 1
)

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

// GetPool 获取连接池
func GetPool(target string, min, max int) (pool.Pool, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
	}

	//factory 创建连接的方法
	// factory := func() (interface{}, error) { return net.Dial("tcp", addr) }
	factory := func() (interface{}, error) { //*grpc.ClientConn
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		return grpc.DialContext(ctx, target, opts...)
	}

	//close 关闭连接的方法
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

// GetConn 获取链接
func GetConn(p pool.Pool, timeout int) (*grpc.ClientConn, error) {
	skipTime := time.Now().Add(time.Second * time.Duration(timeout))

	for {
		if time.Now().After(skipTime) {
			return nil, errors.New("get connection timeout")
		}
		iface, err := p.Get()
		if err == pool.ErrMaxActiveConnReached {
			fmt.Println(200)
			time.Sleep(time.Millisecond * 50)
			continue
		}

		if err != nil {
			return nil, err
		}

		conn := iface.(*grpc.ClientConn)
		return conn, nil
	}
}

// ReturnConn 返回链接
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

// SetLoggerLevel 设置日志级别
/* e.g.
if level.Level() > zapcore.DebugLevel {
	machinery.SetLoggerLevel(false)
}
*/
func SetLoggerLevel(logger logging.ILogger) {
	l := NewMachineryLogger(logger)
	log.Set(l)
}
