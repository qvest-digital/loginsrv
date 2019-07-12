package oauth2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	. "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var discordTestUserResponse = `{
	"id": "80351110224678912",
	"username": "Nelly",
	"discriminator": "1337",
	"avatar": "8342729096ea3675442027381ff50dfe",
	"verified": true,
	"email": "nelly@discordapp.com",
	"flags": 64,
	"premium_type": 1
  }`

var discordTestUserGuildsResponse = `[
	{
		"id": "80351110224678912",
		"name": "1337 Krew",
		"icon": "8342729096ea3675442027381ff50dfe",
		"owner": true,
		"permissions": 36953089
	}
]`

type DiscordTestSuite struct {
	suite.Suite
	Server *httptest.Server
}

func (suite *DiscordTestSuite) SetupTest() {
	r := mux.NewRouter()

	r.HandleFunc("/users/@me", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(r.Header.Get("Authentication"), "Bearer secret")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(discordTestUserResponse))
	})

	r.HandleFunc("/users/@me/guilds", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(r.Header.Get("Authentication"), "Bearer secret")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(discordTestUserGuildsResponse))
	})

	suite.Server = httptest.NewServer(r)
}

func (suite *DiscordTestSuite) Test_Discord_getUserInfo(t *testing.T) {
	discordAPI = suite.Server.URL

	u, rawJSON, err := providerDiscord.GetUserInfo(TokenInfo{AccessToken: "secret"})
	NoError(t, err)
	Equal(t, "Nelly#1337", u.Sub)
	Equal(t, "nelly@discordapp.com", u.Email)
	Equal(t, "Nelly", u.Name)
	Equal(t, []string{"80351110224678912"}, u.Groups)
	Equal(t, discordTestUserResponse, rawJSON)
}
