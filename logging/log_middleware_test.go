package logging

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_LogMiddleware_Panic(t *testing.T) {
	a := assert.New(t)

	// given: a logger
	b := bytes.NewBuffer(nil)
	Logger.Out = b

	// and a handler which raises a panic
	lm := NewLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := []int{}
		i[100]++
	}))

	r, _ := http.NewRequest("GET", "http://www.example.org/foo", nil)

	lm.ServeHTTP(httptest.NewRecorder(), r)

	data := logRecordFromBuffer(b)
	a.Contains(data.Error, "logging.Test_LogMiddleware_Panic.func1")
	a.Contains(data.Error, "runtime error: index out of range")
	a.Contains(data.Message, "ERROR ->GET /foo")
	a.Equal(data.Level, "error")
}

func Test_LogMiddleware_Log_implicit200(t *testing.T) {
	a := assert.New(t)

	// given: a logger
	b := bytes.NewBuffer(nil)
	Logger.Out = b

	// and a handler which gets an 200er code implicitly
	lm := NewLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))

	r, _ := http.NewRequest("GET", "http://www.example.org/foo", nil)

	lm.ServeHTTP(httptest.NewRecorder(), r)

	data := logRecordFromBuffer(b)
	a.Equal("", data.Error)
	a.Equal("200 ->GET /foo", data.Message)
	a.Equal(200, data.ResponseStatus)
	a.Equal("info", data.Level)
}

func Test_LogMiddleware_Log_404(t *testing.T) {
	a := assert.New(t)

	// given: a logger
	b := bytes.NewBuffer(nil)
	Logger.Out = b

	// and a handler which gets an 404er code explicitly
	lm := NewLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))

	r, _ := http.NewRequest("GET", "http://www.example.org/foo", nil)

	lm.ServeHTTP(httptest.NewRecorder(), r)

	data := logRecordFromBuffer(b)
	a.Equal("", data.Error)
	a.Equal("404 ->GET /foo", data.Message)
	a.Equal(404, data.ResponseStatus)
	a.Equal("warning", data.Level)
}
