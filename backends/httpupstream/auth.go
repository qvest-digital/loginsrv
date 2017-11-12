package httpupstream

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

// Auth is the httpupstream authenticater
type Auth struct {
	upstream   *url.URL
	skipverify bool
	timeout    time.Duration
}

// NewAuth creates an httpupstream authenticater
func NewAuth(upstream *url.URL, timeout time.Duration, skipverify bool) (*Auth, error) {
	a := &Auth{
		upstream:   upstream,
		skipverify: skipverify,
		timeout:    timeout,
	}

	return a, nil
}

// Authenticate the user
func (a *Auth) Authenticate(username, password string) (bool, error) {
	c := &http.Client{
		Timeout: a.timeout,
	}

	if a.upstream.Scheme == "https" && a.skipverify {
		c.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	req, err := http.NewRequest("GET", a.upstream.String(), nil)
	if err != nil {
		return false, err
	}

	req.SetBasicAuth(username, password)

	resp, err := c.Do(req)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != 200 {
		return false, nil
	}

	return true, nil
}
