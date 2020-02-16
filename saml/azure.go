package saml

import (
	"encoding/base64"
	"fmt"
	samllib "github.com/crewjam/saml"
	"github.com/tarent/loginsrv/logging"
	"github.com/tarent/loginsrv/model"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	RegisterProvider(providerAzureAD)
}

var providerAzureAD = Provider{
	Name: "azure",
}

// Authenticate parses and validates SAML Response originating
// at Azure AD.
func Authenticate(sp *samllib.ServiceProvider, r *http.Request) (model.UserInfo, error) {
	user := model.UserInfo{
		Expiry: time.Now().Add(time.Duration(900) * time.Second).Unix(),
	}
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		return user, fmt.Errorf("The Azure AD authorization POST request is not application/x-www-form-urlencoded")
	}
	if r.FormValue("SAMLResponse") == "" {
		return user, fmt.Errorf("The Azure AD authorization POST request has no SAMLResponse")
	}
	samlpRespRaw, err := base64.StdEncoding.DecodeString(r.FormValue("SAMLResponse"))
	if err != nil {
		return user, fmt.Errorf("The Azure AD authorization POST request with SAMLResponse failed base64 decoding: %s", err)
	}

	samlAssertions, err := sp.ParseXMLResponse(samlpRespRaw, []string{""})
	if err != nil {
		return user, fmt.Errorf("The Azure AD validation failure: %s", err)
	}

	for _, attrStatement := range samlAssertions.AttributeStatements {
		for _, attrEntry := range attrStatement.Attributes {
			if len(attrEntry.Values) == 0 {
				continue
			}
			if strings.HasSuffix(attrEntry.Name, "Attributes/MaxSessionDuration") {
				multiplier, err := strconv.Atoi(attrEntry.Values[0].Value)
				if err != nil {
					logging.Logger.Errorf("Failed parsing Attributes/MaxSessionDuration: %v", err)
					continue
				}
				user.Expiry = time.Now().Add(time.Duration(multiplier) * time.Second).Unix()
				continue
			}

			if strings.HasSuffix(attrEntry.Name, "identity/claims/displayname") {
				user.Name = attrEntry.Values[0].Value
				continue
			}

			if strings.HasSuffix(attrEntry.Name, "identity/claims/emailaddress") {
				user.Email = attrEntry.Values[0].Value
				continue
			}

			if strings.HasSuffix(attrEntry.Name, "identity/claims/identityprovider") {
				user.Origin = attrEntry.Values[0].Value
				continue
			}

			if strings.HasSuffix(attrEntry.Name, "identity/claims/name") {
				user.Sub = attrEntry.Values[0].Value
				continue
			}

			if strings.HasSuffix(attrEntry.Name, "Attributes/Role") {
				for _, attrEntryElement := range attrEntry.Values {
					user.Groups = append(user.Groups, attrEntryElement.Value)
				}
				continue
			}
		}
	}

	if user.Email == "" || user.Name == "" || len(user.Groups) == 0 {
		return user, fmt.Errorf("The Azure AD authorization failed, mandatory attributes not found: %v", user)
	}

	return user, nil
}
