package logger

import (
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/k0kubun/pp"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

func isDebugMode() bool {
	env := strings.ToLower(os.Getenv("APP_DEBUG"))
	if env != "prod" {
		return true
	}
	return false
}

func init() {
	pp.ColoringEnabled = false
}

var (
	isDebug = isDebugMode()
	debug   = log.New(os.Stderr, "[DEBUG] ", log.LstdFlags)
	notice  = log.New(os.Stderr, "[NOTICE] ", log.LstdFlags)
	warn    = log.New(os.Stderr, "[WARNING] ", log.LstdFlags)
	err     = log.New(os.Stderr, "[ERROR] ", log.LstdFlags)
	crit    = log.New(os.Stderr, "[CRITICAL] ", log.LstdFlags)
)

// SetLogFile ...
func SetLogFile(path string) {
	f, _ := os.OpenFile(path+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	wrt := io.MultiWriter(os.Stdout, f)

	notice.SetOutput(wrt)
	warn.SetOutput(wrt)
	err.SetOutput(wrt)
	crit.SetOutput(wrt)
}

// SetDailyLogFile ...
func SetDailyLogFile(path string) {
	timestamp := time.Now().Format("2006-01-02")
	f, _ := os.OpenFile(path+"-"+timestamp+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	wrt := io.MultiWriter(os.Stdout, f)

	notice.SetOutput(wrt)
	warn.SetOutput(wrt)
	err.SetOutput(wrt)
	crit.SetOutput(wrt)
}

// SetRotatingLogFile ...
func SetRotatingLogFile(
	pattern string,
	options ...rotatelogs.Option,
) {
	rl, _ := rotatelogs.New(
		pattern,
		options...,
	)
	notice.SetOutput(rl)
	warn.SetOutput(rl)
	err.SetOutput(rl)
	crit.SetOutput(rl)
}

// LogDebugf ...
func LogDebugf(fmt string, vs ...interface{}) {
	if !isDebug {
		return
	}
	ppStr := pp.Sprintf(fmt, vs...)
	debug.Println(ppStr)
}

// LogDebug ...
func LogDebug(msg ...interface{}) {
	if !isDebug {
		return
	}
	ppStr := pp.Sprint(msg...)
	debug.Println(ppStr)
}

// LogNoticef ...
func LogNoticef(fmt string, vs ...interface{}) {
	notice.Printf(fmt, vs...)
}

// LogNotice ...
func LogNotice(msg interface{}) {
	notice.Println(msg)
}

// LogWarnf ...
func LogWarnf(fmt string, vs ...interface{}) {
	warn.Printf(fmt, vs...)
}

// LogWarn ...
func LogWarn(msg interface{}) {
	warn.Println(msg)
}

// LogErrorf ...
func LogErrorf(fmt string, vs ...interface{}) {
	err.Printf(fmt, vs...)
}

// LogError ...
func LogError(msg interface{}) {
	err.Println(msg)
}

// LogCritf ...
func LogCritf(fmt string, vs ...interface{}) {
	crit.Fatalf(fmt, vs...)
}

// LogCrit ...
func LogCrit(msg interface{}) {
	crit.Fatalln(msg)
}

// LogSetOutput ...
func LogSetOutput(w io.Writer) {
	log.SetOutput(w)
	debug.SetOutput(w)
	notice.SetOutput(w)
	warn.SetOutput(w)
	err.SetOutput(w)
	crit.SetOutput(w)
}
