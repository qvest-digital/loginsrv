package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/tarent/logrus"
	"net/http"
	"os"
	"testing"
	"time"
)

type logReccord struct {
	Type              string            `json:"type"`
	Timestamp         string            `json:"@timestamp"`
	CorrelationId     string            `json:"correlation_id"`
	RemoteIp          string            `json:"remote_ip"`
	Host              string            `json:"host"`
	URL               string            `json:"url"`
	Method            string            `json:"method"`
	Proto             string            `json:"proto"`
	Duration          int               `json:"duration"`
	ResponseStatus    int               `json:"response_status"`
	UserCorrelationId string            `json:"user_correlation_id"`
	Cookies           map[string]string `json:"cookies"`
	Error             string            `json:"error"`
	Message           string            `json:"message"`
	Level             string            `json:"level"`
	UserAgent         string            `json:"User_Agent"`
}

func Test_Logger_Set(t *testing.T) {
	a := assert.New(t)

	// given: an error logger in text format
	Set("error", true)
	defer Set("info", false)
	Logger.Formatter.(*logrus.TextFormatter).DisableColors = true
	b := bytes.NewBuffer(nil)
	Logger.Out = b

	// when: I log something
	Logger.Info("should be ignored ..")
	Logger.WithField("foo", "bar").Error("oops")

	// then: only the error text is contained
	// and it is text formated
	a.Regexp(`^time.* level\=error msg\=oops foo\=bar.*`, b.String())
}

func Test_Logger_Call(t *testing.T) {
	a := assert.New(t)

	// given a logger
	b := bytes.NewBuffer(nil)
	Logger.Out = b
	AccessLogCookiesBlacklist = []string{"ignore", "user_id"}
	UserCorrelationCookie = "user_id"

	// and a request
	r, _ := http.NewRequest("GET", "http://www.example.org/foo?q=bar", nil)
	r.Header = http.Header{
		CorrelationIdHeader: {"correlation-123"},
		"Cookie":            {"user_id=user-id-xyz; ignore=me; foo=bar;"},
	}

	resp := &http.Response{
		StatusCode: 404,
		Header:     http.Header{"Content-Type": {"text/html"}},
	}

	// when: We log a request with access
	start := time.Now().Add(-1 * time.Second)
	Call(r, resp, start, nil)

	// then: all fields match
	data := &logReccord{}
	err := json.Unmarshal(b.Bytes(), data)
	a.NoError(err)

	a.Equal("warning", data.Level)
	a.Equal("correlation-123", data.CorrelationId)
	a.Equal("user-id-xyz", data.UserCorrelationId)
	a.InDelta(1000, data.Duration, 0.5)
	a.Equal("", data.Error)
	a.Equal("www.example.org", data.Host)
	a.Equal("GET", data.Method)
	a.Equal("404 GET-> http://www.example.org/foo?q=bar", data.Message)
	a.Equal(404, data.ResponseStatus)
	a.Equal("call", data.Type)
	a.Equal("/foo?q=bar", data.URL)

	// when we call with an error
	b.Reset()
	start = time.Now().Add(-1 * time.Second)
	Call(r, nil, start, errors.New("oops"))

	// then: all fields match
	data = &logReccord{}
	err = json.Unmarshal(b.Bytes(), data)
	a.NoError(err)

	a.Equal("error", data.Level)
	a.Equal("oops", data.Error)
	a.Equal("oops", data.Message)
	a.Equal("correlation-123", data.CorrelationId)
	a.Equal("user-id-xyz", data.UserCorrelationId)
	a.InDelta(1000, data.Duration, 0.5)
	a.Equal("www.example.org", data.Host)
	a.Equal("GET", data.Method)
	a.Equal("call", data.Type)
	a.Equal("/foo?q=bar", data.URL)
}

func Test_Logger_Access(t *testing.T) {
	a := assert.New(t)

	// given a logger
	b := bytes.NewBuffer(nil)
	Logger.Out = b
	AccessLogCookiesBlacklist = []string{"ignore", "user_id"}
	UserCorrelationCookie = "user_id"

	// Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.84 Safari/537.36

	// and a request
	r, _ := http.NewRequest("GET", "http://www.example.org/foo?q=bar", nil)
	r.Header = http.Header{
		CorrelationIdHeader: {"correlation-123"},
		"Cookie":            {"user_id=user-id-xyz; ignore=me; foo=bar;"},
		"User-Agent":        {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.84 Safari/537.36"},
	}
	r.RemoteAddr = "127.0.0.1"

	// when: We log a request with access
	start := time.Now().Add(-1 * time.Second)
	Access(r, start, 201)

	// then: all fields match
	data := &logReccord{}
	err := json.Unmarshal(b.Bytes(), data)
	a.NoError(err)

	a.Equal("info", data.Level)
	a.Equal(map[string]string{"foo": "bar"}, data.Cookies)
	a.Equal("correlation-123", data.CorrelationId)
	a.InDelta(1000, data.Duration, 0.5)
	a.Equal("", data.Error)
	a.Equal("www.example.org", data.Host)
	a.Equal("GET", data.Method)
	a.Equal("HTTP/1.1", data.Proto)
	a.Equal("201 ->GET /foo?...", data.Message)
	a.Equal("127.0.0.1", data.RemoteIp)
	a.Equal(201, data.ResponseStatus)
	a.Equal("access", data.Type)
	a.Equal("/foo?q=bar", data.URL)
	a.Equal("user-id-xyz", data.UserCorrelationId)
	a.Equal("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.84 Safari/537.36", data.UserAgent)
}

func Test_Logger_Access_ErrorCases(t *testing.T) {
	a := assert.New(t)

	// given a logger
	b := bytes.NewBuffer(nil)
	Logger.Out = b

	// and a request
	r, _ := http.NewRequest("GET", "http://www.example.org/foo", nil)

	// when a status 404 is logged
	Access(r, time.Now(), 404)
	// then: all fields match
	data := logRecordFromBuffer(b)
	a.Equal("warning", data.Level)
	a.Equal("404 ->GET /foo", data.Message)

	// when a status 500 is logged
	b.Reset()
	Access(r, time.Now(), 500)
	// then: all fields match
	data = logRecordFromBuffer(b)
	a.Equal("error", data.Level)

	// when an error is logged
	b.Reset()
	AccessError(r, time.Now(), errors.New("oops"))
	// then: all fields match
	data = logRecordFromBuffer(b)
	a.Equal("error", data.Level)
	a.Equal("oops", data.Error)
	a.Equal("ERROR ->GET /foo", data.Message)
}

func Test_Logger_Application(t *testing.T) {
	a := assert.New(t)

	// given:
	UserCorrelationCookie = "user_id"
	header := http.Header{
		CorrelationIdHeader: {"correlation-123"},
		"Cookie":            {"user_id=user-id-xyz;"},
	}

	// when:
	entry := Application(header)

	// then:
	a.Equal("correlation-123", entry.Data["correlation_id"])
	a.Equal("user-id-xyz", entry.Data["user_correlation_id"])
}

func Test_Logger_LifecycleStart(t *testing.T) {
	a := assert.New(t)

	// given a logger
	b := bytes.NewBuffer(nil)
	Logger.Out = b

	// and
	someArguments := struct {
		Foo    string
		Number int
	}{
		Foo:    "bar",
		Number: 42,
	}

	// and an Environment Variable with the Build Number is set
	os.Setenv("BUILD_NUMBER", "b666")

	// when a LifecycleStart is logged
	LifecycleStart("my-app", someArguments)

	// then: it is logged
	data := mapFromBuffer(b)
	a.Equal("info", data["level"])
	a.Equal("lifecycle", data["type"])
	a.Equal("start", data["event"])
	a.Equal("bar", data["Foo"])
	a.Equal(42.0, data["Number"])
	a.Equal("b666", data["build_number"])
}

func Test_Logger_LifecycleStop_ByInterrupt(t *testing.T) {
	a := assert.New(t)

	// given a logger
	b := bytes.NewBuffer(nil)
	Logger.Out = b

	// and an Environment Variable with the Build Number is set
	os.Setenv("BUILD_NUMBER", "b666")

	// when a LifecycleStart is logged
	LifecycleStop("my-app", os.Interrupt, nil)

	// then: it is logged
	data := mapFromBuffer(b)
	a.Equal("info", data["level"])
	a.Equal("stopping application: my-app (interrupt)", data["message"])
	a.Equal("lifecycle", data["type"])
	a.Equal("stop", data["event"])
	a.Equal("interrupt", data["signal"])
	a.Equal("b666", data["build_number"])
}

func Test_Logger_LifecycleStop_ByError(t *testing.T) {
	a := assert.New(t)

	// given a logger
	b := bytes.NewBuffer(nil)
	Logger.Out = b

	// and an Environment Variable with the Build Number is set
	os.Setenv("BUILD_NUMBER", "b666")

	// when a LifecycleStart is logged
	LifecycleStop("my-app", nil, errors.New("error"))

	// then: it is logged
	data := mapFromBuffer(b)
	a.Equal("error", data["level"])
	a.Equal("stopping application: my-app (error)", data["message"])
	a.Equal("lifecycle", data["type"])
	a.Equal("stop", data["event"])
	a.Equal(nil, data["signal"])
	a.Equal("b666", data["build_number"])
}

func Test_Logger_ServerClosed(t *testing.T) {
	a := assert.New(t)

	// given a logger
	b := bytes.NewBuffer(nil)
	Logger.Out = b

	// and an Environment Variable with the Build Number is set
	os.Setenv("BUILD_NUMBER", "b666")

	// when a LifecycleStart is logged
	ServerClosed("my-app")

	// then: it is logged
	data := mapFromBuffer(b)
	a.Equal("info", data["level"])
	a.Equal("http server was closed: my-app", data["message"])
	a.Equal("application", data["type"])
	a.Equal("stop", data["event"])
	a.Equal("b666", data["build_number"])
}

func Test_Logger_Cacheinfo(t *testing.T) {
	a := assert.New(t)

	// given a logger
	Set("debug", false)
	defer Set("info", false)
	b := bytes.NewBuffer(nil)
	Logger.Out = b

	// when a positive cachinfo is logged
	Cacheinfo("/foo", true)

	// then: it is logged
	data := mapFromBuffer(b)
	a.Equal("/foo", data["url"])
	a.Equal("cacheinfo", data["type"])
	a.Equal(true, data["hit"])
	a.Equal("cache hit: /foo", data["message"])

	b.Reset()
	// logging a non hit
	Cacheinfo("/foo", false)
	data = mapFromBuffer(b)
	a.Equal(false, data["hit"])
	a.Equal("cache miss: /foo", data["message"])
}

func Test_Logger_GetRemoteIp1(t *testing.T) {
	a := assert.New(t)
	req, _ := http.NewRequest("GET", "test.com", nil)
	req.Header["X-Cluster-Client-Ip"] = []string{"1234"}
	ret := getRemoteIp(req)
	a.Equal("1234", ret)
}

func Test_Logger_GetRemoteIp2(t *testing.T) {
	a := assert.New(t)
	req, _ := http.NewRequest("GET", "test.com", nil)
	req.Header["X-Real-Ip"] = []string{"1234"}
	ret := getRemoteIp(req)
	a.Equal("1234", ret)
}

func Test_Logger_GetRemoteIp3(t *testing.T) {
	a := assert.New(t)
	req, _ := http.NewRequest("GET", "test.com", nil)
	req.RemoteAddr = "1234:80"
	ret := getRemoteIp(req)
	a.Equal("1234", ret)
}

func logRecordFromBuffer(b *bytes.Buffer) *logReccord {
	data := &logReccord{}
	err := json.Unmarshal(b.Bytes(), data)
	if err != nil {
		panic(err.Error() + " " + b.String())
	}
	return data
}

func mapFromBuffer(b *bytes.Buffer) map[string]interface{} {
	data := map[string]interface{}{}
	err := json.Unmarshal(b.Bytes(), &data)
	if err != nil {
		panic(err.Error() + " " + b.String())
	}
	return data
}
