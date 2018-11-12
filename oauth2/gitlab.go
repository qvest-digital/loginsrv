package oauth2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tarent/loginsrv/model"
)

var gitlabAPI = "https://gitlab.com/api/v4"

func init() {
	RegisterProvider(providerGitlab)
}

// GitlabUser is used for parsing the gitlab response
type GitlabUser struct {
	Username  string `json:"username,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
}

type GitlabGroup struct {
	FullPath string `json:"full_path,omitempty"`
}

var providerGitlab = Provider{
	Name:     "gitlab",
	AuthURL:  "https://gitlab.com/oauth/authorize",
	TokenURL: "https://gitlab.com/oauth/token",
	GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
		gu := GitlabUser{}
		url := fmt.Sprintf("%v/user?access_token=%v", gitlabAPI, token.AccessToken)
		resp, err := http.Get(url)
		if err != nil {
			return model.UserInfo{}, "", err
		}

		if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			return model.UserInfo{}, "", fmt.Errorf("wrong content-type on gitlab get user info: %v", resp.Header.Get("Content-Type"))
		}

		if resp.StatusCode != 200 {
			return model.UserInfo{}, "", fmt.Errorf("got http status %v on gitlab get user info", resp.StatusCode)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error reading gitlab get user info: %v", err)
		}

		err = json.Unmarshal(b, &gu)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error parsing gitlab get user info: %v", err)
		}

		gg := []*GitlabGroup{}
		url = fmt.Sprintf("%v/groups?access_token=%v", gitlabAPI, token.AccessToken)
		resp, err = http.Get(url)
		if err != nil {
			return model.UserInfo{}, "", err
		}

		if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			return model.UserInfo{}, "", fmt.Errorf("wrong content-type on gitlab get groups info: %v", resp.Header.Get("Content-Type"))
		}

		if resp.StatusCode != 200 {
			return model.UserInfo{}, "", fmt.Errorf("got http status %v on gitlab get groups info", resp.StatusCode)
		}

		g, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error reading gitlab get groups info: %v", err)
		}

		err = json.Unmarshal(g, &gg)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error parsing gitlab get groups info: %v", err)
		}

		groups := make([]string, len(gg))
		for i := 0; i < len(gg); i++ {
			groups[i] = gg[i].FullPath
		}

		return model.UserInfo{
			Sub:     gu.Username,
			Picture: gu.AvatarURL,
			Name:    gu.Name,
			Email:   gu.Email,
			Groups:  groups,
			Origin:  "gitlab",
		}, `{"user":` + string(b) + `,"groups":` + string(g) + `}`, nil
	},
}
