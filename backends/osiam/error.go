package osiam

import (
	"encoding/json"
)

// OsiamError represents an error response from osiam.
type OsiamError struct {
	Error   string
	Message string
}

// ParseOsiamError creates an OsiamError out of a json
func ParseOsiamError(jsonBody []byte) OsiamError {
	m := map[string]interface{}{}
	err := json.Unmarshal(jsonBody, &m)
	if err != nil {
		return OsiamError{
			"client_parse_error",
			"osiam response is no valid json: " + string(jsonBody),
		}
	}
	e := OsiamError{}
	if v, exist := m["error"]; exist {
		if vCasted, ok := v.(string); ok {
			e.Error = vCasted
		}
	}

	if v, exist := m["message"]; exist {
		if vCasted, ok := v.(string); ok {
			e.Message = vCasted
		}
	}

	if v, exist := m["error_description"]; exist {
		if vCasted, ok := v.(string); ok {
			e.Message = vCasted
		}
	}

	if e.Error == "" && e.Message == "" {
		return OsiamError{
			"client_parse_error",
			"not a valid osiam error message: " + string(jsonBody),
		}
	}
	return e
}

// IsLoginError checks, if the grant was invalid
func (e OsiamError) IsLoginError() bool {
	return e.Error == "invalid_grant"
}

// IsUnauthorized checks, if the error reason was Unauthorized
func (e OsiamError) IsUnauthorized() bool {
	return e.Error == "Unauthorized"
}
