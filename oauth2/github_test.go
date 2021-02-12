package oauth2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/stretchr/testify/assert"
)

var githubTestUserResponse = `{
  "login": "octocat",
  "id": 1,
  "avatar_url": "https://github.com/images/error/octocat_happy.gif",
  "gravatar_id": "",
  "url": "https://api.github.com/users/octocat",
  "html_url": "https://github.com/octocat",
  "followers_url": "https://api.github.com/users/octocat/followers",
  "following_url": "https://api.github.com/users/octocat/following{/other_user}",
  "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
  "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
  "subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
  "organizations_url": "https://api.github.com/users/octocat/orgs",
  "repos_url": "https://api.github.com/users/octocat/repos",
  "events_url": "https://api.github.com/users/octocat/events{/privacy}",
  "received_events_url": "https://api.github.com/users/octocat/received_events",
  "type": "User",
  "site_admin": false,
  "name": "monalisa octocat",
  "company": "GitHub",
  "blog": "https://github.com/blog",
  "location": "San Francisco",
  "email": "octocat@github.com",
  "hireable": false,
  "bio": "There once was...",
  "public_repos": 2,
  "public_gists": 1,
  "followers": 20,
  "following": 0,
  "created_at": "2008-01-14T04:33:35Z",
  "updated_at": "2008-01-14T04:33:35Z"
}`

func Test_Github_getUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equal(t, "token secret", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(githubTestUserResponse))
	}))
	defer server.Close()

	githubAPI = server.URL

	u, rawJSON, err := providerGithub.GetUserInfo(TokenInfo{AccessToken: "secret"})
	NoError(t, err)
	Equal(t, "octocat", u.Sub)
	Equal(t, "octocat@github.com", u.Email)
	Equal(t, "monalisa octocat", u.Name)
	Equal(t, githubTestUserResponse, rawJSON)
}

func Test_Github_getUserInfoNegative(t *testing.T) {
	t.Run("server connection failed", func(t *testing.T) {
		_, rawJSON, err := providerGithub.GetUserInfo(TokenInfo{AccessToken: "secret"})
		Error(t, err)
		Equal(t, "", rawJSON)
	})

	t.Run("server respond with error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Equal(t, "token secret", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		githubAPI = server.URL

		_, rawJSON, err := providerGithub.GetUserInfo(TokenInfo{AccessToken: "secret"})
		Error(t, err)
		Equal(t, "", rawJSON)
	})

	t.Run("server respond not with json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Equal(t, "token secret", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		githubAPI = server.URL

		_, rawJSON, err := providerGithub.GetUserInfo(TokenInfo{AccessToken: "secret"})
		Error(t, err)
		Equal(t, "", rawJSON)
	})

	t.Run("server respond with invalid json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Equal(t, "token secret", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		githubAPI = server.URL

		_, rawJSON, err := providerGithub.GetUserInfo(TokenInfo{AccessToken: "secret"})
		Error(t, err)
		Equal(t, "", rawJSON)
	})
}
