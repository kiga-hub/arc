package tracing

import (
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/kiga-hub/arc/logging"
)

// TraceIDKey is a constant value for trace id
const TraceIDKey = "traceID"

// LoggerWithSpan  logger
type LoggerWithSpan struct {
	Span           opentracing.Span
	OriginalLogger logging.ILogger
}

func (l *LoggerWithSpan) getTraceIDKVArgs(keysAndValues []interface{}) []interface{} {
	if l.Span == nil {
		return keysAndValues
	}
	traceID := GetTraceIDFromSpan(l.Span)
	newArgs := make([]interface{}, 0, len(keysAndValues)+2)
	newArgs = append(newArgs, TraceIDKey)
	newArgs = append(newArgs, traceID)
	newArgs = append(newArgs, keysAndValues...)
	return newArgs
}

func (l *LoggerWithSpan) patchMsgBySpan(msg string) string {
	if l.Span == nil {
		return msg
	}
	traceID := GetTraceIDFromSpan(l.Span)
	return fmt.Sprintf("%s=%s | %s", TraceIDKey, traceID, msg)
}

func (l *LoggerWithSpan) patchTemplateBySpan(template string) string {
	return l.patchMsgBySpan(template)
}

func (l *LoggerWithSpan) patchArgsBySpan(args []interface{}) []interface{} {
	if l.Span == nil {
		return args
	}
	traceID := GetTraceIDFromSpan(l.Span)
	//spew.Dump(strSlice)
	newArgs := make([]interface{}, 0, len(args)+1)
	newArgs = append(newArgs, fmt.Sprintf("%s=%s | ", TraceIDKey, traceID))
	newArgs = append(newArgs, args...)
	return newArgs
}

// Debugw  Debug
func (l *LoggerWithSpan) Debugw(msg string, keysAndValues ...interface{}) {
	if l.Span != nil {
		keysAndValues = append(keysAndValues, "level", "debug", "msg", msg)
		l.Span.LogKV(keysAndValues...)
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Debugw(l.patchMsgBySpan(msg), l.getTraceIDKVArgs(keysAndValues)...)
	}
}

// Infow info
func (l *LoggerWithSpan) Infow(msg string, keysAndValues ...interface{}) {
	if l.Span != nil {
		keysAndValues = append(keysAndValues, "level", "info", "msg", msg)
		l.Span.LogKV(keysAndValues...)
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Infow(l.patchMsgBySpan(msg), l.getTraceIDKVArgs(keysAndValues)...)
	}
}

// Warnw warn
func (l *LoggerWithSpan) Warnw(msg string, keysAndValues ...interface{}) {
	if l.Span != nil {
		keysAndValues = append(keysAndValues, "level", "warn", "msg", msg)
		l.Span.LogKV(keysAndValues...)
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Warnw(l.patchMsgBySpan(msg), l.getTraceIDKVArgs(keysAndValues)...)
	}
}

// Errorw  wrong error
func (l *LoggerWithSpan) Errorw(msg string, keysAndValues ...interface{}) {
	if l.Span != nil {
		ext.Error.Set(l.Span, true)
		keysAndValues = append(keysAndValues, "level", "error", "msg", msg)
		l.Span.LogKV(keysAndValues...)
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Errorw(l.patchMsgBySpan(msg), l.getTraceIDKVArgs(keysAndValues)...)
	}
}

// Panicw Panic
func (l *LoggerWithSpan) Panicw(msg string, keysAndValues ...interface{}) {
	if l.Span != nil {
		ext.Error.Set(l.Span, true)
		keysAndValues = append(keysAndValues, "level", "panic", "msg", msg)
		l.Span.LogKV(keysAndValues...)
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Panicw(l.patchMsgBySpan(msg), l.getTraceIDKVArgs(keysAndValues)...)
	}
}

// Fatalw Fatal
func (l *LoggerWithSpan) Fatalw(msg string, keysAndValues ...interface{}) {
	if l.Span != nil {
		ext.Error.Set(l.Span, true)
		keysAndValues = append(keysAndValues, "level", "fatal", "msg", msg)
		l.Span.LogKV(keysAndValues...)
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Fatalw(l.patchMsgBySpan(msg), l.getTraceIDKVArgs(keysAndValues)...)
	}
}

// Panic  Panic
func (l *LoggerWithSpan) Panic(args ...interface{}) {
	if l.Span != nil {
		ext.Error.Set(l.Span, true)
		l.Span.LogKV("level", "panic", "msg", fmt.Sprint(args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Panic(l.patchArgsBySpan(args)...)
	}
}

// Fatal  Fatal
func (l *LoggerWithSpan) Fatal(args ...interface{}) {
	if l.Span != nil {
		ext.Error.Set(l.Span, true)
		l.Span.LogKV("level", "fatal", "msg", fmt.Sprint(args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Fatal(l.patchArgsBySpan(args)...)
	}
}

// Error  Error
func (l *LoggerWithSpan) Error(args ...interface{}) {
	if l.Span != nil {
		ext.Error.Set(l.Span, true)
		l.Span.LogKV("level", "error", "msg", fmt.Sprint(args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Error(l.patchArgsBySpan(args)...)
	}
}

// Debug Debug
func (l *LoggerWithSpan) Debug(args ...interface{}) {
	if l.Span != nil {
		l.Span.LogKV("level", "debug", "msg", fmt.Sprint(args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Debug(l.patchArgsBySpan(args)...)
	}
}

// Info Info
func (l *LoggerWithSpan) Info(args ...interface{}) {
	if l.Span != nil {
		l.Span.LogKV("level", "info", "msg", fmt.Sprint(args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Info(l.patchArgsBySpan(args)...)
	}
}

// Warn Warning
func (l *LoggerWithSpan) Warn(args ...interface{}) {
	if l.Span != nil {
		l.Span.LogKV("level", "warn", "msg", fmt.Sprint(args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Warn(l.patchArgsBySpan(args)...)
	}
}

// Errorf Errorf
func (l *LoggerWithSpan) Errorf(template string, args ...interface{}) {
	if l.Span != nil {
		ext.Error.Set(l.Span, true)
		l.Span.LogKV("level", "error", "msg", fmt.Sprintf(template, args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Errorf(l.patchTemplateBySpan(template), args...)
	}
}

// Debugf Debugf
func (l *LoggerWithSpan) Debugf(template string, args ...interface{}) {
	if l.Span != nil {
		l.Span.LogKV("level", "debug", "msg", fmt.Sprintf(template, args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Debugf(l.patchTemplateBySpan(template), args...)
	}
}

// Infof Info
func (l *LoggerWithSpan) Infof(template string, args ...interface{}) {
	if l.Span != nil {
		l.Span.LogKV("level", "info", "msg", fmt.Sprintf(template, args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Infof(l.patchTemplateBySpan(template), args...)
	}
}

// Warnf Warn
func (l *LoggerWithSpan) Warnf(template string, args ...interface{}) {
	if l.Span != nil {
		l.Span.LogKV("level", "warn", "msg", fmt.Sprintf(template, args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Warnf(l.patchTemplateBySpan(template), args...)
	}
}

// Fatalf Fatal
func (l *LoggerWithSpan) Fatalf(template string, args ...interface{}) {
	if l.Span != nil {
		l.Span.LogKV("level", "fatal", "msg", fmt.Sprintf(template, args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Fatalf(l.patchTemplateBySpan(template), args...)
	}
}

// Panicf Panic
func (l *LoggerWithSpan) Panicf(template string, args ...interface{}) {
	if l.Span != nil {
		l.Span.LogKV("level", "panic", "msg", fmt.Sprintf(template, args...))
	}
	if l.OriginalLogger != nil {
		l.OriginalLogger.Panicf(l.patchTemplateBySpan(template), args...)
	}
}
