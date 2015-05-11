// Package spoor provides logging modeled after the Python logging package.
package spoor

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	CRITICAL
	FATAL
)

// String returns the description of one of the predefined LogLevel constants.
func (loglevel LogLevel) String() string {
	switch loglevel {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	case FATAL:
		return "FATAL"
	}
	return "UNKNOWN"
}

type LogRecord struct {
	level LogLevel
	name  string
	msg   string
	args  []interface{}
}

func NewLogRecord(level LogLevel, name string, msg string, args ...interface{}) *LogRecord {
	logRecord := new(LogRecord)
	logRecord.level = level
	logRecord.name = name
	logRecord.msg = msg
	logRecord.args = args
	return logRecord
}

func (r *LogRecord) GetLevel() LogLevel {
	return r.level
}

type Formatter struct {
	fmt     string
	dateFmt string
}

func NewFormatter(params ...string) *Formatter {
	formatter := new(Formatter)
	if len(params) > 0 {
		formatter.fmt = params[0]
	}
	if len(params) > 1 {
		formatter.dateFmt = params[1]
	}
	return formatter
}

func (f *Formatter) Format(logRecord *LogRecord) string {
	now := time.Now()
	msg := ""
	if len(logRecord.args) > 0 {
		msg = fmt.Sprintf(logRecord.msg, logRecord.args...)
	} else {
		msg = logRecord.msg
	}
	format := strings.Replace(f.fmt, "{levelname}", logRecord.level.String(), 1)
	format = strings.Replace(format, "{message}", msg, 1)
	format = strings.Replace(format, "{asctime}", now.Format(f.dateFmt), 1)
	return format
}

type ILogHandler interface {
	GetLevel() LogLevel
	SetLevel(LogLevel)
	GetFormatter() *Formatter
	SetFormatter(*Formatter)
	Handle(*LogRecord)
	Emit(*LogRecord)
	Format(*LogRecord) string
}

type LogHandler struct {
	level     LogLevel
	formatter *Formatter
	logger    io.Writer
}

func NewLogHandler() *LogHandler {
	logHandler := new(LogHandler)
	logHandler.SetLevel(config.level)
	logHandler.SetFormatter(NewFormatter(config.format, config.datefmt))
	return logHandler
}

func (h *LogHandler) GetLevel() LogLevel {
	return h.level
}

func (h *LogHandler) SetLevel(level LogLevel) {
	h.level = level
}

func (h *LogHandler) GetFormatter() *Formatter {
	return h.formatter
}

func (h *LogHandler) SetFormatter(formatter *Formatter) {
	h.formatter = formatter
}

func (h *LogHandler) Handle(logRecord *LogRecord) {
	h.Emit(logRecord)
}

func (h *LogHandler) Emit(logRecord *LogRecord) {
	fmt.Fprintln(h.logger, h.Format(logRecord))
}

func (h *LogHandler) Format(logRecord *LogRecord) string {
	return h.formatter.Format(logRecord)
}

type StreamHandler struct {
	LogHandler
}

func NewStreamHandler(stream ...io.Writer) *StreamHandler {
	var handlerStream io.Writer = os.Stderr

	if len(stream) > 0 {
		handlerStream = stream[0]
	}
	streamHandler := new(StreamHandler)
	streamHandler.LogHandler = *NewLogHandler()
	streamHandler.logger = handlerStream
	return streamHandler
}

type Logger struct {
	level    LogLevel
	name     string
	handlers []ILogHandler
}

func NewLogger(name string) *Logger {
	logger := new(Logger)
	logger.name = name
	logger.handlers = make([]ILogHandler, 0, 2)
	return logger
}

func (l *Logger) GetName() string {
	return l.name
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) Log(level LogLevel, msg string, args ...interface{}) {
	logRecord := NewLogRecord(level, l.name, msg, args...)
	for _, handler := range l.handlers {
		if level >= handler.GetLevel() {
			handler.Handle(logRecord)
		}
	}
	if level == FATAL {
		os.Exit(1)
	}
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	l.Log(DEBUG, msg, args...)
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.Log(INFO, msg, args...)
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.Log(WARNING, msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.Log(ERROR, msg, args...)
}

func (l *Logger) Critical(msg string, args ...interface{}) {
	l.Log(CRITICAL, msg, args...)
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Log(FATAL, msg, args...)
}

func (l *Logger) AddHandler(handler ILogHandler) {
	l.handlers = append(l.handlers, handler)
}

var loggers = struct {
	sync.RWMutex
	items map[string]*Logger
}{items: make(map[string]*Logger)}

var config = struct {
	filename string // Specifies that a FileHandler be created, using the specified filename, rather than a StreamHandler.
	filemode string // Specifies the mode to open the file, if filename is specified
	// (if filemode is unspecified, it defaults to ‘a’).
	format  string    // Use the specified format string for the handler.
	datefmt string    // Use the specified date/time format.
	level   LogLevel  // Set the root logger level to the specified level.
	stream  io.Writer // Use the specified stream to initialize the StreamHandler.
	// Note that this argument is incompatible with ‘filename’ - if both are present, ‘stream’ is ignored.
}{level: INFO, format: "{levelname}: {asctime} - {message}", datefmt: "2006-01-02 15:04:05"}

func BasicConfig(conf map[string]interface{}) {
	for k, value := range conf {
		key := strings.ToLower(k)
		switch key {
		case "filename":
			config.filename = value.(string)
		case "filemode":
			config.filemode = value.(string)
		case "format":
			config.format = value.(string)
		case "datefmt":
			config.datefmt = value.(string)
		case "stream":
			config.stream = value.(io.Writer)
		}
	}
}

func GetLogger(loggername ...string) *Logger {
	loggers.Lock()
	defer loggers.Unlock()
	name := "root"
	if len(loggername) == 1 {
		name = loggername[0]
	}
	if logger, ok := loggers.items[name]; ok {
		return logger
	}
	logger := NewLogger(name)
	loggers.items[name] = logger
	return logger
}
