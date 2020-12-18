package golug_tracing

type tracingLogger struct{}

func (l *tracingLogger) Error(msg string) {
	log.Error(msg)
}

func (l *tracingLogger) Infof(msg string, args ...interface{}) {
	log.Infof(msg, args...)
}
