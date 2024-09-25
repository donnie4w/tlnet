// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"fmt"
	"github.com/donnie4w/simplelog/logging"
)

var logger = newLog()

type level uint8

const (
	_ level = iota
	DEBUG
	INFO
	WARN
	ERROR
)

type log struct {
	IsVaild bool
	logger  *logging.Logging
}

func newLog() *log {
	return &log{logger: logging.NewLogger().SetFormat(logging.FORMAT_DATE | logging.FORMAT_TIME | logging.FORMAT_MICROSECONDS)}
}

func (l *log) SetLogger(on bool) {
	l.IsVaild = on
}

func (l *log) SetLoggerLevel(level level) {
	switch level {
	case DEBUG:
		l.logger.SetLevel(logging.LEVEL_DEBUG)
	case INFO:
		l.logger.SetLevel(logging.LEVEL_INFO)
	case WARN:
		l.logger.SetLevel(logging.LEVEL_WARN)
	case ERROR:
		l.logger.SetLevel(logging.LEVEL_ERROR)
	}
}

func (l *log) Debug(v ...interface{}) {
	if l.IsVaild {
		l.logger.Debug(v...)
	}
}

func (l *log) Debugf(format string, v ...interface{}) {
	if l.IsVaild {
		l.logger.Debug(fmt.Sprintf(format, v...))
	}
}

func (l *log) Info(v ...interface{}) {
	if l.IsVaild {
		l.logger.Info(v...)
	}
}

func (l *log) Infof(format string, v ...interface{}) {
	if l.IsVaild {
		l.logger.Info(fmt.Sprintf(format, v...))
	}
}

func (l *log) Warn(v ...interface{}) {
	if l.IsVaild {
		l.logger.Warn(v...)
	}
}

func (l *log) Warnf(format string, v ...interface{}) {
	if l.IsVaild {
		l.logger.Warn(fmt.Sprintf(format, v...))
	}
}

func (l *log) Error(v ...interface{}) {
	if l.IsVaild {
		l.logger.Error(v...)
	}
}

func (l *log) Errorf(format string, v ...interface{}) {
	if l.IsVaild {
		l.logger.Error(fmt.Sprintf(format, v...))
	}
}
