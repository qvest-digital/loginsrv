package oauth2

import (
	. "github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var facebookTestUserResponse = `{
    "id": "23456789012345678",
    "name": "Facebook User",
    "picture": {
        "data": {
            "height": 100,
            "is_silhouette": false,
            "url": "https://scontent.xx.fbcdn.net/v/t1.0-1/p100x100/example_facebook_image.jpg",
            "width": 100
        }
    },
    "email": "facebookuser@facebook.com"
}`

func Test_Facebook_getUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equal(t, "secret", r.FormValue("access_token"))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(facebookTestUserResponse))
	}))
	defer server.Close()

	facebookAPI = server.URL

	u, rawJSON, err := providerfacebook.GetUserInfo(TokenInfo{AccessToken: "secret"})
	NoError(t, err)
	Equal(t, "23456789012345678", u.Sub)
	Equal(t, "facebookuser@facebook.com", u.Email)
	Equal(t, "Facebook User", u.Name)
	Equal(t, facebookTestUserResponse, rawJSON)
}

func Test_Facebook_getUserInfo_WrongContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equal(t, "secret", r.FormValue("access_token"))
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Write([]byte(facebookTestUserResponse))
	}))
	defer server.Close()

	facebookAPI = server.URL

	_, _, err := providerfacebook.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Error(t, err)
}

func Test_Facebook_getUserInfo_WrongStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equal(t, "secret", r.FormValue("access_token"))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(facebookTestUserResponse))
	}))
	defer server.Close()

	facebookAPI = server.URL

	_, _, err := providerfacebook.GetUserInfo(TokenInfo{AccessToken: "secret"})
	Error(t, err)
}
