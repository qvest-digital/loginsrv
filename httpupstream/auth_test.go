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
