package lib

import (
	"fmt"
	dlog "initkit/log"
	"strings"
)

// 通用DLTag常量定义
const (
	DLTagUndefind      = "_undef"
	DLTagMySqlFailed   = "_com_mysql_failure"
	DLTagRedisFailed   = "_com_redis_failure"
	DLTagMySqlSuccess  = "_com_mysql_success"
	DLTagRedisSuccess  = "_com_redis_success"
	DLTagThriftFailed  = "_com_thrift_failure"
	DLTagThriftSuccess = "_com_thrift_success"
	DLTagHTTPSuccess   = "_com_http_success"
	DLTagHTTPFailed    = "_com_http_failure"
	DLTagTCPFailed     = "_com_tcp_failure"
	DLTagRequestIn     = "_com_request_in"
	DLTagRequestOut    = "_com_request_out"
)

const (
	_dlTag          = "dltag"
	_traceId        = "traceid"
	_spanId         = "spanid"
	_childSpanId    = "cspanid"
	_dlTagBizPrefix = "_com_"
	_dlTagBizUndef  = "_com_undef"
)

type LoggerType interface {
	TagInfo(trace *TraceContext, dltag string, m map[string]interface{})
	TagWarn(trace *TraceContext, dltag string, m map[string]interface{})
	TagError(trace *TraceContext, dltag string, m map[string]interface{})
	TagTrace(trace *TraceContext, dltag string, m map[string]interface{})
	TagDebug(trace *TraceContext, dltag string, m map[string]interface{})
	Close()
}

var Log LoggerType

type Logger struct {
}

func (l *Logger) TagInfo(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = checkDLTag(dltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	dlog.Info(parseParams(m))
}

func (l *Logger) TagWarn(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = checkDLTag(dltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	dlog.Warn(parseParams(m))
}

func (l *Logger) TagError(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = checkDLTag(dltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	dlog.Error(parseParams(m))
}

func (l *Logger) TagTrace(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = checkDLTag(dltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	dlog.Trace(parseParams(m))
}

func (l *Logger) TagDebug(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = checkDLTag(dltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	dlog.Debug(parseParams(m))
}

func (l *Logger) Close() {
	dlog.Close()
}

func NewLogger() LoggerType {
	return &Logger{}
}

// 生成业务dltag
func CreateBizDLTag(tagName string) string {
	if tagName == "" {
		return _dlTagBizUndef
	}

	return _dlTagBizPrefix + tagName
}

// 校验dltag合法性
func checkDLTag(dltag string) string {
	if strings.HasPrefix(dltag, _dlTagBizPrefix) {
		return dltag
	}

	if strings.HasPrefix(dltag, "_com_") {
		return dltag
	}

	if dltag == DLTagUndefind {
		return dltag
	}
	return dltag
}

// map格式化为string
func parseParams(m map[string]interface{}) string {
	var dltag string = "_undef"
	if _dltag, _have := m["dltag"]; _have {
		if __val, __ok := _dltag.(string); __ok {
			dltag = __val
		}
	}
	for _key, _val := range m {
		if _key == "dltag" {
			continue
		}
		dltag = dltag + "||" + fmt.Sprintf("%v=%+v", _key, _val)
	}
	dltag = strings.Trim(fmt.Sprintf("%q", dltag), "\"")
	return dltag
}

type logConf struct {
	LogLevel   string `mapstructure:"log_level"`
	FileWriter struct {
		On              bool   `mapstructure:"on"`
		LogPath         string `mapstructure:"log_path"`
		RotateLogPath   string `mapstructure:"rotate_log_path"`
		WfLogPath       string `mapstructure:"wf_log_path"`
		RotateWfLogPath string `mapstructure:"rotate_wf_log_path"`
	} `mapstructure:"file_writer"`
	ConsoleWriter struct {
		On    bool `mapstructure:"on"`
		Color bool `mapstructure:"color"`
	} `mapstructure:"console_writer"`
}

func initLogger() error {
	conf := &logConf{
		LogLevel: "trace",
		FileWriter: struct {
			On              bool   `mapstructure:"on"`
			LogPath         string `mapstructure:"log_path"`
			RotateLogPath   string `mapstructure:"rotate_log_path"`
			WfLogPath       string `mapstructure:"wf_log_path"`
			RotateWfLogPath string `mapstructure:"rotate_wf_log_path"`
		}{
			On:              true,
			LogPath:         "./logs/gin_scaffold.inf.log",
			RotateLogPath:   "./logs/gin_scaffold.inf.log.%Y%M%D%H",
			WfLogPath:       "./logs/gin_scaffold.wf.log",
			RotateWfLogPath: "./logs/gin_scaffold.wf.log.%Y%M%D%H",
		},
		ConsoleWriter: struct {
			On    bool `mapstructure:"on"`
			Color bool `mapstructure:"color"`
		}{
			On:    false,
			Color: false,
		},
	}
	if IsSetConf("log") {
		err := viperConf.UnmarshalKey("log", &conf)
		if err != nil {
			return err
		}
	}
	//配置日志
	logConf := dlog.LogConfig{
		Level: conf.LogLevel,
		FW: dlog.ConfFileWriter{
			On:              conf.FileWriter.On,
			LogPath:         conf.FileWriter.LogPath,
			RotateLogPath:   conf.FileWriter.RotateLogPath,
			WfLogPath:       conf.FileWriter.WfLogPath,
			RotateWfLogPath: conf.FileWriter.RotateWfLogPath,
		},
		CW: dlog.ConfConsoleWriter{
			On:    conf.ConsoleWriter.On,
			Color: conf.ConsoleWriter.Color,
		},
	}
	if err := dlog.SetupDefaultLogWithConf(logConf); err != nil {
		panic(err)
	}
	dlog.SetLayout("2006-01-02T15:04:05.000")
	return nil
}
