// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"context"
	"github.com/apache/thrift/lib/go/thrift"
	"net/http"
)

func processorHandler(w http.ResponseWriter, r *http.Request, processor thrift.TProcessor, _ttype tf_pco_type) {
	var protocolFactory thrift.TProtocolFactory
	switch _ttype {
	case JSON:
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case BINARY:
		protocolFactory = thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{})
	case COMPACT:
		protocolFactory = thrift.NewTCompactProtocolFactoryConf(&thrift.TConfiguration{})
	}
	transport := thrift.NewStreamTransport(r.Body, w)
	ioProtocol := protocolFactory.GetProtocol(transport)
	hc := newHttpContext(w, r)
	_, err := processor.Process(context.WithValue(context.Background(), "HttpContext", hc), ioProtocol, ioProtocol)
	if err != nil {
		logger.Error("processorHandler Error:", err)
	}
}
