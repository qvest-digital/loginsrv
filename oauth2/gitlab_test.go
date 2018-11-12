package oauth2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/stretchr/testify/assert"
)

var gitlabTestUserResponse = `{
	"id": 1,
	"username": "john_smith",
	"email": "john@example.com",
	"name": "John Smith",
	"state": "active",
	"avatar_url": "http://localhost:3000/uploads/user/avatar/1/index.jpg",
	"web_url": "http://localhost:3000/john_smith",
	"created_at": "2012-05-23T08:00:58Z",
	"bio": null,
	"location": null,
	"public_email": "john@example.com",
	"skype": "",
	"linkedin": "",
	"twitter": "",
	"website_url": "",
	"organization": "",
	"last_sign_in_at": "2012-06-01T11:41:01Z",
	"confirmed_at": "2012-05-23T09:05:22Z",
	"theme_id": 1,
	"last_activity_on": "2012-05-23",
	"color_scheme_id": 2,
	"projects_limit": 100,
	"current_sign_in_at": "2012-06-02T06:36:55Z",
	"identities": [
	  {"provider": "github", "extern_uid": "2435223452345"},
	  {"provider": "bitbucket", "extern_uid": "john_smith"},
	  {"provider": "google_oauth2", "extern_uid": "8776128412476123468721346"}
	],
	"can_create_group": true,
	"can_create_project": true,
	"two_factor_enabled": true,
	"external": false,
	"private_profile": false
  }`

var gitlabTestGroupsResponse = `[
	{
	  "id": 1,
	  "web_url": "https://gitlab.com/groups/example",
	  "name": "example",
	  "path": "example",
	  "description": "",
	  "visibility": "private",
	  "lfs_enabled": true,
	  "avatar_url": null,
	  "request_access_enabled": true,
	  "full_name": "example",
	  "full_path": "example",
	  "parent_id": null,
	  "ldap_cn": null,
	  "ldap_access": null
	},
	{
		"id": 2,
		"web_url": "https://gitlab.com/groups/example/subgroup",
		"name": "subgroup",
		"path": "subgroup",
		"description": "",
		"visibility": "private",
		"lfs_enabled": true,
		"avatar_url": null,
		"request_access_enabled": true,
		"full_name": "example / subgroup",
		"full_path": "example/subgroup",
		"parent_id": null,
		"ldap_cn": null,
		"ldap_access": null
	}
]`

func Test_Gitlab_getUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			Equal(t, "secret", r.FormValue("access_token"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(gitlabTestUserResponse))
		} else if r.URL.Path == "/groups" {
			Equal(t, "secret", r.FormValue("access_token"))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(gitlabTestGroupsResponse))
		}
	}))
	defer server.Close()

	gitlabAPI = server.URL

	u, rawJSON, err := providerGitlab.GetUserInfo(TokenInfo{AccessToken: "secret"})
	NoError(t, err)
	Equal(t, "john_smith", u.Sub)
	Equal(t, "john@example.com", u.Email)
	Equal(t, "John Smith", u.Name)
	Equal(t, []string{"example", "example/subgroup"}, u.Groups)
	Equal(t, `{"user":`+gitlabTestUserResponse+`,"groups":`+gitlabTestGroupsResponse+`}`, rawJSON)
}
