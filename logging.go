package main

import (
	"log"

	"github.com/gookit/color"
)

// oj/onlinejudge_command/logging.py
var prefix = map[string]string{
	"status":    "[" + color.Magenta.Sprint("x") + "] ",
	"success":   "[" + color.Green.Sprint("+") + "] ",
	"failure":   "[" + color.Red.Sprint("-") + "] ",
	"debug":     "[" + color.Red.Sprint("DEBUG") + "] ",
	"info":      "[" + color.LightBlue.Sprint("*") + "] ",
	"warning":   "[" + color.Yellow.Sprint("!") + "] ",
	"error":     "[" + color.Red.Sprint("ERROR") + "] ",
	"exception": "[" + color.Red.Sprint("EXCEPTION") + "] ",
	"critical":  "[" + color.Red.Sprint("CRITICAL") + "] ",
}

func LogEmit(f string, v ...interface{}) {
	log.Printf(f, v...)
}

func LogStatus(f string, v ...interface{}) {
	log.Printf(prefix["status"]+f, v...)
}

func LogSuccess(f string, v ...interface{}) {
	log.Printf(prefix["success"]+f, v...)
}

func LogFailure(f string, v ...interface{}) {
	log.Printf(prefix["failure"]+f, v...)
}

func LogDebug(f string, v ...interface{}) {
	log.Printf(prefix["debug"]+f, v...)
}

func LogInfo(f string, v ...interface{}) {
	log.Printf(prefix["info"]+f, v...)
}

func LogWarning(f string, v ...interface{}) {
	log.Printf(prefix["warning"]+f, v...)
}

func LogError(f string, v ...interface{}) {
	log.Printf(prefix["error"]+f, v...)
}

func LogException(f string, v ...interface{}) {
	log.Printf(prefix["exception"]+f, v...)
}

func LogCritical(f string, v ...interface{}) {
	log.Printf(prefix["critical"]+f, v...)
}
