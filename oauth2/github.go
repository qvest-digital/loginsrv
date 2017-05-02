package oauth2

import (
	"encoding/json"
	"fmt"
	"github.com/tarent/loginsrv/model"
	"io/ioutil"
	"net/http"
	"strings"
)

var githubApi = "https://api.github.com"

func init() {
	RegisterProvider(providerGithub)
}

type GithubUser struct {
	Login     string `json:"login,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
}

var providerGithub = Provider{
	Name:     "github",
	AuthURL:  "https://github.com/login/oauth/authorize",
	TokenURL: "https://github.com/login/oauth/access_token",
	GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
		gu := GithubUser{}
		url := fmt.Sprintf("%v/user?access_token=%v", githubApi, token.AccessToken)
		fmt.Println("url: ", url)
		resp, err := http.Get(url)
		if err != nil {
			return model.UserInfo{}, "", err
		}

		if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			return model.UserInfo{}, "", fmt.Errorf("wrong content-type on github get user info: %v", resp.Header.Get("Content-Type"))
		}

		if resp.StatusCode != 200 {
			return model.UserInfo{}, "", fmt.Errorf("got http status %v on github get user info", resp.StatusCode)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error reading github get user info: %v", err)
		}

		err = json.Unmarshal(b, &gu)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error parsing github get user info: %v", err)
		}

		return model.UserInfo{
			Sub:     gu.Login,
			Picture: gu.AvatarURL,
			Name:    gu.Name,
			Email:   gu.Email,
			Origin:  "github",
		}, string(b), nil
	},
}
