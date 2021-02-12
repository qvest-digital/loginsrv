package oauth2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/stretchr/testify/assert"
)

var (
	googleTestUserResponse = `{
	"sub": "10467329456789",
	"name": "Testy Test",
	"given_name": "Testy",
	"family_name": "Test",
	"profile": "https://plus.google.com/10467329456789",
	"picture": "https://lh6.googleusercontent.com/-alknmlknzT_YQ/AAAAAAAAAAI/AAAAAAAAABU/4gNvDUeED14/photo.jpg",
	"email": "test@example.com",
	"email_verified": true,
	"gender": "male",
	"locale": "de",
	"hd": "example.com"
	}`
	googleTestUserResponseNoEmail = `{
		"sub": "10467329456789",
		"name": "Testy Test",
		"given_name": "Testy",
		"family_name": "Test",
		"profile": "https://plus.google.com/10467329456789",
		"picture": "https://lh6.googleusercontent.com/-alknmlknzT_YQ/AAAAAAAAAAI/AAAAAAAAABU/4gNvDUeED14/photo.jpg",
		"email_verified": false,
		"gender": "male",
		"locale": "de",
		"hd": "example.com"
	  }`
	googleTestUserResponseUnverifiedEmail = `{
		"sub": "10467329456789",
		"name": "Testy Test",
		"given_name": "Testy",
		"family_name": "Test",
		"profile": "https://plus.google.com/10467329456789",
		"picture": "https://lh6.googleusercontent.com/-alknmlknzT_YQ/AAAAAAAAAAI/AAAAAAAAABU/4gNvDUeED14/photo.jpg",
		"email": "test@example.com",
		"email_verified": false,
		"gender": "male",
		"locale": "de",
		"hd": "example.com"
	  }`
)

func Test_Google_getUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equal(t, "secret", r.FormValue("access_token"))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(googleTestUserResponse))
	}))
	defer server.Close()

	googleUserinfoEndpoint = server.URL

	u, rawJSON, err := providerGoogle.GetUserInfo(TokenInfo{AccessToken: "secret"})
	NoError(t, err)
	Equal(t, "test@example.com", u.Sub)
	Equal(t, "test@example.com", u.Email)
	Equal(t, "https://lh6.googleusercontent.com/-alknmlknzT_YQ/AAAAAAAAAAI/AAAAAAAAABU/4gNvDUeED14/photo.jpg", u.Picture)
	Equal(t, "Testy Test", u.Name)
	Equal(t, "example.com", u.Domain)
	Equal(t, googleTestUserResponse, rawJSON)
}

func Test_Google_getUserInfoNegative(t *testing.T) {
	t.Run("server respond with no email", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Equal(t, "secret", r.FormValue("access_token"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(googleTestUserResponseNoEmail))
		}))
		defer server.Close()

		googleUserinfoEndpoint = server.URL

		_, rawJSON, err := providerGoogle.GetUserInfo(TokenInfo{AccessToken: "secret"})
		Error(t, err)
		Equal(t, "", rawJSON)
	})

	t.Run("server respond with unverified email", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Equal(t, "secret", r.FormValue("access_token"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(googleTestUserResponseUnverifiedEmail))
		}))
		defer server.Close()

		googleUserinfoEndpoint = server.URL

		_, rawJSON, err := providerGoogle.GetUserInfo(TokenInfo{AccessToken: "secret"})
		Error(t, err)
		Equal(t, "", rawJSON)
	})

	t.Run("server connection failed", func(t *testing.T) {
		_, rawJSON, err := providerGoogle.GetUserInfo(TokenInfo{AccessToken: "secret"})
		Error(t, err)
		Equal(t, "", rawJSON)
	})

	t.Run("server respond with error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Equal(t, "secret", r.FormValue("access_token"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		googleUserinfoEndpoint = server.URL

		_, rawJSON, err := providerGoogle.GetUserInfo(TokenInfo{AccessToken: "secret"})
		Error(t, err)
		Equal(t, "", rawJSON)
	})

	t.Run("server respond not with json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Equal(t, "secret", r.FormValue("access_token"))
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		googleUserinfoEndpoint = server.URL

		_, rawJSON, err := providerGoogle.GetUserInfo(TokenInfo{AccessToken: "secret"})
		Error(t, err)
		Equal(t, "", rawJSON)
	})

	t.Run("server respond with invalid json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Equal(t, "secret", r.FormValue("access_token"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		googleUserinfoEndpoint = server.URL

		_, rawJSON, err := providerGoogle.GetUserInfo(TokenInfo{AccessToken: "secret"})
		Error(t, err)
		Equal(t, "", rawJSON)
	})
}
