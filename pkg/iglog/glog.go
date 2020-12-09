package iglog

import (
	"alert"
	"context"
	"fmt"
	"reqid"
	"sync"

	"github.com/golang/glog"
)

type logContextKey struct {
}

var (
	contextLogKey = logContextKey{}
	pool          = sync.Pool{
		New: func() interface{} {
			return &Logger{}
		},
	}
)

func FromContext(ctx context.Context) *Logger {
	dl, ok := ctx.Value(contextLogKey).(*Logger)
	if !ok {
		return newWith(ctx)
	}
	return dl
}

func WithLog(ctx context.Context) context.Context {
	dl := newWith(ctx)
	return context.WithValue(ctx, contextLogKey, dl)
}

type Logger struct {
	format string
}

func New(requestID string) *Logger {
	l := pool.Get().(*Logger)
	l.format = fmt.Sprintf("[%s] ", requestID)
	return l
}

func newWith(ctx context.Context) *Logger {
	if requestID, ok := reqid.FromContext(ctx); ok {
		return New(requestID)
	}
	return New("")
}

func Put(logger *Logger) {
	pool.Put(logger)
}

func (l *Logger) Info(args ...interface{}) {
	glog.InfoDepth(1, l.format, args)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	glog.InfoDepth(1, l.format, fmt.Sprintf(format, args...))
}

func (l *Logger) InfoDepth(depth int, args ...interface{}) {
	glog.InfoDepth(depth+1, l.format, args)
}

func (l *Logger) InfofDepth(depth int, format string, args ...interface{}) {
	glog.InfoDepth(depth+1, l.format, fmt.Sprintf(format, args...))
}

func (l *Logger) Warning(args ...interface{}) {
	glog.WarningDepth(1, l.format, args)
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	glog.WarningDepth(1, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(args ...interface{}) {
	glog.ErrorDepth(1, l.format, args)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	glog.ErrorDepth(1, l.format, fmt.Sprintf(format, args...))
}

func (l *Logger) Alert(args ...interface{}) {
	msg := fmt.Sprintf("[ALERT]" + fmt.Sprint(args...))
	glog.WarningDepth(1, l.format, msg)
	l.alert(l.format + msg)
}

func (l *Logger) Alertf(format string, args ...interface{}) {
	msg := fmt.Sprintf("[ALERT]"+format, args...)
	glog.WarningDepth(1, l.format, msg)
	l.alert(l.format + msg)
}

func (l *Logger) alert(msg string) {
	alert.AsyncSend(msg)
}
