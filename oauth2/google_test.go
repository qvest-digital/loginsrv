package oauth2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/stretchr/testify/assert"
)

var googleTestUserResponse = `{
  "kind": "plus#person",
  "etag": "\"XX\"",
  "gender": "male",
  "emails": [
    {
      "value": "test@gmail.com",
      "type": "account"
    }
  ],
  "objectType": "person",
  "id": "1",
  "displayName": "Testy Test",
  "name": {
    "familyName": "Test",
    "givenName": "Testy"
  },
  "url": "https://plus.google.com/X",
  "image": {
    "url": "https://lh3.googleusercontent.com/X/X/X/X/photo.jpg?sz=50",
    "isDefault": true
  },
  "isPlusUser": true,
  "circledByCount": 0,
  "verified": false,
  "domain": "gmail.com"
}`

func Test_Google_getUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equal(t, "secret", r.FormValue("access_token"))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(googleTestUserResponse))
	}))
	defer server.Close()

	googleAPI = server.URL

	u, rawJSON, err := providerGoogle.GetUserInfo(TokenInfo{AccessToken: "secret"})
	NoError(t, err)
	Equal(t, "test@gmail.com", u.Sub)
	Equal(t, "test@gmail.com", u.Email)
	Equal(t, "Testy Test", u.Name)
	Equal(t, "gmail.com", u.Domain)
	Equal(t, googleTestUserResponse, rawJSON)
}
