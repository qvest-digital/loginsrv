package osiam

import (
	"fmt"
	"strconv"
	"time"
)

// Token represents an osiam auth token
type Token struct {
	TokenType             string    `json:"token_type"`               // example "bearer"
	AccessToken           string    `json:"access_token"`             // example "79f479c2-c0d7-458a-8464-7eb887dbc943"
	RefreshToken          string    `json:"refresh_token"`            // example "3c7c4a87-dc91-4dd0-8ec8-d229a237a47c"
	ClientId              string    `json:"client_id"`                // example "example-client"
	UserName              string    `json:"user_name"`                // example "admin"
	Userid                string    `json:"user_id"`                  // example "84f6cffa-4505-48ec-a851-424160892283"
	Scope                 string    `json:"scope"`                    // example "ME"
	RefreshTokenExpiresAt Timestamp `json:"refresh_token_expires_at"` // example 1479309001813
	ExpiresAt             Timestamp `json:"expires_at"`               // example 1479251401814
	ExpiresIn             int       `json:"expires_in"`               // example 28795
}

type Timestamp struct {
	T time.Time
}

func (timestamp *Timestamp) UnmarshalJSON(b []byte) (err error) {
	i, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}
	timestamp.T = time.Unix(i, 0)
	return nil
}

func (timestamp *Timestamp) MarshalJSON() ([]byte, error) {
	if timestamp.T.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%d", timestamp.T.Unix())), nil
}

var nilTime = (time.Time{}).UnixNano()

func (timestamp *Timestamp) IsSet() bool {
	return timestamp.T.UnixNano() != nilTime
}
