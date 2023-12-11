package logging

//NoopLogger  日志记录器
type NoopLogger struct{}

//Print 打印
func (l *NoopLogger) Print(...interface{}) {}

//Printf 格式化打印
func (l *NoopLogger) Printf(string, ...interface{}) {}

//Println 换行打印
func (l *NoopLogger) Println(...interface{}) {}

//Fatal 打印
func (l *NoopLogger) Fatal(...interface{}) {}

//Fatalf 打印
func (l *NoopLogger) Fatalf(string, ...interface{}) {}

//Fatalln 打印
func (l *NoopLogger) Fatalln(...interface{}) {}

//Panic 打印
func (l *NoopLogger) Panic(...interface{}) {}

//Panicf 打印
func (l *NoopLogger) Panicf(string, ...interface{}) {}

//Panicln 打印
func (l *NoopLogger) Panicln(...interface{}) {}

// Debugw -
func (l *NoopLogger) Debugw(string, ...interface{}) {}

// Infow -
func (l *NoopLogger) Infow(string, ...interface{}) {}

// Warnw -
func (l *NoopLogger) Warnw(string, ...interface{}) {}

// Errorw -
func (l *NoopLogger) Errorw(string, ...interface{}) {}

// Panicw -
func (l *NoopLogger) Panicw(string, ...interface{}) {}

// Fatalw -
func (l *NoopLogger) Fatalw(string, ...interface{}) {}

// Error -
func (l *NoopLogger) Error(args ...interface{}) {}

// Debug -
func (l *NoopLogger) Debug(args ...interface{}) {}

// Info -
func (l *NoopLogger) Info(args ...interface{}) {}

// Warn -
func (l *NoopLogger) Warn(args ...interface{}) {}

// Debugf -
func (l *NoopLogger) Debugf(string, ...interface{}) {}

// Infof -
func (l *NoopLogger) Infof(string, ...interface{}) {}

// Warnf -
func (l *NoopLogger) Warnf(string, ...interface{}) {}

// Errorf -
func (l *NoopLogger) Errorf(string, ...interface{}) {}
