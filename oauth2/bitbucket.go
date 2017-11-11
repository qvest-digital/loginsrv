package oauth2

import (
	"encoding/json"
	"fmt"
	"github.com/tarent/loginsrv/model"
	"io/ioutil"
	"net/http"
	"strings"
)

var bitbucketAPI = "https://api.bitbucket.org/2.0"
// Using the avatar url to be able to fetch 128px image. By default Bitbucket API return 32px image.
var bitbucketAvatarURL = "https://bitbucket.org/account/%v/avatar/128/"

func init() {
	RegisterProvider(providerBitbucket)
}

// BitbucketUser is used for parsing the github response
type BitbucketUser struct {
	Username     string `json:"username,omitempty"`
	DisplayName      string `json:"display_name,omitempty"`
	Email     string `json:"email,omitempty"`
}

var providerBitbucket = Provider{
	Name:     "bitbucket",
	AuthURL:  "https://bitbucket.org/site/oauth2/authorize",
	TokenURL: "https://bitbucket.org/site/oauth2/access_token",
	GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
		gu := BitbucketUser{}
		url := fmt.Sprintf("%v/user?access_token=%v", bitbucketAPI, token.AccessToken)
		resp, err := http.Get(url)
		if err != nil {
			return model.UserInfo{}, "", err
		}

		if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			return model.UserInfo{}, "", fmt.Errorf("wrong content-type on bitbucket get user info: %v", resp.Header.Get("Content-Type"))
		}

		if resp.StatusCode != 200 {
			return model.UserInfo{}, "", fmt.Errorf("got http status %v on bitbucket get user info", resp.StatusCode)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error reading bitbucket get user info: %v", err)
		}

		err = json.Unmarshal(b, &gu)

		if err != nil {
			return model.UserInfo{}, "", fmt.Errorf("error parsing bitbucket get user info: %v", err)
		}

		return model.UserInfo{
			Sub:     gu.Username,
			Picture: fmt.Sprintf(bitbucketAvatarURL, gu.Username),
			Name:    gu.DisplayName,
			Email:   gu.Email,
			Origin:  "bitbucket",
		}, string(b), nil
	},
}
