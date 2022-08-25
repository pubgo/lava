package logconfig

import (
	"net/url"
	"strings"
	"time"

	"github.com/pubgo/funk/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultKey = "default"

var (
	levelEncoder = map[string]zapcore.LevelEncoder{
		"capital":      zapcore.CapitalLevelEncoder,
		"capitalColor": zapcore.CapitalColorLevelEncoder,
		"color":        zapcore.LowercaseColorLevelEncoder,
		defaultKey:     zapcore.LowercaseLevelEncoder,
	}
	timeEncoder = map[string]zapcore.TimeEncoder{
		"rfc3339": _RFC3339MilliTimeEncoder,
		"RFC3339": _RFC3339MilliTimeEncoder,
		//"rfc3339":     zapcore.RFC3339TimeEncoder,
		//"RFC3339":     zapcore.RFC3339TimeEncoder,
		"rfc3339nano": zapcore.RFC3339NanoTimeEncoder,
		"RFC3339Nano": zapcore.RFC3339NanoTimeEncoder,
		"iso8601":     zapcore.ISO8601TimeEncoder,
		"ISO8601":     zapcore.ISO8601TimeEncoder,
		"millis":      zapcore.EpochMillisTimeEncoder,
		"nanos":       zapcore.EpochNanosTimeEncoder,
		defaultKey:    zapcore.EpochTimeEncoder,
	}
	durationEncoder = map[string]zapcore.DurationEncoder{
		"string":   zapcore.StringDurationEncoder,
		"nanos":    zapcore.NanosDurationEncoder,
		defaultKey: zapcore.SecondsDurationEncoder,
	}
	callerEncoder = map[string]zapcore.CallerEncoder{
		"full":     zapcore.FullCallerEncoder,
		defaultKey: zapcore.ShortCallerEncoder,
	}
	nameEncoder = map[string]zapcore.NameEncoder{
		"full":     zapcore.FullNameEncoder,
		defaultKey: zapcore.FullNameEncoder,
	}
	encoderNameToConstructor = map[string]func(zapcore.EncoderConfig) (zapcore.Encoder, error){
		"console": func(encoderConfig zapcore.EncoderConfig) (zapcore.Encoder, error) {
			return zapcore.NewConsoleEncoder(encoderConfig), nil
		},
		"json": func(encoderConfig zapcore.EncoderConfig) (zapcore.Encoder, error) {
			return zapcore.NewJSONEncoder(encoderConfig), nil
		},
	}
	sinkFactories = map[string]func(*url.URL) (zap.Sink, error){
		//"rotate": newRotateSink,
		//"file":   newFileSink,
	}
)

func init() {
	for k, v := range encoderNameToConstructor {
		if err := zap.RegisterEncoder(k, v); err != nil {
			if !strings.Contains(err.Error(), "already registered") {
				assert.Must(err)
			}
		}
	}

	for k, v := range sinkFactories {
		if err := zap.RegisterSink(k, v); err != nil {
			if !strings.Contains(err.Error(), "already registered") {
				assert.Must(err)
			}
		}
	}
}

func _RFC3339MilliTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, "2006-01-02T15:04:05.000Z07:00")
		return
	}

	enc.AppendString(t.Format("2006-01-02T15:04:05.000Z07:00"))
}

//func getWriter(outputPath string, rotateTime string, maxAge int64, name string, cfg *rotateCfg) (writer io.Writer, err error) {
//	var logName = fmt.Sprintf("%s.%s.%s", name, cfg.Format, cfg.Ext)
//	outputPath = outputPath + string(os.PathSeparator)
//	rotateDuration, err := time.ParseDuration(rotateTime)
//	writer, err = rotatelogs.New(filepath.Join(outputPath, logName),
//		rotatelogs.WithRotationTime(rotateDuration), rotatelogs.WithMaxAge(time.Duration(maxAge)*rotateDuration),
//		rotatelogs.WithLinkName(filepath.Join(outputPath, fmt.Sprintf("%s.%s", name, cfg.Ext))))
//	return
//}

// http://xlog.api/log
//func newHttpSink(u *url.URL) (zap.Sink, error) {
//	if u.User != nil {
//		return nil, fmt.Errorf("user and password not allowed with file URLs: got %v", u)
//	}
//	if u.Fragment != "" {
//		return nil, fmt.Errorf("fragments not allowed with file URLs: got %v", u)
//	}
//	if u.RawQuery != "" {
//		return nil, fmt.Errorf("query parameters not allowed with file URLs: got %v", u)
//	}
//	// Error messages are better if we check hostname and port separately.
//	if u.Port() != "" {
//		return nil, fmt.Errorf("ports not allowed with file URLs: got %v", u)
//	}
//	if hn := u.Hostname(); hn != "" && hn != "localhost" {
//		return nil, fmt.Errorf("file URLs must leave host empty or use localhost: got %v", u)
//	}
//
//	query := u.Query()
//	var cfg = rotate.NewWriterConfig()
//	for k := range query {
//		v := query.Value(k)
//		switch k {
//		case "dir":
//			cfg.Dir = v
//		case "sub":
//			cfg.Sub = v
//		case "name":
//			cfg.Filename = v
//		case "age":
//			cfg.Age = xerror.PanicErr(time.ParseDuration(v)).(time.Duration)
//		case "dur":
//			cfg.Duration = xerror.PanicErr(time.ParseDuration(v)).(time.Duration)
//		case "pattern":
//			cfg.Pattern = v
//		case "count":
//			cfg.Count = uint(xerror.PanicErr(strconv.Atoi(v)).(int))
//		}
//	}
//
//	w, err := rotate.NewRotateLogger(cfg)
//	return &nopCloserSink{zapcore.AddSync(w)}, err
//}

// rotate:///logPath
//func newRotateSink(u *url.URL) (zap.Sink, error) {
//	if u.User != nil {
//		return nil, fmt.Errorf("user and password not allowed with file URLs: got %v", u)
//	}
//	if u.Fragment != "" {
//		return nil, fmt.Errorf("fragments not allowed with file URLs: got %v", u)
//	}
//	if u.RawQuery != "" {
//		return nil, fmt.Errorf("query parameters not allowed with file URLs: got %v", u)
//	}
//	// Error messages are better if we check hostname and port separately.
//	if u.Port() != "" {
//		return nil, fmt.Errorf("ports not allowed with file URLs: got %v", u)
//	}
//	if hn := u.Hostname(); hn != "" && hn != "localhost" {
//		return nil, fmt.Errorf("file URLs must leave host empty or use localhost: got %v", u)
//	}
//
//	query := u.Query()
//	var cfg = rotate.NewWriterConfig()
//	for k := range query {
//		v := query.Value(k)
//		switch k {
//		case "dir":
//			cfg.Dir = v
//		case "sub":
//			cfg.Sub = v
//		case "name":
//			cfg.Filename = v
//		case "age":
//			cfg.Age = xerror.PanicErr(time.ParseDuration(v)).(time.Duration)
//		case "dur":
//			cfg.Duration = xerror.PanicErr(time.ParseDuration(v)).(time.Duration)
//		case "pattern":
//			cfg.Pattern = v
//		case "count":
//			cfg.Count = uint(xerror.PanicErr(strconv.Atoi(v)).(int))
//		}
//	}
//
//	w, err := rotate.NewRotateLogger(cfg)
//	return &nopCloserSink{zapcore.AddSync(w)}, err
//}

//func newFileSink(u *url.URL) (_ zap.Sink, err error) {
//	defer xerror.RespErr(&err)
//
//	xerror.Assert(u.User != nil, "user and password not allowed with file URLs: got %v", u)
//	xerror.Assert(u.Fragment != "", "fragments not allowed with file URLs: got %v", u)
//	xerror.Assert(u.RawQuery != "", "query parameters not allowed with file URLs: got %v", u)
//
//	// Error messages are better if we check hostname and port separately.
//	xerror.Assert(u.Port() != "", "ports not allowed with file URLs: got %v", u)
//
//	hn := u.Hostname()
//	xerror.Assert(hn != "" && hn != "localhost", "file URLs must leave host empty or use localhost: got %v", u)
//
//	switch u.Path {
//	case "stdout":
//		return nopCloserSink{os.Stdout}, nil
//	case "stderr":
//		return nopCloserSink{os.Stderr}, nil
//	}
//
//	u.Path = os.ExpandEnv(u.Path)
//	return os.OpenFile(u.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
//}
