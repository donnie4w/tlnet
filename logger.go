// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	logging "log"
)

var logger = newLog()

type log struct {
	IsVaild bool
	logger  *logging.Logger
}

func newLog() *log {
	return &log{logger: logging.New(logging.Writer(), "[tlnet]", logging.Ldate|logging.Ltime)}
}

func (l *log) SetLogger(on bool) {
	l.IsVaild = on
}

func (l *log) Debug(v ...interface{}) {
	if l.IsVaild {
		l.logger.Println(v...)
	}
}

func (l *log) Error(v ...interface{}) {
	if l.IsVaild {
		l.logger.Fatal(v...)
	}
}
