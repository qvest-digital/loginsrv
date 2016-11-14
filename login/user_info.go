package login

type UserInfo struct {
	Username string `json:"sub"`
}

// this interface implementation
// lets us use the user info as Claim for jwt-go
func (u UserInfo) Valid() error {
	return nil
}
