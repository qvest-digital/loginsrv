package oauth2

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	. "github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/model"
)

var linkedinTestUserResponse = `{
	"firstName": {
		"preferredLocale": {
			"country": "US",
			"language": "en"
		},
		"localized": {
			"en_US": "John"
		}
	},
	"lastName": {
		"preferredLocale": {
			"country": "US",
			"language": "en"
		},
		"localized": {
			"en_US": "Smith"
		}
	},
	"id": "john_smith",
	"profilePicture": {
		"displayImage~": {
			"elements": [
				{
					"identifiers": [{
						"identifier": "http://placehold.it/100x100"
					}]
				}
			]
		}
	}
}`

var linkedinTestPrimaryContactResponse = `{
	"elements": [
		{
			"handle": "",
			"type": "EMAIL",
			"handle~": {
				"emailAddress": "john@example.com"
			}
		}
	]
}`

func testGetAccessToken(h http.Header) string {
	if at := h.Get("Authorization"); at != "" {
		sep := strings.Split(at, "Bearer")
		if len(sep) != 2 {
			return ""
		}
		at = strings.TrimSpace(sep[1])
		return at
	}

	return ""
}

func Test_LinkedIn_getUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := testGetAccessToken(r.Header)
		if r.URL.Path == "/me" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(linkedinTestUserResponse))
		} else if r.URL.Path == "/clientAwareMemberHandles" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(linkedinTestPrimaryContactResponse))
		}
	}))
	defer server.Close()

	linkedinAPI = server.URL

	u, rawJSON, err := providerLinkedIn.GetUserInfo(TokenInfo{AccessToken: "secret"})
	NoError(t, err)
	Equal(t, "john_smith", u.Sub)
	Equal(t, "john@example.com", u.Email)
	Equal(t, "John Smith", u.Name)
	Equal(t, `{"user":`+linkedinTestUserResponse+`, "email":`+linkedinTestPrimaryContactResponse+`}`, rawJSON)
}

func Test_LinkedIn_getUserInfo_NoServer(t *testing.T) {
	linkedinAPI = "http://localhost"

	u, rawJSON, err := providerLinkedIn.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`connection refused$`), err.Error())
}

func Test_LinkedIn_getUserInfo_UserContentTypeNegative(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := testGetAccessToken(r.Header)
		if r.URL.Path == "/me" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte(linkedinTestUserResponse))
		} else if r.URL.Path == "/clientAwareMemberHandles" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(linkedinTestPrimaryContactResponse))
		}
	}))
	defer server.Close()

	linkedinAPI = server.URL

	u, rawJSON, err := providerLinkedIn.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`^wrong content-type on linkedin get user info`), err.Error())
}

func Test_LinkedIn_getUserInfo_PrimaryContentTypeNegative(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := testGetAccessToken(r.Header)
		if r.URL.Path == "/me" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(linkedinTestUserResponse))
		} else if r.URL.Path == "/clientAwareMemberHandles" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte(linkedinTestPrimaryContactResponse))
		}
	}))
	defer server.Close()

	linkedinAPI = server.URL

	u, rawJSON, err := providerLinkedIn.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`^wrong content-type on linkedin get user email`), err.Error())
}

func Test_LinkedIn_getUserInfo_UserStatusCodeNegative(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := testGetAccessToken(r.Header)
		if r.URL.Path == "/me" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(linkedinTestUserResponse))
		} else if r.URL.Path == "/clientAwareMemberHandles" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(linkedinTestPrimaryContactResponse))
		}
	}))
	defer server.Close()

	linkedinAPI = server.URL

	u, rawJSON, err := providerLinkedIn.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`^got http status [0-9]{3} on linkedin get user info`), err.Error())
}

func Test_LinkedIn_getUserInfo_PrimaryStatusCodeNegative(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := testGetAccessToken(r.Header)
		if r.URL.Path == "/me" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(linkedinTestUserResponse))
		} else if r.URL.Path == "/clientAwareMemberHandles" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(linkedinTestPrimaryContactResponse))
		}
	}))
	defer server.Close()

	linkedinAPI = server.URL

	u, rawJSON, err := providerLinkedIn.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`^got http status [0-9]{3} on linkedin get user email`), err.Error())
}

func Test_LinkedIn_getUserInfo_UserJSONNegative(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := testGetAccessToken(r.Header)
		if r.URL.Path == "/me" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte("[]"))
		} else if r.URL.Path == "/clientAwareMemberHandles" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(linkedinTestPrimaryContactResponse))
		}
	}))
	defer server.Close()

	linkedinAPI = server.URL

	u, rawJSON, err := providerLinkedIn.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`^error parsing linkedin get user info`), err.Error())
}

func Test_LinkedIn_getUserInfo_PrimaryJSONNegative(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := testGetAccessToken(r.Header)
		if r.URL.Path == "/me" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(linkedinTestUserResponse))
		} else if r.URL.Path == "/clientAwareMemberHandles" {
			Equal(t, "secret", accessToken)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte("[]"))
		}
	}))
	defer server.Close()

	linkedinAPI = server.URL

	u, rawJSON, err := providerLinkedIn.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`^error parsing linkedin get user email`), err.Error())
}
