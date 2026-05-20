package supervisor

import (
	"errors"
)

// 预定义错误
var (
	// ErrServiceNotFound 服务未找到
	ErrServiceNotFound = errors.New("service not found")
	// ErrServiceAlreadyExists 服务已存在
	ErrServiceAlreadyExists = errors.New("service already exists")
	// ErrServiceStopped 服务已停止
	ErrServiceStopped = errors.New("service is stopped")
	// ErrServiceRunning 服务正在运行
	ErrServiceRunning = errors.New("service is running")
	// ErrDoNotRestart 不要重启服务
	ErrDoNotRestart = errors.New("do not restart")
	// ErrTerminateSupervisor 终止 supervisor
	ErrTerminateSupervisor = errors.New("terminate supervisor")
)

// FatalErr 致命错误，将导致 supervisor 终止
type FatalErr struct {
	Err    error
	Status ExitStatus
}

// AsFatalErr 将错误包装为 FatalErr
func AsFatalErr(err error, status ExitStatus) *FatalErr {
	var fErr *FatalErr
	if errors.As(err, &fErr) {
		return fErr
	}
	return &FatalErr{Err: err, Status: status}
}

// IsFatal 判断是否为致命错误
func IsFatal(err error) bool {
	var fErr *FatalErr
	return errors.As(err, &fErr)
}

func (e *FatalErr) Error() string {
	return e.Err.Error()
}

func (e *FatalErr) Unwrap() error {
	return e.Err
}

func (*FatalErr) Is(target error) bool {
	return target == ErrTerminateSupervisor
}

// NoRestartErr 包装错误，使其不会触发自动重启
func NoRestartErr(err error) error {
	if err == nil {
		return ErrDoNotRestart
	}
	return &noRestartErr{err}
}

// IsNoRestartErr 判断是否是不需要重启的错误
func IsNoRestartErr(err error) bool {
	return errors.Is(err, ErrDoNotRestart)
}

// IsFatalErr 判断是否是致命错误
func IsFatalErr(err error) bool {
	return IsFatal(err) || errors.Is(err, ErrTerminateSupervisor)
}

type noRestartErr struct {
	err error
}

func (e *noRestartErr) Error() string {
	return e.err.Error()
}

func (e *noRestartErr) Unwrap() error {
	return e.err
}

func (*noRestartErr) Is(target error) bool {
	return target == ErrDoNotRestart
}

// ExitStatus 退出状态码
type ExitStatus int

const (
	ExitSuccess            ExitStatus = 0
	ExitError              ExitStatus = 1
	ExitNoUpgradeAvailable ExitStatus = 2
	ExitRestart            ExitStatus = 3
	ExitUpgrade            ExitStatus = 4
)

func (s ExitStatus) AsInt() int {
	return int(s)
}
