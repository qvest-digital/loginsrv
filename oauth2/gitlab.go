package oauth2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/tarent/loginsrv/model"
)

const defaultGitlabURL = "https://gitlab.com"

func init() {
	gitlabURL, ok := os.LookupEnv("GITLAB_URL")
	if !ok {
		gitlabURL = defaultGitlabURL
	}
	provider := MakeGitlabProvider(gitlabURL)
	RegisterProvider(provider)
}

// GitlabUserInfo is used for parsing the gitlab response
type GitlabUserInfo struct {
	model.UserInfo
	Sub string `json:"nickname"`
}

func (i GitlabUserInfo) toUserInfo() model.UserInfo {
	res := model.UserInfo{
		Sub:       i.Sub,
		Picture:   i.Picture,
		Name:      i.Name,
		Email:     i.Email,
		Origin:    "gitlab",
		Expiry:    i.Expiry,
		Refreshes: i.Refreshes,
		Groups:    i.Groups,
	}

	email := i.Email
	emailComponents := strings.Split(email, "@")
	if len(emailComponents) == 2 {
		res.Domain = emailComponents[1]
	}
	return res
}

// MakeGitlabProvider make's a gitlab provider with the given url
func MakeGitlabProvider(gitlabURL string) Provider {
	return Provider{
		Name:          "gitlab",
		AuthURL:       gitlabURL + "/oauth/authorize",
		TokenURL:      gitlabURL + "/oauth/token",
		DefaultScopes: "email openid",
		GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
			url := fmt.Sprintf("%s/oauth/userinfo?access_token=%v", gitlabURL, token.AccessToken)

			var info GitlabUserInfo

			var respUser *http.Response
			respUser, err := http.Get(url)
			if err != nil {
				return model.UserInfo{}, "", err
			}
			defer respUser.Body.Close()

			if !strings.Contains(respUser.Header.Get("Content-Type"), "application/json") {
				return model.UserInfo{}, "", fmt.Errorf("wrong content-type on gitlab get user info: %v", respUser.Header.Get("Content-Type"))
			}

			if respUser.StatusCode != 200 {
				return model.UserInfo{}, "", fmt.Errorf("got http status %v on gitlab get user info", respUser.StatusCode)
			}

			b, err := ioutil.ReadAll(respUser.Body)
			if err != nil {
				return model.UserInfo{}, "", fmt.Errorf("error reading gitlab get user info: %v", err)
			}

			err = json.Unmarshal(b, &info)
			if err != nil {
				return model.UserInfo{}, "", fmt.Errorf("error parsing gitlab get user info: %v", err)
			}

			return info.toUserInfo(), "", nil
		},
	}
}
