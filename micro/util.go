package micro

import (
	"fmt"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/kiga-hub/arc/logging"
)

func zapLoggerEchoMiddleware(logFunc func() logging.ILogger, skipSuccess bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			fields := []interface{}{
				"remote_ip", c.RealIP(),
				"latency", time.Since(start).String(),
				"host", req.Host,
				"request", fmt.Sprintf("%s %s", req.Method, req.RequestURI),
				"status", res.Status,
				"size", res.Size,
				"user_agent", req.UserAgent(),
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
				fields = append(fields, "request_id", id)
			}

			n := res.Status
			switch {
			case n >= 500:
				logFunc().Errorw("Server error", fields...)
			case n >= 400:
				logFunc().Warnw("Client error", fields...)
			case n >= 300:
				logFunc().Debugw("Redirection", fields...)
			default:
				if !skipSuccess {
					logFunc().Debugw("Success", fields...)
				}
			}

			return nil
		}
	}
}

// MonitorStat is the state of the runtime
type MonitorStat struct {
	runtime.MemStats
	LiveObjects  uint64 `json:"live_objects,omitempty"`  // Live objects = Mallocs - Frees
	NumGoroutine int    `json:"num_goroutine,omitempty"` // Number of goroutines
}

// DoMonitor start a loop for monitor
func DoMonitor(duration int, callback func(*MonitorStat)) {
	interval := time.Duration(duration) * time.Second
	timer := time.Tick(interval)
	for range timer {
		var rtm runtime.MemStats
		runtime.ReadMemStats(&rtm)
		callback(&MonitorStat{
			MemStats:     rtm,
			NumGoroutine: runtime.NumGoroutine(),
			LiveObjects:  rtm.Mallocs - rtm.Frees,
		})
	}
}

// GenerateLoggerForModule generate a func for get logger for module w/ logger or logger group
func GenerateLoggerForModule(server *Server, module string) func() logging.ILogger {
	element := server.GetElement(&LoggingElementKey)
	if element != nil {
		l := element.(logging.ILogger)
		return func() logging.ILogger {
			return l
		}
	}
	element = server.GetElement(&LoggerGroupElementKey)
	if element != nil {
		group := element.(*logging.LoggerGroup)
		group.SetLevel(module, "")
		return func() logging.ILogger {
			return group.M(module)
		}
	}
	return func() logging.ILogger {
		return &logging.NoopLogger{}
	}
}
