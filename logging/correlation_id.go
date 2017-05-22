package logging

import (
	"github.com/tarent/lib-compose/util"
	"math/rand"
	"net/http"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

var CorrelationIdHeader = "X-Correlation-Id"
var UserCorrelationCookie = ""

// EnsureCorrelationId returns the correlation from of the request.
// If the request does not have a correlation id, one will be generated and set to the request.
func EnsureCorrelationId(r *http.Request) string {
	id := r.Header.Get(CorrelationIdHeader)
	if id == "" {
		id = randStringBytes(10)
		r.Header.Set(CorrelationIdHeader, id)
	}
	return id
}

// GetCorrelationId returns the correlation from of the request.
func GetCorrelationId(h http.Header) string {
	return h.Get(CorrelationIdHeader)
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// GetCorrelationId returns the users correlation id of the headers.
func GetUserCorrelationId(h http.Header) string {
	if UserCorrelationCookie != "" {
		if value, found := util.ReadCookieValue(h, UserCorrelationCookie); found {
			return value
		}
	}
	return ""
}
