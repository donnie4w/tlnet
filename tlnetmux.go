// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"github.com/donnie4w/gofer/hashmap"
	"github.com/donnie4w/gofer/util"
	"net/http"
	"strings"
)

// serveMux implements http.Handler interface
type serveMux struct {
	trie *trie
}

func NewTlnetMux() TlnetMux {
	return &serveMux{trie: newTrie()}
}

func NewTlnetMuxWithLimit(limit int) TlnetMux {
	return &serveMux{trie: newTrieWithLimit(limit)}
}

func (r *serveMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler, ok, path := r.findHandler(req)
	if ok {
		handler.ServeHTTP(w, req)
	} else if path != "" && !strings.HasSuffix(path, "/") {
		http.Redirect(w, req, path+"/", http.StatusMovedPermanently)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (t *serveMux) Handle(path string, handler http.Handler) {
	t.trie.handle(path, handler)
}

func (t *serveMux) HandleFunc(path string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	t.trie.handle(path, http.HandlerFunc(handlerFunc))
}

func (t *serveMux) findHandler(req *http.Request) (http.Handler, bool, string) {
	if logger.IsVaild {
		logger.Debug("[url.path]", req.URL.Path)
	}
	return t.trie.findHandler(req)
}

type trieNode struct {
	children map[string]*trieNode
	handler  http.Handler
	wildcard *trieNode
	param    *trieNode
	path     string
}

type trie struct {
	root *trieNode
	m    *hashmap.LimitHashMap[uint64, http.Handler]
}

func newTrie() *trie {
	return &trie{
		root: &trieNode{children: make(map[string]*trieNode)},
		m:    hashmap.NewLimitHashMap[uint64, http.Handler](_SS),
	}
}

func newTrieWithLimit(limit int) *trie {
	return &trie{
		root: &trieNode{children: make(map[string]*trieNode)},
		m:    hashmap.NewLimitHashMap[uint64, http.Handler](limit),
	}
}

func (t *trie) handle(path string, handler http.Handler) {
	path = strings.TrimSpace(path)
	parts := strings.Split(path, "/")[1:]
	node := t.root
	for _, part := range parts {
		if part == "*" || part == "" {
			if node.wildcard == nil {
				node.wildcard = &trieNode{children: make(map[string]*trieNode), path: part}
			}
			node = node.wildcard
		} else if strings.HasPrefix(part, ":") {
			if node.param == nil {
				node.param = &trieNode{children: make(map[string]*trieNode), path: part[1:]}
			}
			node = node.param
		} else {
			if node.children[part] == nil {
				node.children[part] = &trieNode{children: make(map[string]*trieNode)}
			}
			node = node.children[part]
		}
	}
	node.handler = handler
}

func (t *trie) findHandler(req *http.Request) (http.Handler, bool, string) {
	path := req.URL.Path
	ph := util.Hash64([]byte(path))

	if h, b := t.m.Get(ph); b {
		return h, b, ""
	}

	parts := strings.Split(path[1:], "/")
	node := t.root
	var kvmap map[string]string
	var nodewildcard *trieNode
	for i, part := range parts {
		if node.children[part] != nil {
			if node.wildcard != nil {
				nodewildcard = node.wildcard
			}
			node = node.children[part]
			if i == len(parts)-1 && node.handler == nil && node.wildcard != nil {
				return nil, false, path
			}
		} else if node.param != nil {
			if kvmap == nil {
				kvmap = map[string]string{}
			}
			kvmap[node.param.path] = part
			if node.wildcard != nil {
				nodewildcard = node.wildcard
			}
			node = node.param
		} else if node.wildcard != nil {
			node = node.wildcard
			nodewildcard = node
		} else {
			if nodewildcard != nil {
				node = nodewildcard
				break
			} else {
				return nil, false, ""
			}
		}
	}

	if r := node.handler; r != nil {
		if len(kvmap) > 0 {
			queryParams := req.URL.Query()
			for k, v := range kvmap {
				queryParams.Add(k, v)
			}
			req.URL.RawQuery = queryParams.Encode()
		} else if len(path) < 256 {
			t.m.Put(ph, r)
		}
		//t.m.Put(ph, newmkv(r, kvmap).RawQueryEncode(req))
		return r, true, ""
	}
	return nil, false, ""
}

type mkv struct {
	handle http.Handler
	kv     map[string]string
}

func newmkv(handle http.Handler, kv map[string]string) *mkv {
	return &mkv{handle, kv}
}

func (m *mkv) RawQueryEncode(req *http.Request) *mkv {
	if len(m.kv) > 0 {
		queryParams := req.URL.Query()
		for k, v := range m.kv {
			queryParams.Add(k, v)
		}
		req.URL.RawQuery = queryParams.Encode()
	}
	return m
}
