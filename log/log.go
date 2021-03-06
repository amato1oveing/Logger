package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

type LEVEL int //日志等级

const (
	DEBUG LEVEL = iota
	INFO
	WARN
	ERROR
)

const (
	COLOR_GREEN  = "\033[0;32m" // 绿 DEBUG
	COLOR_WHITE  = "\033[0;37m" // 白 INFO
	COLOR_YELLOW = "\033[0;33m" // 黄 WARN
	COLOR_RED    = "\033[0;31m" // 红 ERROR
	COLOR_END    = "\033[0m"
)

var logFile *os.File

func InitFile() {
	if logFile != nil {
		return
	}
	var err error
	logFile, err = os.OpenFile("./zc.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return
	}
}

type Logger struct {
	LogLever LEVEL
	MsgCh    chan MsgWithColor
}

type MsgWithColor struct {
	msg   string
	color string
}

func NewLogger(level string) *Logger {
	log := &Logger{
		LogLever: parse(level),
		MsgCh:    make(chan MsgWithColor, 10000),
	}
	go log.run()
	return log
}

func (this Logger) run() {
	InitFile()
	for i := 0; i < 10; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("log panic:", r)
				}
			}()
			for MsgWithColor := range this.MsgCh {
				fmt.Fprintln(logFile, MsgWithColor.msg)
				fmt.Println(fmt.Sprintf("%s%s%s", MsgWithColor.color, MsgWithColor.msg, COLOR_END))
			}
		}()
	}
}

func (this Logger) Close() {
	t := time.Tick(10 * time.Millisecond)
	<-t
	close(this.MsgCh)
	logFile.Close()
}

//解析日志等级
func parse(str string) LEVEL {
	switch strings.ToLower(str) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return DEBUG //默认日志等级为DEBUG
	}
}

//判断日志等级
func (this Logger) checkLevel(level LEVEL) bool {
	if this.LogLever <= level {
		return false
	}
	return true
}

func (this Logger) writeMsg(flag string, color string, msg string, args ...interface{}) {
	if this.checkLevel(parse(flag)) {
		return
	}
	msg = fmt.Sprintf(msg, args...)

	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	msg = fmt.Sprintf("[%s] <%s> %s:%d:%s %s", flag, time.Now().Format("2006-01-02 15:04:05"), file, line, funcName, msg)

	this.MsgCh <- MsgWithColor{
		msg:   msg,
		color: color,
	}
}

func (this Logger) DEBUG(msg string, args ...interface{}) {
	this.writeMsg("DEBUG", COLOR_GREEN, msg, args...)
}

func (this Logger) INFO(msg string, args ...interface{}) {
	this.writeMsg("INFO", COLOR_WHITE, msg, args...)
}

func (this Logger) WARN(msg string, args ...interface{}) {
	this.writeMsg("WARN", COLOR_YELLOW, msg, args...)
}

func (this Logger) ERROR(msg string, args ...interface{}) {
	this.writeMsg("ERROR", COLOR_RED, msg, args...)
}
