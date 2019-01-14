package caddy

import (
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/tarent/loginsrv/login"
	"context"
	"net/http"
	"strings"
)

// CaddyHandler is the loginsrv handler wrapper for caddy
type CaddyHandler struct {
	next         httpserver.Handler
	config       *login.Config
	loginHandler *login.Handler
}

// NewCaddyHandler create the handler
func NewCaddyHandler(next httpserver.Handler, loginHandler *login.Handler, config *login.Config) *CaddyHandler {
	h := &CaddyHandler{
		next:         next,
		config:       config,
		loginHandler: loginHandler,
	}
	return h
}

func (h *CaddyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	//Fetch jwt token. If valid set a Caddy replacer for {user}
	userInfo, valid := h.loginHandler.GetToken(r)
	if valid {
		// let upstream middleware (e.g. fastcgi and cgi) know about authenticated
		// user; this replaces the request with a wrapped instance
		r = r.WithContext(context.WithValue(r.Context(),
		httpserver.RemoteUserCtxKey, userInfo.Sub))
	
		// Provide username to be used in log by replacer
		repl := httpserver.NewReplacer(r, nil, "-")
		repl.Set("user", userInfo.Sub)
	}

	if strings.HasPrefix(r.URL.Path, h.config.LoginPath) {
		h.loginHandler.ServeHTTP(w, r)
		return 0, nil
	}

	return h.next.ServeHTTP(w, r)
}
