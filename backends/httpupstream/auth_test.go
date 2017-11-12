package httpupstream

import (
	"net/url"
	"testing"
	"time"

	. "github.com/stretchr/testify/assert"
)

func TestAuth_UnknownUser(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	u, _ := url.Parse(ts.URL)

	auth, err := NewAuth(u, time.Second, false)
	NoError(t, err)

	authenticated, err := auth.Authenticate("unknown", "secret")
	NoError(t, err)
	False(t, authenticated)
}

func TestAuth_KnownUser(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	u, _ := url.Parse(ts.URL)

	auth, err := NewAuth(u, time.Second, false)
	NoError(t, err)

	authenticated, err := auth.Authenticate("bob-bcrypt", "s3krud")
	NoError(t, err)
	False(t, authenticated)
}

func TestAuth_ValidCredentials(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	u, _ := url.Parse(ts.URL)

	auth, err := NewAuth(u, time.Second, false)
	NoError(t, err)

	authenticated, err := auth.Authenticate("bob-bcrypt", "secret")
	NoError(t, err)
	True(t, authenticated)
}

func TestAuth_InvalidUrl(t *testing.T) {
	invalidUrl := &url.URL{Scheme: "\\\\"}

	auth, err := NewAuth(invalidUrl, time.Second, false)
	NoError(t, err)

	_, err = auth.Authenticate("foo", "bar")
	Error(t, err)
}

func TestAuth_InvalidHost(t *testing.T) {
	invalidServer, _ := url.Parse("http://0.0.0.0.0")

	auth, err := NewAuth(invalidServer, time.Second, false)
	NoError(t, err)

	_, err = auth.Authenticate("foo", "bar")
	Error(t, err)
}
