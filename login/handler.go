package login

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/tarent/loginsrv/logging"
	"github.com/tarent/loginsrv/model"
	"github.com/tarent/loginsrv/oauth2"
)

const contentTypeHTML = "text/html; charset=utf-8"
const contentTypeJWT = "application/jwt"
const contentTypeJSON = "application/json"
const contentTypePlain = "text/plain"

type userClaimsFunc func(userInfo model.UserInfo) (jwt.Claims, error)

// Handler is the mail login handler.
// It serves the login ressource and does the authentication against the backends or oauth provider.
type Handler struct {
	backends         []Backend
	oauth            oauthManager
	config           *Config
	signingMethod    jwt.SigningMethod
	signingKey       interface{}
	signingVerifyKey interface{}
	userClaims       userClaimsFunc
}

// NewHandler creates a login handler based on the supplied configuration.
func NewHandler(config *Config) (*Handler, error) {
	if len(config.Backends) == 0 && len(config.Oauth) == 0 {
		return nil, errors.New("No login backends or oauth provider configured")
	}

	backends := []Backend{}
	for pName, opts := range config.Backends {
		p, exist := GetProvider(pName)
		if !exist {
			return nil, fmt.Errorf("No such provider: %v", pName)
		}
		b, err := p(opts)
		if err != nil {
			return nil, err
		}
		backends = append(backends, b)
	}

	oauth := oauth2.NewManager()
	for providerName, opts := range config.Oauth {
		err := oauth.AddConfig(providerName, opts)
		if err != nil {
			return nil, err
		}
	}

	userClaims, err := NewUserClaims(config)
	if err != nil {
		return nil, err
	}

	return &Handler{
		backends:   backends,
		config:     config,
		oauth:      oauth,
		userClaims: userClaims.Claims,
	}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.config.LoginPath) {
		h.respondNotFound(w, r)
		return
	}

	h.setRedirectCookie(w, r)

	_, err := h.oauth.GetConfigFromRequest(r)
	if err == nil {
		h.handleOauth(w, r)
		return
	}

	h.handleLogin(w, r)
	return
}

func (h *Handler) handleOauth(w http.ResponseWriter, r *http.Request) {
	startedFlow, authenticated, userInfo, err := h.oauth.Handle(w, r)

	if startedFlow {
		// the oauth flow started
		return
	}

	if err != nil {
		logging.Application(r.Header).WithError(err).Error()
		h.respondError(w, r)
		return
	}

	if authenticated {
		logging.Application(r.Header).
			WithField("username", userInfo.Sub).Info("successfully authenticated")
		h.respondAuthenticated(w, r, userInfo)
		return
	}
	logging.Application(r.Header).
		WithField("username", userInfo.Sub).Info("failed authentication")

	h.respondAuthFailure(w, r)
	return
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !(r.Method == "GET" || r.Method == "DELETE" ||
		(r.Method == "POST" &&
			(strings.HasPrefix(contentType, contentTypeJSON) ||
				strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
				strings.HasPrefix(contentType, "multipart/form-data") ||
				contentType == ""))) {
		h.respondBadRequest(w, r)
		return
	}

	r.ParseForm()
	if r.Method == "DELETE" || r.FormValue("logout") == "true" {
		h.deleteToken(w)
		if h.config.LogoutURL != "" {
			w.Header().Set("Location", h.config.LogoutURL)
			w.WriteHeader(303)
			return
		}
		writeLoginForm(w,
			loginFormData{
				Config: h.config,
			})
		return
	}

	if r.Method == "GET" {
		userInfo, valid := h.GetToken(r)
		if wantJSON(r) {
			if valid {
				w.Header().Set("Content-Type", contentTypeJSON)
				enc := json.NewEncoder(w)
				enc.Encode(userInfo) // ignore error of encoding
			} else {
				h.respondAuthFailure(w, r)
			}
			return
		}
		writeLoginForm(w,
			loginFormData{
				Config:        h.config,
				Authenticated: valid,
				UserInfo:      userInfo,
			})
		return
	}

	if r.Method == "POST" {
		username, password, err := getCredentials(r)
		if err != nil {
			h.respondBadRequest(w, r)
			return
		}
		if username != "" {
			// No token found or credentials found, assuming new authentication
			h.handleAuthentication(w, r, username, password)
			return
		}
		userInfo, valid := h.GetToken(r)
		if valid {
			h.handleRefresh(w, r, userInfo)
			return
		}
		if username == "" {
			h.respondAuthFailure(w, r)
			return
		}

		h.respondBadRequest(w, r)
		return
	}
}

func (h *Handler) handleAuthentication(w http.ResponseWriter, r *http.Request, username string, password string) {
	authenticated, userInfo, err := h.authenticate(username, password)
	if err != nil {
		logging.Application(r.Header).WithError(err).Error()
		h.respondError(w, r)
		return
	}

	if authenticated {
		logging.Application(r.Header).
			WithField("username", username).Info("successfully authenticated")
		h.respondAuthenticated(w, r, userInfo)
		return
	}
	logging.Application(r.Header).
		WithField("username", username).Info("failed authentication")

	h.respondAuthFailure(w, r)
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request, userInfo model.UserInfo) {
	if userInfo.Refreshes >= h.config.JwtRefreshes {
		h.respondMaxRefreshesReached(w, r)
	} else {
		userInfo.Refreshes++
		h.respondAuthenticated(w, r, userInfo)
		logging.Application(r.Header).WithField("username", userInfo.Sub).Info("refreshed jwt")
	}
}

func (h *Handler) deleteToken(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     h.config.CookieName,
		Value:    "delete",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		Path:     "/",
	}
	if h.config.CookieDomain != "" {
		cookie.Domain = h.config.CookieDomain
	}
	http.SetCookie(w, cookie)
}

func (h *Handler) respondAuthenticated(w http.ResponseWriter, r *http.Request, userInfo model.UserInfo) {
	userInfo.Expiry = time.Now().Add(h.config.JwtExpiry).Unix()
	token, err := h.createToken(userInfo)
	if err != nil {
		logging.Application(r.Header).WithError(err).Error()
		h.respondError(w, r)
		return
	}

	if wantHTML(r) {
		cookie := &http.Cookie{
			Name:     h.config.CookieName,
			Value:    token,
			HttpOnly: h.config.CookieHTTPOnly,
			Path:     "/",
		}
		if h.config.CookieExpiry != 0 {
			cookie.Expires = time.Now().Add(h.config.CookieExpiry)
		}
		if h.config.CookieDomain != "" {
			cookie.Domain = h.config.CookieDomain
		}

		http.SetCookie(w, cookie)

		w.Header().Set("Location", h.redirectURL(r, w))
		h.deleteRedirectCookie(w, r)
		w.WriteHeader(303)
		return
	}

	w.Header().Set("Content-Type", contentTypeJWT)
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", token)
}

func (h *Handler) createToken(userInfo model.UserInfo) (string, error) {
	var claims jwt.Claims = userInfo
	if h.userClaims != nil {
		var err error
		claims, err = h.userClaims(userInfo)
		if err != nil {
			return "", err
		}
	}

	signingMethod, key, _, err := h.signingInfo()
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(signingMethod, claims)
	return token.SignedString(key)
}

func (h *Handler) GetToken(r *http.Request) (userInfo model.UserInfo, valid bool) {
	c, err := r.Cookie(h.config.CookieName)
	if err != nil {
		return model.UserInfo{}, false
	}

	token, err := jwt.ParseWithClaims(c.Value, &model.UserInfo{}, func(*jwt.Token) (interface{}, error) {
		_, _, verifyKey, err := h.signingInfo()
		return verifyKey, err
	})
	if err != nil {
		return model.UserInfo{}, false
	}

	u, ok := token.Claims.(*model.UserInfo)
	if !ok {
		return model.UserInfo{}, false
	}

	return *u, u.Valid() == nil
}

func (h *Handler) signingInfo() (signingMethod jwt.SigningMethod, key, verifyKey interface{}, err error) {
	if h.signingMethod == nil || h.signingKey == nil || h.signingVerifyKey == nil {
		h.signingMethod = jwt.GetSigningMethod(h.config.JwtAlgo)
		if h.signingMethod == nil {
			return nil, nil, nil, errors.New("invalid signing method: " + h.config.JwtAlgo)
		}

		keyString := h.config.JwtSecret
		switch h.config.JwtAlgo {
		case "ES256", "ES384", "ES512":
			if !strings.Contains(string(keyString), "-----") {
				keyString = "-----BEGIN EC PRIVATE KEY-----\n" + keyString + "\n-----END EC PRIVATE KEY-----"
			}

			key, err := jwt.ParseECPrivateKeyFromPEM([]byte(keyString))
			if err != nil {
				return nil, nil, nil, errors.Wrap(err, "can not parse PEM formated EC private key")
			}
			h.signingKey = key
			h.signingVerifyKey = key.Public()
		default:
			h.signingKey = []byte(keyString)
			h.signingVerifyKey = h.signingKey
		}
	}
	return h.signingMethod, h.signingKey, h.signingVerifyKey, nil
}

func (h *Handler) respondError(w http.ResponseWriter, r *http.Request) {
	if wantHTML(r) {
		username, _, _ := getCredentials(r)
		writeLoginForm(w,
			loginFormData{
				Error:    true,
				Config:   h.config,
				UserInfo: model.UserInfo{Sub: username},
			})
		return
	}
	w.Header().Set("Content-Type", contentTypePlain)
	w.WriteHeader(500)
	fmt.Fprintf(w, "Internal Server Error")
}

func (h *Handler) respondBadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	fmt.Fprintf(w, "Bad Request: Method or content-type not supported")
}

func (h *Handler) respondNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprintf(w, "Not Found: The requested page does not exist")
}

func (h *Handler) respondMaxRefreshesReached(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(403)
	fmt.Fprint(w, "Max JWT refreshes reached")
}

func (h *Handler) respondAuthFailure(w http.ResponseWriter, r *http.Request) {
	if wantHTML(r) {
		w.Header().Set("Content-Type", contentTypeHTML)
		w.WriteHeader(403)
		username, _, _ := getCredentials(r)
		writeLoginForm(w,
			loginFormData{
				Failure:  true,
				Config:   h.config,
				UserInfo: model.UserInfo{Sub: username},
			})
		return
	}

	if wantJSON(r) {
		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(403)
		fmt.Fprintf(w, `{"error": "Wrong credentials"}`)
	} else {
		w.Header().Set("Content-Type", contentTypePlain)
		w.WriteHeader(403)
		fmt.Fprintf(w, "Wrong credentials")
	}

}

func wantHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

func wantJSON(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), contentTypeJSON)
}

func getCredentials(r *http.Request) (string, string, error) {
	if r.Header.Get("Content-Type") == contentTypeJSON {
		m := map[string]string{}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return "", "", err
		}
		err = json.Unmarshal(body, &m)
		if err != nil {
			return "", "", err
		}
		return m["username"], m["password"], nil
	}
	return r.PostForm.Get("username"), r.PostForm.Get("password"), nil
}

func (h *Handler) authenticate(username, password string) (bool, model.UserInfo, error) {
	for _, b := range h.backends {
		authenticated, userInfo, err := b.Authenticate(username, password)
		if err != nil {
			return false, model.UserInfo{}, err
		}
		if authenticated {
			return authenticated, userInfo, nil
		}
	}
	return false, model.UserInfo{}, nil
}

type oauthManager interface {
	Handle(w http.ResponseWriter, r *http.Request) (
		startedFlow bool,
		authenticated bool,
		userInfo model.UserInfo,
		err error)
	AddConfig(providerName string, opts map[string]string) error
	GetConfigFromRequest(r *http.Request) (oauth2.Config, error)
}
