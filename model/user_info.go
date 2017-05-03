package model

type UserInfo struct {
	Sub     string `json:"sub"`
	Picture string `json:"picture,omitempty"`
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
	Origin  string `json:"origin,omitempty"`
	Expiry  int64  `json:"exp,omitempty"`
}

// this interface implementation
// lets us use the user info as Claim for jwt-go
func (u UserInfo) Valid() error {
	return nil
}
