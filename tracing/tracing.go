package tracing

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	jaegerZap "github.com/uber/jaeger-client-go/log/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"github.com/kiga-hub/arc/logging"
	microConf "github.com/kiga-hub/arc/micro/conf"
)

const (
	skipSpanType = "opentracing.noopSpan"
)

// SetupGlobleTracer 设置这个整体示踪剂
//
//goland:noinspection GoUnusedExportedFunction
func SetupGlobleTracer(basic microConf.BasicConfig, trace TraceConfig, zlog *zap.Logger) (opentracing.Tracer, io.Closer, error) {
	tracer, closer, err := CreateTracer(basic, trace, zlog)
	if err == nil {
		opentracing.SetGlobalTracer(tracer)
	}
	return tracer, closer, err
}

// CreateTracer 创建 tracer
func CreateTracer(basic microConf.BasicConfig, trace TraceConfig, zlog *zap.Logger) (opentracing.Tracer, io.Closer, error) {
	// tracing
	traceCfg := jaegerConfig.Configuration{
		ServiceName: basic.Service,
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			CollectorEndpoint:   fmt.Sprintf("%s/api/traces", trace.JaegerCollectorAddr),
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	opts := []jaegerConfig.Option{
		jaegerConfig.Tag("zone", basic.Zone),
		jaegerConfig.Tag("node", basic.Node),
		jaegerConfig.Tag("machine", basic.Machine),
		jaegerConfig.Tag("instance", basic.Instance),
		jaegerConfig.Tag("service", basic.Service),
		jaegerConfig.Tag("appname", basic.AppName),
		jaegerConfig.Tag("appversion", basic.AppVersion),
	}
	if zlog != nil {
		opts = append(opts, jaegerConfig.Logger(jaegerZap.NewLogger(zlog)))
	}
	return traceCfg.NewTracer(opts...)
}

// MarshalSpanToJSON  解码span 到json
//
//goland:noinspection GoUnusedExportedFunction
func MarshalSpanToJSON(span opentracing.Span) (string, error) {
	if span == nil {
		return "", errors.New("nil span")
	}
	tc := opentracing.TextMapCarrier{}
	err := span.Tracer().Inject(span.Context(), opentracing.TextMap, tc)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(tc)
	if err != nil {
		return "", err
	}
	return string(b[:]), nil
}

// UnmarshalJSONToCarrier unmarshal a JSON string to an opentracing.TextMapCarrier object
func UnmarshalJSONToCarrier(marshaled string) (opentracing.TextMapCarrier, error) {
	tc := opentracing.TextMapCarrier{}
	err := json.Unmarshal([]byte(marshaled), &tc)
	return tc, err
}

// StartChildSpanFromJSON   开启子任务 Span 来自 json
//
//goland:noinspection GoUnusedExportedFunction
func StartChildSpanFromJSON(tracer opentracing.Tracer, operaterName, marshaled string) (opentracing.Span, error) {
	if tracer == nil {
		return nil, errors.New("tracer not set")
	}
	tc, err := UnmarshalJSONToCarrier(marshaled)
	if err != nil {
		return nil, err
	}
	ctx, err := tracer.Extract(opentracing.TextMap, tc)
	if err != nil {
		return nil, err
	}
	span := tracer.StartSpan(
		operaterName,
		opentracing.ChildOf(ctx),
	)
	return span, nil
}

// ContinueSpanFromJSON ContinueSpanFromJsons
//
//goland:noinspection GoUnusedExportedFunction
func ContinueSpanFromJSON(tracer opentracing.Tracer, operaterName, marshaled string) (opentracing.Span, error) {
	if tracer == nil {
		return nil, errors.New("tracer not set")
	}
	tc, err := UnmarshalJSONToCarrier(marshaled)
	if err != nil {
		return nil, err
	}
	ctx, err := tracer.Extract(opentracing.TextMap, tc)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		return nil, errors.New("no ctx found")
	}
	span := tracer.StartSpan(
		operaterName,
		opentracing.FollowsFrom(ctx),
	)
	return span, nil
}

// GetTraceIDFromSpan get traceID from a Span
func GetTraceIDFromSpan(span opentracing.Span) string {
	if span == nil {
		return ""
	}
	strs := strings.Split(fmt.Sprintf("%+v", span), ":")
	return strs[0]
}

// GetTraceIDFromJSON get traceID from a JSON string
//
//goland:noinspection GoUnusedExportedFunction
func GetTraceIDFromJSON(marshaled string) string {
	tc, err := UnmarshalJSONToCarrier(marshaled)
	if err != nil {
		return ""
	}
	span, ok := tc["uber-trace-id"]
	if !ok {
		return ""
	}
	strs := strings.Split(fmt.Sprintf("%+v", span), ":")
	return strs[0]
}

// ErrorToSpan  ErrorTo Span
//
//goland:noinspection GoUnusedExportedFunction
func ErrorToSpan(span opentracing.Span, err error) {
	if span == nil || err == nil {
		return
	}
	ext.Error.Set(span, true)
	span.LogKV("level", "error", "msg", err.Error())
}

var (
	emptyFinishSpan = func() {}
)

// StartChildSpan Start ChildSpan
func StartChildSpan(
	ctx context.Context,
	logger logging.ILogger,
	tracer opentracing.Tracer,
	operaterName string,
	tags ...opentracing.StartSpanOption,
) (context.Context, func(), logging.ILogger) { // ctx, finish(), logger
	if tracer == nil {
		return ctx, emptyFinishSpan, logger
	}
	p := opentracing.SpanFromContext(ctx)
	if p == nil {
		return ctx, emptyFinishSpan, logger
	}
	// machinery will bring a noopSpan in the context if no other span set
	pt := reflect.TypeOf(p)
	if pt.String() == skipSpanType {
		return ctx, emptyFinishSpan, logger
	}

	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, operaterName, tags...)
	return ctx, span.Finish, &LoggerWithSpan{
		Span:           span,
		OriginalLogger: logger,
	}
}

var (
	//SpanKind "span.kind"
	SpanKind = "span.kind"
	//TraceKeyComponent "component"
	TraceKeyComponent = "component"
	//grpcTag
	grpcTag = opentracing.Tag{Key: TraceKeyComponent, Value: "gRPC"}
	//SpanKindClient  客户端
	SpanKindClient = opentracing.Tag{Key: SpanKind, Value: "client"}
	//SpanKindServer 服务端
	SpanKindServer = opentracing.Tag{Key: SpanKind, Value: "server"}
	//SpanKindPortal  接口
	SpanKindPortal = opentracing.Tag{Key: SpanKind, Value: "portal"}
	//SpanKindProducer 生产者
	SpanKindProducer = opentracing.Tag{Key: SpanKind, Value: "producer"}
	//SpanKindConsumer 消费者
	SpanKindConsumer = opentracing.Tag{Key: SpanKind, Value: "consumer"}
	//SpanKindWorker 工作
	SpanKindWorker = opentracing.Tag{Key: SpanKind, Value: "worker"}
	//SpanKindComputation 计算
	SpanKindComputation = opentracing.Tag{Key: SpanKind, Value: "computation"}
	//SpanKindDB 数据库
	SpanKindDB = opentracing.Tag{Key: SpanKind, Value: "database"}
)

type clientSpanTagKey struct{}

// GetGRPCClientSpan 获取grpc客户端span
//
//goland:noinspection GoUnusedExportedFunction
func GetGRPCClientSpan(
	ctx context.Context,
	logger logging.ILogger,
	tracer opentracing.Tracer,
	operationName string,
	mustFindParent bool,
) (context.Context, func(), logging.ILogger) { // ctx, finish(), logger
	if tracer == nil {
		return ctx, emptyFinishSpan, logger
	}
	parent := opentracing.SpanFromContext(ctx)
	// machinery will bring a noopSpan in the context if no other span set
	pt := reflect.TypeOf(parent)
	if pt.String() == skipSpanType {
		parent = nil
	}
	//fmt.Println(pt.String())
	if mustFindParent && parent == nil {
		return ctx, emptyFinishSpan, logger
	}

	opts := []opentracing.StartSpanOption{
		ext.SpanKindRPCClient,
		grpcTag,
	}
	if parent != nil {
		opts = append(opts, opentracing.ChildOf(parent.Context()))
	}
	if tagx := ctx.Value(clientSpanTagKey{}); tagx != nil {
		if opt, ok := tagx.(opentracing.StartSpanOption); ok {
			opts = append(opts, opt)
		}
	}
	span := tracer.StartSpan(operationName, opts...)
	// Make sure we add this to the metadata of the call, so it gets propagated:
	md := metautils.ExtractOutgoing(ctx).Clone()
	if err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, metadataTextMap(md)); err != nil {
		logger.Error(err)
		return ctx, emptyFinishSpan, logger
	}
	ctxWithMetadata := md.ToOutgoing(ctx)
	return opentracing.ContextWithSpan(ctxWithMetadata, span), span.Finish, &LoggerWithSpan{
		Span:           span,
		OriginalLogger: logger,
	}
}

/*
// MicroServerTraceWrapper MicroServer TraceWrapper
func MicroServerTraceWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		tracer := opentracing.GlobalTracer()
		operationName := req.Method()

		if tracer == nil || req.Stream() || operationName == "Health.Check" {
			return fn(ctx, req, rsp)
		}

		//extract metadata to context
		md := metautils.ExtractIncoming(ctx).Clone()
		parentSpanContext, err := tracer.Extract(opentracing.HTTPHeaders, metadataTextMap(md))
		if err != nil {
			if err == opentracing.ErrSpanContextNotFound {
				return fn(ctx, req, rsp)
			}
			grpclog.Infof("grpc_opentracing: failed parsing trace information: %v", err)

		}
		if parentSpanContext == nil {
			return fn(ctx, req, rsp)
		}
		span := tracer.StartSpan(
			operationName,
			opentracing.ChildOf(parentSpanContext),
			SpanKindServer,
			grpcTag,
		)
		defer span.Finish()

		ctx = opentracing.ContextWithSpan(ctx, span)
		// ctx = context.WithValue(ctx, NeedSample, true)
		return fn(ctx, req, rsp)
	}
}
*/

const (
	binHdrSuffix = "-bin"
)

// metadataTextMap extends a metadata.MD to be an opentracing textmap
type metadataTextMap metadata.MD

// Set is an opentracing.TextMapReader interface that extracts values.
func (m metadataTextMap) Set(key, val string) {
	// gRPC allows for complex binary values to be written.
	encodedKey, encodedVal := encodeKeyValue(key, val)
	// The metadata object is a mul-map, and previous values may exist, but for opentracing headers, we do not append
	// we just override.
	m[encodedKey] = []string{encodedVal}
}

// ForeachKey is an opentracing.TextMapReader interface that extracts values.
func (m metadataTextMap) ForeachKey(callback func(key, val string) error) error {
	for k, vv := range m {
		for _, v := range vv {
			if err := callback(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// encodeKeyValue encodes key and value qualified for transmission via gRPC.
// note: copied pasted from private values of grpc.metadata
func encodeKeyValue(k, v string) (string, string) {
	k = strings.ToLower(k)
	if strings.HasSuffix(k, binHdrSuffix) {
		v = base64.StdEncoding.EncodeToString([]byte(v))

	}
	return k, v
}
