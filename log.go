// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"log"
)

func SetLogger(on bool) {
	logger.SetLogger(on)
}

var logger = newLog()

type logu struct {
	IsVaild bool
	logger  *log.Logger
}

func newLog() *logu {
	return &logu{logger: log.New(log.Writer(), "[tlnet]", log.Ldate|log.Ltime)}
}

func (l *logu) SetLogger(on bool) {
	l.IsVaild = on
}

func (l *logu) Debug(v ...interface{}) {
	if l.IsVaild {
		l.logger.Println(v...)
	}
}

func (l *logu) Error(v ...interface{}) {
	if l.IsVaild {
		l.logger.Println(v...)
	}
}
