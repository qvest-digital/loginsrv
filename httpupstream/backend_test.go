package httpupstream

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	. "github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/login"
)

func TestSetup(t *testing.T) {
	p, exist := login.GetProvider(ProviderName)
	True(t, exist)
	NotNil(t, p)

	backend, err := p(map[string]string{
		"upstream":   "https://google.com",
		"skipverify": "true",
		"timeout":    "20s",
	})

	NoError(t, err)
	Equal(t,
		"https://google.com",
		backend.(*Backend).auth.upstream.String())
	Equal(t,
		true,
		backend.(*Backend).auth.skipverify)
	Equal(t,
		time.Second*20,
		backend.(*Backend).auth.timeout)
}

func TestSetup_Default(t *testing.T) {
	p, exist := login.GetProvider(ProviderName)
	True(t, exist)
	NotNil(t, p)

	backend, err := p(map[string]string{
		"upstream": "https://google.com",
	})

	NoError(t, err)
	Equal(t,
		"https://google.com",
		backend.(*Backend).auth.upstream.String())
	Equal(t,
		false,
		backend.(*Backend).auth.skipverify)
	Equal(t,
		time.Second*60,
		backend.(*Backend).auth.timeout)
}

func TestSetup_Error(t *testing.T) {
	p, exist := login.GetProvider(ProviderName)
	True(t, exist)
	NotNil(t, p)

	_, err := p(map[string]string{})
	Error(t, err)
}

func TestSimpleBackend_Authenticate(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	u, _ := url.Parse(ts.URL)

	backend, err := NewBackend(u, time.Second, false)
	NoError(t, err)

	authenticated, userInfo, err := backend.Authenticate("bob-bcrypt", "secret")
	True(t, authenticated)
	Equal(t, "bob-bcrypt", userInfo.Sub)
	NoError(t, err)

	authenticated, userInfo, err = backend.Authenticate("bob-bcrypt", "fooo")
	False(t, authenticated)
	Equal(t, "", userInfo.Sub)
	NoError(t, err)

	authenticated, userInfo, err = backend.Authenticate("", "")
	False(t, authenticated)
	Equal(t, "", userInfo.Sub)
	NoError(t, err)
}

func newTestServer() *httptest.Server {
	passwordCheck := func(w http.ResponseWriter, r *http.Request) {
		u, p, k := r.BasicAuth()
		if !k {
			w.Header().Set("WWW-Authenticate", `Basic realm="test"`)
		}

		if !(u == "bob-bcrypt" && p == "secret") {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}

	return httptest.NewServer(http.HandlerFunc(passwordCheck))
}
