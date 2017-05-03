package caddy

import (
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/tarent/loginsrv/login"
	_ "github.com/tarent/loginsrv/osiam"
	"net/http"
	"path"
	"strings"
)

type CaddyHandler struct {
	next         httpserver.Handler
	path         string
	config       *login.Config
	loginHandler *login.Handler
}

func NewCaddyHandler(next httpserver.Handler, path string, loginHandler *login.Handler, config *login.Config) *CaddyHandler {
	h := &CaddyHandler{
		next:         next,
		path:         path,
		config:       config,
		loginHandler: loginHandler,
	}
	return h
}

func (h *CaddyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if strings.HasPrefix(r.URL.Path, path.Join(h.path, "/login")) {
		h.loginHandler.ServeHTTP(w, r)
		return 0, nil
	} else {
		return h.next.ServeHTTP(w, r)
	}
}
