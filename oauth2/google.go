package oauth2

import (
	"encoding/json"
	"fmt"
	"github.com/tarent/loginsrv/model"
	"io/ioutil"
	"net/http"
	"strings"
)

var googleAPI = "https://www.googleapis.com/plus/v1"

func init() {
	RegisterProvider(providerGoogle)
}

type GoogleUser struct {
	DisplayName string
	Emails      []struct {
		Value string
	}
	Image struct {
		Url string
	}
}

var providerGoogle = Provider{
	Name:     "google",
	AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
	TokenURL: "https://www.googleapis.com/oauth2/v4/token",
	GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
		gu := GoogleUser{}
		url := fmt.Sprintf("%v/people/me?alt=json&access_token=%v", googleAPI, token.AccessToken)
		resp, err := http.Get(url)

		if err != nil {
			return model.UserInfo{}, "", err
		}

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

		if len(gu.Emails) == 0 {
			return model.UserInfo{}, "", fmt.Errorf("invalid google response: no email address returned.")
		}

		return model.UserInfo{
			Sub:     gu.Emails[0].Value,
			Picture: gu.Image.Url,
			Name:    gu.DisplayName,
			Email:   gu.Emails[0].Value,
			Origin:  "google",
		}, string(b), nil
	},
}
