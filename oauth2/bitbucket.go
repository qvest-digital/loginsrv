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
// Using the avatar url to be able to fetch 128px image. By default BitbucketAPI return 32px image.
var bitbucketAvatarURL = "https://bitbucket.org/account/%v/avatar/128/"

func init() {
	RegisterProvider(providerBitbucket)
}

// BitbucketUser is used for parsing the github response
type BitbucketUser struct {
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
}

// emails is used to parse user email information
type emails struct {
	Page    int `json:"page,omitempty"`
	PageLen int `json:"pagelen,omitempty"`
	Size    int `json:"size,omitempty"`
	Values  []email
}

// email used to parse one user's email
type email struct {
	Email       string `json:"email,omitempty"`
	IsConfirmed bool   `json:"is_confirmed,omitempty"`
	IsPrimary   bool   `json:"is_primary,omitempty"`
	Links struct {
		Self struct {
			Href string
		}
	} `json:"links,omitempty"`
	Type string `json:"type,omitempty"`
}

// GetPrimaryEmailAddress retrieve the primary email address of the user
func (e *emails) GetPrimaryEmailAddress() string {
	for _, val := range e.Values {
		if val.IsPrimary {
			return val.Email
		}
	}
	return ""
}

// getBitbucketEmails Retrieves bitbucket user emails hiting the Bitbucket API emails service
func getBitbucketEmails(token TokenInfo) (emails, error) {
	emailUrl := fmt.Sprintf("%v/user/emails?access_token=%v", bitbucketAPI, token.AccessToken)
	userEmails := emails{}
	resp, err := http.Get(emailUrl)

	if err != nil {
		return emails{}, err
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return emails{}, fmt.Errorf("wrong content-type on bitbucket get user emails: %v", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		return emails{}, fmt.Errorf("got http status %v on bitbucket get user emails", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return emails{}, fmt.Errorf("error reading bitbucket get user emails: %v", err)
	}

	err = json.Unmarshal(b, &userEmails)

	if err != nil {
		return emails{}, fmt.Errorf("error parsing bitbucket get user emails: %v", err)
	}

	return userEmails, nil
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

		userEmails, err := getBitbucketEmails(token)

		return model.UserInfo{
			Sub:     gu.Username,
			Picture: fmt.Sprintf(bitbucketAvatarURL, gu.Username),
			Name:    gu.DisplayName,
			Email:   userEmails.GetPrimaryEmailAddress(),
			Origin:  "bitbucket",
		}, string(b), nil
	},
}
