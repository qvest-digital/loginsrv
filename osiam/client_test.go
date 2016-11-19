package osiam

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetTokenByPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(osiamMockHandler))
	defer server.Close()

	client := NewClient(server.URL, "example-client", "secret")
	authenticated, token, err := client.GetTokenByPassword("admin", "koala")

	assert.NoError(t, err)
	assert.True(t, authenticated)
	assert.Equal(t,
		&Token{
			Userid:                "84f6cffa-4505-48ec-a851-424160892283",
			ExpiresIn:             28493,
			ExpiresAt:             Timestamp{time.Unix(1479600891424, 0)},
			RefreshToken:          "15b22304-f838-48c2-9c40-18bf285060a6",
			ClientId:              "example-client",
			RefreshTokenExpiresAt: Timestamp{time.Unix(1479575304893, 0)},
			UserName:              "admin",
			TokenType:             "bearer",
			AccessToken:           "59f39ef8-1dc3-4c0d-8dea-c9597ef0a8ef",
			Scope:                 "ME",
		},
		token)
	assert.True(t, len(token.RefreshToken) > 0)
	assert.True(t, token.ExpiresIn > 0)
}

func TestClient_GetTokenByPasswordErrorCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(osiamMockHandler))
	defer server.Close()

	// wrong credentials
	client := NewClient(server.URL, "example-client", "secret")
	authenticated, _, err := client.GetTokenByPassword("admin", "XXX")
	assert.NoError(t, err)
	assert.False(t, authenticated)

	// wrong url -> 404
	client = NewClient(server.URL+"/Foo", "example-client", "secret")
	_, _, err = client.GetTokenByPassword("admin", "koala")
	assert.Error(t, err)

	// wrong client secret
	client = NewClient(server.URL, "example-client", "XXX")
	_, _, err = client.GetTokenByPassword("admin", "koala")
	assert.Error(t, err)

	// invalid url
	client = NewClient("://", "example-client", "secret")
	_, _, err = client.GetTokenByPassword("admin", "koala")
	assert.Error(t, err)

}

func TestClient_GetTokenByPasswordInvalidJson(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(200)
		fmt.Fprintf(w, "{...")
		return
	}))
	defer server.Close()

	client := NewClient(server.URL, "example-client", "secret")
	_, _, err := client.GetTokenByPassword("admin", "koala")
	assert.Error(t, err)
}

func TestClient_GetTokenByPasswordUnknownError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(201)
		fmt.Fprintf(w, `{"error":"foo bar","error_description":"some message!"}`)
		return
	}))
	defer server.Close()

	client := NewClient(server.URL, "example-client", "secret")
	_, _, err := client.GetTokenByPassword("admin", "koala")
	assert.Error(t, err)
}

func TestClient_GetTokenByPasswordNoServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	server.Close()

	client := NewClient(server.URL, "example-client", "secret")
	_, _, err := client.GetTokenByPassword("admin", "koala")
	assert.Error(t, err)
}

func osiamMockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(400)
		fmt.Fprintf(w, `Method not supported`)
		return
	}

	if r.URL.Path != "/oauth/token" {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Not found %q", r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	if r.Header.Get("Authorization") != "Basic ZXhhbXBsZS1jbGllbnQ6c2VjcmV0" {
		w.WriteHeader(401)
		fmt.Fprintf(w, `{"timestamp":1479572095876,"status":401,"error":"Unauthorized","message":"Full authentication is required to access this resource","path":"/oauth/token"}`)
		return
	}
	b, _ := ioutil.ReadAll(r.Body)

	if string(b) != "grant_type=password&username=admin&password=koala&scope=ME" {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error":"invalid_grant","error_description":"some message!"}`)
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, `{
                "user_id" : "84f6cffa-4505-48ec-a851-424160892283",
                "expires_in" : 28493,
                "expires_at" : 1479600891424,
                "refresh_token" : "15b22304-f838-48c2-9c40-18bf285060a6",
                "client_id" : "example-client",
                "refresh_token_expires_at" : 1479575304893,
                "user_name" : "admin",
                "token_type" : "bearer",
                "access_token" : "59f39ef8-1dc3-4c0d-8dea-c9597ef0a8ef",
                "scope" : "ME"}`)
}
