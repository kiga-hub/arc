package error

import (
	"testing"
)

// TestNewError 测试NewError
func TestNewError(t *testing.T) {
	// 输出error.go 里面的错误
	t.Log(ErrNoServiceAvailable)
	t.Log(ErrNoServiceAvailable)
	t.Log(ErrNoTaskInTaskResult)
	t.Log(ErrFileProperty)
	t.Log(ErrUnavailableHost)
	t.Log(ErrConfigNotHaveConfig)
	t.Log(ErrLockFail)
	t.Log(ErrTaskNotCompleted)
	t.Log(ErrNoAdapterServer)
	t.Log(ErrUnknownType)
	t.Log(ErrWorkerDidNotReturn)
	t.Log(ErrClientIDInvalid)
	t.Log(ErrTaskTypeInvalid)
	t.Log(ErrAlreadyExit)

}
