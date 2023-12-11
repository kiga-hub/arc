package machinery

import (
	"fmt"

	"common/logging"
)

// GetQueueNameByCluster ...
func GetQueueNameByCluster(taskType, cluster string) string {
	return fmt.Sprintf("%s_%s", taskType, cluster)
}

// Logger Logger
type Logger struct {
	logger logging.ILogger
}

// NewMachineryLogger -
func NewMachineryLogger(logger logging.ILogger) *Logger {
	return &Logger{
		logger: logger,
	}
}

// Print -
func (m *Logger) Print(args ...interface{}) {
	m.logger.Info(args)
}

// Printf -
func (m *Logger) Printf(str string, args ...interface{}) {
	m.logger.Infof(str, args)
}

// Println -
func (m *Logger) Println(args ...interface{}) {
	m.logger.Info(args)
}

// Fatal -
func (m *Logger) Fatal(args ...interface{}) {
	m.logger.Fatal(args)
}

// Fatalf -
func (m *Logger) Fatalf(str string, args ...interface{}) {
	m.logger.Fatalf(str, args)
}

// Fatalln -
func (m *Logger) Fatalln(args ...interface{}) {
	m.logger.Fatal(args)
}

// Panic -
func (m *Logger) Panic(args ...interface{}) {
	m.logger.Panic(args)
}

// Panicf -
func (m *Logger) Panicf(str string, args ...interface{}) {
	m.logger.Panicf(str, args)
}

// Panicln -
func (m *Logger) Panicln(args ...interface{}) {
	m.logger.Panic(args)
}
