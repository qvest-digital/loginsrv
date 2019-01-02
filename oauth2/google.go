package oauth2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tarent/loginsrv/model"
)

//var googleAPI = "https://www.googleapis.com/plus/v1"

var googleUserinfoEndpoint = "https://www.googleapis.com/oauth2/v3/userinfo"

func init() {
	RegisterProvider(providerGoogle)
}

type GoogleUser struct {
	Name               string `json:"name"`
	Email              string `json:"email"`
	EmailVerified      bool   `json:"email_verified"`
	Picture            string `json:"picture"`
	HostedGsuiteDomain string `json:"hd"`
}

var providerGoogle = Provider{
	Name:          "google",
	AuthURL:       "https://accounts.google.com/o/oauth2/v2/auth",
	TokenURL:      "https://www.googleapis.com/oauth2/v4/token",
	DefaultScopes: "email profile",
	GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
		gu := GoogleUser{}
		url := fmt.Sprintf("%v?access_token=%v", googleUserinfoEndpoint, token.AccessToken)
		resp, err := http.Get(url)

		if err != nil {
			return model.UserInfo{}, "", err
		}
		defer resp.Body.Close()

		if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			return model.UserInfo{}, "", fmt.Errorf("wrong content-type on google get user info: %v", resp.Header.Get("Content-Type"))
		}

		if resp.StatusCode != 200 {
			return model.UserInfo{}, "", fmt.Errorf("got http status %v on google get user info", resp.StatusCode)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error reading google get user info: %v", err)
		}

		err = json.Unmarshal(b, &gu)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error parsing google get user info: %v", err)
		}

		if len(gu.Email) == 0 {
			return model.UserInfo{}, "", fmt.Errorf("invalid google response: no email address returned.")
		}

		if !gu.EmailVerified {
			return model.UserInfo{}, "", fmt.Errorf("invalid google response: users email address not verified by google.")
		}

		return model.UserInfo{
			Sub:     gu.Email,
			Picture: gu.Picture,
			Name:    gu.Name,
			Email:   gu.Email,
			Origin:  "google",
			Domain:  gu.HostedGsuiteDomain,
		}, string(b), nil
	},
}
