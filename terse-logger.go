package main

import (
	"fmt"

	"github.com/pion/logging"
)

type terseLogger struct {
}

// Print all messages except trace
func (c terseLogger) Trace(msg string)                          {}
func (c terseLogger) Tracef(format string, args ...interface{}) {}

func (c terseLogger) Debug(msg string) { fmt.Printf("d %s\n", msg) }
func (c terseLogger) Debugf(format string, args ...interface{}) {
	c.Debug(fmt.Sprintf(format, args...))
}
func (c terseLogger) Info(msg string) { fmt.Printf("i %s\n", msg) }
func (c terseLogger) Infof(format string, args ...interface{}) {
	c.Info(fmt.Sprintf(format, args...))
}
func (c terseLogger) Warn(msg string) { fmt.Printf("w %s\n", msg) }
func (c terseLogger) Warnf(format string, args ...interface{}) {
	c.Warn(fmt.Sprintf(format, args...))
}
func (c terseLogger) Error(msg string) { fmt.Printf("e %s\n", msg) }
func (c terseLogger) Errorf(format string, args ...interface{}) {
	c.Error(fmt.Sprintf(format, args...))
}

// terseLoggerFactory satisfies the interface logging.LoggerFactory
// This allows us to create different loggers per subsystem. So we can
// add custom behavior
type terseLoggerFactory struct {
}

func (c terseLoggerFactory) NewLogger(subsystem string) logging.LeveledLogger {
	tl := terseLogger{}
	tl.Infof("Creating logger for %s", subsystem)
	return tl
}
