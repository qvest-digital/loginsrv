package oauth2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tarent/loginsrv/model"
)

var discordAPI = "https://discordapp.com/api"
var discordCDN = "https://cdn.discordapp.com"

func init() {
	RegisterProvider(providerDiscord)
}

// DiscordUser is used for parsing the github response
type DiscordUser struct {
	ID            string `json:"id,omitempty"`
	Username      string `json:"username,omitempty"`
	Discriminator string `json:"discriminator,omitempty"`
	AvatarHash    string `json:"avatar,omitempty"`
	MFAEnabled    bool   `json:"mfa_enabled,omitempty"`
	Locale        string `json:"locale,omitempty"`
	Verified      bool   `json:"verified,omitempty"`
	Email         string `json:"email,omitempty"`
	Flags         int    `json:"flags,omitempty"`
	PremiumType   int    `json:"premium_type,omitempty"`
}

// DiscordGuild is a partial guild object returned by the /user/guilds endpoint
type DiscordGuild struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	IconHash    string `json:"icon,omitempty"`
	Owner       bool   `json:"owner,omitempty"`
	Permissions int    `json:"permissions,omitempty"`
}

func discordAPIRequest(endpoint, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/%v", discordAPI, endpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return nil, fmt.Errorf("wrong content-type: %v", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("got http status %v", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %v", err)
	}

	return b, nil
}

var providerDiscord = Provider{
	Name:          "discord",
	AuthURL:       "https://discordapp.com/api/oauth2/authorize?prompt=none",
	TokenURL:      "https://discordapp.com/api/oauth2/token",
	DefaultScopes: "identify email guilds",
	GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
		du := DiscordUser{}
		dg := []DiscordGuild{}
		// Get user info
		raw, err := discordAPIRequest("/users/@me", token.AccessToken)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error getting discord user info: %v", err)
		}
		err = json.Unmarshal(raw, &du)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error parsing discord get user info: %v", err)
		}

		// Get user's guilds (servers)
		raw, err = discordAPIRequest("/users/@me/guilds", token.AccessToken)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error getting discord user guilds: %v", err)
		}
		err = json.Unmarshal(raw, &dg)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error parsing discord guilds: %v", err)
		}

		var guilds []string
		for _, g := range dg {
			guilds = append(guilds, g.ID)
		}

		return model.UserInfo{
			Sub:     fmt.Sprintf("%v#%v", du.Username, du.Discriminator),
			Picture: fmt.Sprintf("%v/avatars/%v/%v.png", discordCDN, du.ID, du.AvatarHash),
			Name:    du.Username,
			Email:   du.Email,
			Origin:  "discord",
			Groups:  guilds,
		}, string(raw), nil
	},
}
