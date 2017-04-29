package oauth2

var githubApi = "https://api.github.com/v3"

func init() {
	RegisterProvider(providerGithub)
}

var providerGithub = Provider{
	Name:     "github",
	AuthURL:  "https://github.com/login/oauth/authorize",
	TokenURL: "https://github.com/login/oauth/access_token",
	GetUserInfo: func(token TokenInfo) (map[string]string, error) {
		//http.Get(fmt.Sprintf("%v/user?access_token=%v"), githubApi, token.AccessToken)
		// https://developer.github.com/v3/users/
		return map[string]string{"username": "demo"}, nil
	},
}
