package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/tarent/lib-compose/logging"
	"github.com/tarent/loginsrv/model"
	"github.com/tarent/loginsrv/oauth2"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const contentTypeHtml = "text/html; charset=utf-8"
const contentTypeJWT = "application/jwt"
const contentTypePlain = "text/plain"

type Handler struct {
	backends []Backend
	oauth    oauthManager
	config   *Config
}

// NewHandler creates a login handler based on the supplied configuration.
func NewHandler(config *Config) (*Handler, error) {
	if len(config.Backends) == 0 && len(config.Oauth) == 0 {
		return nil, errors.New("No login backends or oauth provider configured!")
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

	return &Handler{
		backends: backends,
		config:   config,
		oauth:    oauth,
	}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.config.LoginPath) {
		h.respondNotFound(w, r)
		return
	}

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
			(strings.HasPrefix(contentType, "application/json") ||
				strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
				strings.HasPrefix(contentType, "multipart/form-data")))) {
		h.respondBadRequest(w, r)
		return
	}

	r.ParseForm()
	if r.Method == "DELETE" || r.FormValue("logout") == "true" {
		h.deleteToken(w)
		if h.config.LogoutUrl != "" {
			w.Header().Set("Location", h.config.LogoutUrl)
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
		userInfo, valid := h.getToken(r)
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
		return
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
	http.SetCookie(w, cookie)
}

func (h *Handler) respondAuthenticated(w http.ResponseWriter, r *http.Request, userInfo jwt.Claims) {
	token, err := h.createToken(userInfo)
	if err != nil {
		logging.Application(r.Header).WithError(err).Error()
		h.respondError(w, r)
		return
	}
	if wantHtml(r) {
		// TODO: set livetime
		cookie := &http.Cookie{
			Name:     h.config.CookieName,
			Value:    token,
			HttpOnly: h.config.CookieHttpOnly,
			Path:     "/",
		}
		http.SetCookie(w, cookie)
		w.Header().Set("Location", h.config.SuccessUrl)
		w.WriteHeader(303)
		return
	}

	w.Header().Set("Content-Type", contentTypeJWT)
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", token)
}

func (h *Handler) createToken(userInfo jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, userInfo)
	return token.SignedString([]byte(h.config.JwtSecret))
}

func (h *Handler) getToken(r *http.Request) (userInfo model.UserInfo, valid bool) {
	c, err := r.Cookie(h.config.CookieName)
	if err != nil {
		return model.UserInfo{}, false
	}

	token, err := jwt.ParseWithClaims(c.Value, &model.UserInfo{}, func(*jwt.Token) (interface{}, error) {
		return []byte(h.config.JwtSecret), nil
	})
	if err != nil {
		return model.UserInfo{}, false
	}

	u, v := token.Claims.(*model.UserInfo)
	return *u, v
}

func (h *Handler) respondError(w http.ResponseWriter, r *http.Request) {
	if wantHtml(r) {
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

func (h *Handler) respondAuthFailure(w http.ResponseWriter, r *http.Request) {
	if wantHtml(r) {
		w.Header().Set("Content-Type", contentTypeHtml)
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
	w.Header().Set("Content-Type", contentTypePlain)
	w.WriteHeader(403)
	fmt.Fprintf(w, "Wrong credentials")
}

func wantHtml(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

func getCredentials(r *http.Request) (string, string, error) {
	if r.Header.Get("Content-Type") == "application/json" {
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
