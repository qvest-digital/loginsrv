package oauth2

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	. "github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/model"
)

var gitlabTestUserResponse = `{
	"sub": "1234567",
	"sub_legacy": "e7d33ae82f57ec69415af7dadb01f7b047ad62fd3a7d7957f20d6ceb7643331a",
	"name": "John Smith",
	"nickname": "john_smith",
	"email": "john@example.com",
	"email_verified": true,
	"profile": "https://gitlab.com/jsmith",
	"picture": "https://secure.gravatar.com/avatar/b92a7c822a31fa55c65186f9be24841e?s=80&d=identicon",
	"groups": [
	  "example",
	  "example/subgroup"
	]
  }`

func Test_Gitlab_getUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/userinfo" {
			Equal(t, "secret", r.FormValue("access_token"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(gitlabTestUserResponse))
		}
	}))
	defer server.Close()

	providerGitlab := MakeGitlabProvider(server.URL)

	u, _, err := providerGitlab.GetUserInfo(TokenInfo{AccessToken: "secret"})
	NoError(t, err)
	Equal(t, "john_smith", u.Sub)
	Equal(t, "john@example.com", u.Email)
	Equal(t, "John Smith", u.Name)
	Equal(t, []string{"example", "example/subgroup"}, u.Groups)
}

func Test_Gitlab_getUserInfo_NoServer(t *testing.T) {
	providerGitlab := MakeGitlabProvider("http://localhost:8290")

	u, rawJSON, err := providerGitlab.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`connection refused$`), err.Error())
}

func Test_Gitlab_getUserInfo_UserContentTypeNegative(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/userinfo" {
			Equal(t, "secret", r.FormValue("access_token"))
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte(gitlabTestUserResponse))
		}
	}))
	defer server.Close()

	providerGitlab := MakeGitlabProvider(server.URL)

	u, rawJSON, err := providerGitlab.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`^wrong content-type on gitlab get user info`), err.Error())
}

func Test_Gitlab_getUserInfo_UserStatusCodeNegative(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/userinfo" {
			Equal(t, "secret", r.FormValue("access_token"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(gitlabTestUserResponse))
		}
	}))
	defer server.Close()

	providerGitlab := MakeGitlabProvider(server.URL)

	u, rawJSON, err := providerGitlab.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`^got http status [0-9]{3} on gitlab get user info`), err.Error())
}

func Test_Gitlab_getUserInfo_UserReadNegative(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/userinfo" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(""))

			// hijack the connection to force close
			hj, _ := w.(http.Hijacker)
			conn, _, _ := hj.Hijack()
			conn.Close();
		}
	}))
	defer server.Close()

	providerGitlab := MakeGitlabProvider(server.URL)

	u, rawJSON, err := providerGitlab.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`^error reading gitlab get user info`), err.Error())
}

func Test_Gitlab_getUserInfo_UserJSONNegative(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/userinfo" {
			Equal(t, "secret", r.FormValue("access_token"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte("[]"))
		}
	}))
	defer server.Close()

	providerGitlab := MakeGitlabProvider(server.URL)

	u, rawJSON, err := providerGitlab.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Equal(t, model.UserInfo{}, u)
	Empty(t, rawJSON)
	Error(t, err)
	Regexp(t, regexp.MustCompile(`^error parsing gitlab get user info`), err.Error())
}