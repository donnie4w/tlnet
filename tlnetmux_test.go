// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.como/donnie4w/tlnet

package tlnet

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock handler function to be used in tests
func mockHandler(response string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(response))
	}
}

func TestTrieMux(t *testing.T) {
	// Create a new TlnetMux instance
	trieMux := NewTlnetMux()

	// add routes to the trie
	trieMux.Handle("/path", mockHandler("static path"))
	trieMux.Handle("/path/*", mockHandler("wildcard path"))
	trieMux.Handle("/user/:id", mockHandler("user with id"))
	trieMux.Handle("/static/*", mockHandler("static files"))
	trieMux.Handle("/foo/:bar/baz", mockHandler("complex route"))

	// Test case for static path
	req := httptest.NewRequest("GET", "/path", nil)
	w := httptest.NewRecorder()
	trieMux.ServeHTTP(w, req)
	if w.Body.String() != "static path" {
		t.Errorf("Expected 'static path', got '%s'", w.Body.String())
	}

	// Test case for wildcard path
	req = httptest.NewRequest("GET", "/path/anything", nil)
	w = httptest.NewRecorder()
	trieMux.ServeHTTP(w, req)
	if w.Body.String() != "wildcard path" {
		t.Errorf("Expected 'wildcard path', got '%s'", w.Body.String())
	}

	// Test case for path with parameter
	req = httptest.NewRequest("GET", "/user/123", nil)
	w = httptest.NewRecorder()
	trieMux.ServeHTTP(w, req)
	if w.Body.String() != "user with id" {
		t.Errorf("Expected 'user with id', got '%s'", w.Body.String())
	}

	// Test case for static files wildcard
	req = httptest.NewRequest("GET", "/static/file.css", nil)
	w = httptest.NewRecorder()
	trieMux.ServeHTTP(w, req)
	if w.Body.String() != "static files" {
		t.Errorf("Expected 'static files', got '%s'", w.Body.String())
	}

	// Test case for complex path
	req = httptest.NewRequest("GET", "/foo/bar/baz", nil)
	w = httptest.NewRecorder()
	trieMux.ServeHTTP(w, req)
	if w.Body.String() != "complex route" {
		t.Errorf("Expected 'complex route', got '%s'", w.Body.String())
	}

	// Test case for not found
	req = httptest.NewRequest("GET", "/nonexistent", nil)
	w = httptest.NewRecorder()
	trieMux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got '%d'", w.Code)
	}
}

// Mock handler function to print the matched path
func mockHandler2(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Handler matched: %s\n", name)
		fmt.Printf("Matched handler: %s for path: %s\n", name, r.URL.Path)
	}
}

func Test_trie(t *testing.T) {
	trieMux := NewTlnetMux()
	// Register routes with handlers
	trieMux.Handle("/path", mockHandler2("static /path"))
	trieMux.Handle("/path/*", mockHandler2("wildcard /path/*"))
	trieMux.Handle("/user/:id", mockHandler2("/user/:id"))
	trieMux.Handle("/static/*/aaa", mockHandler2("static /static/*/aaa"))
	trieMux.Handle("/foo/:bar/baz", mockHandler2("/foo/:bar/baz"))

	// Start HTTP server
	addr := ":8080"
	t.Log("Starting server at ", addr)
	http.ListenAndServe(addr, trieMux)
}

func Test_trietlnet(t *testing.T) {
	tlnet := NewTlnet()
	SetLogger(true)
	tlnet.Handle("/aaa/*", func(hc *HttpContext) { hc.ResponseString("ok-a") })
	tlnet.Handle("/bbb/:id/:name/order", func(hc *HttpContext) { hc.ResponseString(hc.GetParam("id") + ":" + hc.GetParam("name")) })
	tlnet.POST("/ccc/*", func(hc *HttpContext) { hc.ResponseString("ccc") })
	tlnet.HandleStatic("/*", "./test", nil)
	tlnet.HttpStart(":8081")
}

func tlnetTextResponseHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, tlnet!"))
}

func tlnetParamResponseHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, tlnet!"))
}

func setupTlnetHttpRouter() TlnetMux {
	mux := NewTlnetMux()
	mux.HandleFunc("/api/v1/status", tlnetTextResponseHandler)
	mux.HandleFunc("/api/v1/users/", tlnetTextResponseHandler) // 模拟路径参数处理
	mux.HandleFunc("/api/v1/users", tlnetTextResponseHandler)
	mux.HandleFunc("/api/v1/products/:id", tlnetParamResponseHandler)
	return mux
}

func BenchmarkTlnet(b *testing.B) {
	router := setupTlnetHttpRouter()
	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

func TestTlnet(t *testing.T) {
	router := setupTlnetHttpRouter()
	req := httptest.NewRequest("GET", "/api/v1/products/123", nil)
	w := httptest.NewRecorder()
	for i := 0; i < 2; i++ {
		router.ServeHTTP(w, req)
		t.Logf("Response %d: %s", i+1, w.Body.String())
	}
}

func TestPath(t *testing.T) {
	path := "/aaa/bbb/ccc"
	n := len(path)
	var get bool
	partStart := 1
	for i := 1; i < n; i++ {
		get = false
		if path[i] == '/' {
			get = true
		} else if i == n-1 {
			i = n
		}
		if get || i == n {
			part := path[partStart:i]
			partStart = i + 1
			fmt.Println(part)
		}
	}
}
