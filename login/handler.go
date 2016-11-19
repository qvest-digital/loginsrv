package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/tarent/lib-compose/logging"
	"io/ioutil"
	"net/http"
	"strings"
)

const contentTypeHtml = "text/html; charset=utf-8"
const contentTypeJWT = "application/jwt"
const contentTypePlain = "text/plain"

type Handler struct {
	backends []Backend
	config   *Config
}

// NewHandler creates a login handler based on the supplied configuration.
func NewHandler(config *Config) (*Handler, error) {
	backends := []Backend{}
	for _, opt := range config.Backends {
		p, exist := GetProvider(opt["provider"])
		if !exist {
			return nil, fmt.Errorf("No such provider: %v", opt["provider"])
		}
		b, err := p(opt)
		if err != nil {
			return nil, err
		}
		backends = append(backends, b)
	}
	if len(backends) == 0 {
		return nil, errors.New("No login backends configured!")
	}
	return &Handler{
		backends: backends,
		config:   config,
	}, nil
}

func (h *Handler) authenticate(username, password string) (bool, UserInfo, error) {
	for _, b := range h.backends {
		authenticated, userInfo, err := b.Authenticate(username, password)
		if err != nil {
			return false, UserInfo{}, err
		}
		if authenticated {
			return authenticated, userInfo, nil
		}
	}
	return false, UserInfo{}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/login") {
		w.WriteHeader(404)
		fmt.Fprintf(w, "404 Ressource not found")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !(r.Method == "GET" ||
		(r.Method == "POST" &&
			(strings.HasPrefix(contentType, "application/json") ||
				strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
				strings.HasPrefix(contentType, "multipart/form-data")))) {
		h.respondBadRequest(w, r)
		return
	}

	r.ParseForm()
	if r.Method == "GET" {
		writeLoginForm(w,
			map[string]interface{}{
				"path":   r.URL.Path,
				"config": h.config,
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
				WithField("username", username).Info("sucessfully authenticated")
			h.respondAuthenticated(w, r, userInfo)
			return
		}
		logging.Application(r.Header).
			WithField("username", username).Info("failed authentication")

		h.respondAuthFailure(w, r)
		return
	}
}

func (h *Handler) respondAuthenticated(w http.ResponseWriter, r *http.Request, userInfo UserInfo) {
	token, err := h.createToken(userInfo)
	if err != nil {
		logging.Application(r.Header).WithError(err).Error()
		h.respondError(w, r)
		return
	}
	if wantHtml(r) {

		// TODO: set livetime
		cookie := &http.Cookie{Name: h.config.CookieName, Value: token, HttpOnly: true}
		http.SetCookie(w, cookie)
		w.Header().Set("Location", h.config.SuccessUrl)
		w.WriteHeader(303)
		return
	}

	w.Header().Set("Content-Type", contentTypeJWT)
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s\n", token)
}

func (h *Handler) createToken(userInfo UserInfo) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, userInfo)
	return token.SignedString([]byte(h.config.JwtSecret))
}

func (h *Handler) respondError(w http.ResponseWriter, r *http.Request) {
	if wantHtml(r) {
		w.Header().Set("Content-Type", contentTypeHtml)
		w.WriteHeader(500)
		username, _, _ := getCredentials(r)
		writeLoginForm(w,
			map[string]interface{}{
				"path":     r.URL.Path,
				"error":    true,
				"config":   h.config,
				"username": username,
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

func (h *Handler) respondAuthFailure(w http.ResponseWriter, r *http.Request) {
	if wantHtml(r) {
		w.Header().Set("Content-Type", contentTypeHtml)
		w.WriteHeader(403)
		username, _, _ := getCredentials(r)
		writeLoginForm(w,
			map[string]interface{}{
				"path":    r.URL.Path,
				"failure": true,
				"config":  h.config,

				"username": username,
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
