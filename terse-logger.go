package main

import (
	"fmt"

	"github.com/pion/logging"
)

// TerseLogger struct stasifying the Logger interface
type TerseLogger struct {
}

// Trace will log message at said level
func (c TerseLogger) Trace(msg string) { fmt.Printf("t %s\n", msg) }

// Tracef will log message at said level,printf style
func (c TerseLogger) Tracef(format string, args ...interface{}) {
	c.Trace(fmt.Sprintf(format, args...))
}

// Debug will log message at said level
func (c TerseLogger) Debug(msg string) { fmt.Printf("d %s\n", msg) }

// Debugf will log message at said level,printf style
func (c TerseLogger) Debugf(format string, args ...interface{}) {
	c.Debug(fmt.Sprintf(format, args...))
}

// Info will log message at said level
func (c TerseLogger) Info(msg string) { fmt.Printf("i %s\n", msg) }

// Infof will log message at said level,printf style
func (c TerseLogger) Infof(format string, args ...interface{}) {
	c.Info(fmt.Sprintf(format, args...))
}

// Warn will log message at said level
func (c TerseLogger) Warn(msg string) { fmt.Printf("w %s\n", msg) }

// Warnf will log message at said level,printf style
func (c TerseLogger) Warnf(format string, args ...interface{}) {
	c.Warn(fmt.Sprintf(format, args...))
}
// Error will log message at said level
func (c TerseLogger) Error(msg string) { fmt.Printf("e %s\n", msg) }

// Errorf will log message at said level,printf style
func (c TerseLogger) Errorf(format string, args ...interface{}) {
	c.Error(fmt.Sprintf(format, args...))
}

// TerseLoggerFactory satisfies the interface logging.LoggerFactory
// This allows us to create different loggers per subsystem. So we can
// add custom behavior
type TerseLoggerFactory struct {
}

// NewLogger creates LeveledLogger with given scope/subsystem name 
func (c TerseLoggerFactory) NewLogger(subsystem string) logging.LeveledLogger {
	tl := TerseLogger{}
	tl.Infof("Creating logger for %s", subsystem)
	return tl
}
