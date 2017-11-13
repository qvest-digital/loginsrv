package oauth2

import (
	. "github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
	"encoding/json"
)

var bitbucketTestUserResponse = `{
  "created_on": "2011-12-20T16:34:07.132459+00:00",
  "display_name": "tutorials account",
  "is_staff": false,
  "links": {
    "avatar": {
      "href": "https://bitbucket.org/account/tutorials/avatar/32/"
    },
    "followers": {
      "href": "https://api.bitbucket.org/2.0/users/tutorials/followers"
    },
    "following": {
      "href": "https://api.bitbucket.org/2.0/users/tutorials/following"
    },
    "hooks": {
      "href": "https://api.bitbucket.org/2.0/users/tutorials/hooks"
    },
    "html": {
      "href": "https://bitbucket.org/tutorials/"
    },
    "repositories": {
      "href": "https://api.bitbucket.org/2.0/repositories/tutorials"
    },
    "self": {
      "href": "https://api.bitbucket.org/2.0/users/tutorials"
    },
    "snippets": {
      "href": "https://api.bitbucket.org/2.0/snippets/tutorials"
    }
  },
  "location": null,
  "type": "user",
  "username": "tutorials",
  "uuid": "{c788b2da-b7a2-404c-9e26-d3f077557007}",
  "website": "https://tutorials.bitbucket.org/"
}`

var bitbucketTestUserEmailResponse = `{
  "page": 1,
  "pagelen": 10,
  "size": 1,
  "values": [
    {
      "email": "tutorials@bitbucket.com",
      "is_confirmed": true,
      "is_primary": true,
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/user/emails/tutorials@bitbucket.com"
        }
      },
      "type": "email"
    },
	{
      "email": "anotheremail@bitbucket.com",
      "is_confirmed": false,
      "is_primary": false,
      "links": {
        "self": {
          "href": "https://api.bitbucket.org/2.0/user/emails/anotheremail@bitbucket.com"
        }
      },
      "type": "email"
    }
  ]
}`

// BitbucketTestSuite Model for the bitbucket test suite
type BitbucketTestSuite struct {
	suite.Suite
	Server *httptest.Server
}

// SetupTest a method that will be run before any method of this suite. It setups a mock server for bitbucket API
func (suite *BitbucketTestSuite) SetupTest() {
	r := mux.NewRouter()

	userHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Equal("secret", r.FormValue("access_token"))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(bitbucketTestUserResponse))
	})

	emailHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Equal("secret", r.FormValue("access_token"))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(bitbucketTestUserEmailResponse))
	})

	r.HandleFunc("/user", userHandler)
	r.HandleFunc("/user/emails", emailHandler)

	suite.Server = httptest.NewServer(r)
}

// Test_Bitbucket_getUserInfo Tests Bitbucket provider returns the expected information
func (suite *BitbucketTestSuite) Test_Bitbucket_getUserInfo() {

	bitbucketAPI = suite.Server.URL

	u, rawJSON, err := providerBitbucket.GetUserInfo(TokenInfo{AccessToken: "secret"})
	suite.NoError(err)
	suite.Equal("tutorials", u.Sub)
	suite.Equal("tutorials@bitbucket.com", u.Email)
	suite.Equal("tutorials account", u.Name)
	suite.Equal(bitbucketTestUserResponse, rawJSON)
}

// Test_Bitbucket_getPrimaryEmailAddress Tests the returned primary email is the expected email
func (suite *BitbucketTestSuite) Test_Bitbucket_getPrimaryEmailAddress()  {
	userEmails := emails{}
	err := json.Unmarshal([]byte(bitbucketTestUserEmailResponse), &userEmails)
	suite.NoError(err)
	suite.Equal("tutorials@bitbucket.com", userEmails.GetPrimaryEmailAddress())
}

// Test_Bitbucket_Suite Runs the entire suite for Bitbucket
func Test_Bitbucket_Suite(t *testing.T) {
	suite.Run(t, new(BitbucketTestSuite))
}
