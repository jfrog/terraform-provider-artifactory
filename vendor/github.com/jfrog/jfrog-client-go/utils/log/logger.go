package log

import (
	"io"
	"log"
	"os"
)

var Logger Log

type LevelType int

const (
	ERROR LevelType = iota
	WARN
	INFO
	DEBUG
)

// Creates a new logger with a given LogLevel.
// All logs are written to Stderr by default (output to Stdout).
// If logToWriter != nil, logging is done to the provided writer instead.
func NewLogger(logLevel LevelType, logToWriter io.Writer) Log {
	logger := new(jfrogLogger)
	logger.SetLogLevel(logLevel)
	logger.SetOutputWriter(os.Stdout)
	logger.SetLogsWriter(logToWriter)
	return logger
}

type jfrogLogger struct {
	LogLevel  LevelType
	OutputLog *log.Logger
	DebugLog  *log.Logger
	InfoLog   *log.Logger
	WarnLog   *log.Logger
	ErrorLog  *log.Logger
}

func SetLogger(newLogger Log) {
	Logger = newLogger
}

func (logger *jfrogLogger) SetLogLevel(LevelEnum LevelType) {
	logger.LogLevel = LevelEnum
}

func (logger *jfrogLogger) SetOutputWriter(writer io.Writer) {
	logger.OutputLog = log.New(writer, "", 0)
}

// Set the logs writer to Stderr unless an alternative one is provided.
func (logger *jfrogLogger) SetLogsWriter(writer io.Writer) {
	if writer == nil {
		writer = os.Stderr
	}
	logger.DebugLog = log.New(writer, "[Debug] ", 0)
	logger.InfoLog = log.New(writer, "[Info] ", 0)
	logger.WarnLog = log.New(writer, "[Warn] ", 0)
	logger.ErrorLog = log.New(writer, "[Error] ", 0)
}

func GetLogLevel() LevelType {
	return Logger.GetLogLevel()
}

func validateLogInit() {
	if Logger == nil {
		panic("Logger not initialized. See API documentation.")
	}
}

func Debug(a ...interface{}) {
	validateLogInit()
	Logger.Debug(a...)
}

func Info(a ...interface{}) {
	validateLogInit()
	Logger.Info(a...)
}

func Warn(a ...interface{}) {
	validateLogInit()
	Logger.Warn(a...)
}

func Error(a ...interface{}) {
	validateLogInit()
	Logger.Error(a...)
}

func Output(a ...interface{}) {
	validateLogInit()
	Logger.Output(a...)
}

func (logger jfrogLogger) GetLogLevel() LevelType {
	return logger.LogLevel
}

func (logger jfrogLogger) Debug(a ...interface{}) {
	if logger.GetLogLevel() >= DEBUG {
		logger.DebugLog.Println(a...)
	}
}

func (logger jfrogLogger) Info(a ...interface{}) {
	if logger.GetLogLevel() >= INFO {
		logger.InfoLog.Println(a...)
	}
}

func (logger jfrogLogger) Warn(a ...interface{}) {
	if logger.GetLogLevel() >= WARN {
		logger.WarnLog.Println(a...)
	}
}

func (logger jfrogLogger) Error(a ...interface{}) {
	if logger.GetLogLevel() >= ERROR {
		logger.ErrorLog.Println(a...)
	}
}

func (logger jfrogLogger) Output(a ...interface{}) {
	logger.OutputLog.Println(a...)
}

type Log interface {
	GetLogLevel() LevelType
	SetLogLevel(LevelType)
	SetOutputWriter(writer io.Writer)
	SetLogsWriter(writer io.Writer)

	Debug(a ...interface{})
	Info(a ...interface{})
	Warn(a ...interface{})
	Error(a ...interface{})
	Output(a ...interface{})
}
