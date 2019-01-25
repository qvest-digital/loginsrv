# loginsrv Caddy middleware

Login plugin for Caddy, based on [tarent/loginsrv](https://github.com/tarent/loginsrv).
The login is checked against a backend and then returned as a JWT token.
This middleware is designed to play together with the [caddy-jwt](https://github.com/BTBurke/caddy-jwt) plugin.

For a full documentation of loginsrv configuration and usage, visit the [loginsrv README.md](https://github.com/tarent/loginsrv).

A small demo can be found in the [./demo](https://github.com/tarent/loginsrv/tree/master/caddy/demo) directory.

## Configuration of `JWT_SECRET`
The jwt secret is taken from the environment variable `JWT_SECRET` if this variable is set.
If a secret was configured in the directive config, this has higher priority and will be used over the environment variable in the case,
that both are set. This way, it is also possible to configure different secrets for multiple hosts. If no secret was set at all,
a random token is generated and used.

To be compatible with caddy-jwt the secret is also written to the environment variable JWT_SECRET, if this variable was not set before.
This enables caddy-jwt to look up the same shared secret, even in the case of a random token. If the configuration uses different tokens
for different server blocks, only the first one will be stored in environment variable. You can't use a random key as the jwt-secret
and a custom one in the same caddyfile. If you want to have better control, of the integration with caddy-jwt, e.g. for multiple server blocks,
you should configure the jwt behaviour in caddy-jwt with the `secret` or `publickey` directive.

## Cookie Name
You can configure the cookie name by `cookie_name`. By default loginsrv and http.jwt use the same cookie name for the JWT token. 
If you don't use the default, set related param `token_source cookie my_cookie_name` in http.jwt.

### Basic configuration
Provide a login resource under /login, for user bob with password secret:
```
login {
    simple bob=secret
}
```

### Full configuration example
```
login {
    success_url /after/login
    cookie_name alternativeName
    cookie_http_only true
    simple bob=secret
    osiam endpoint=http://localhost:8080,client_id=example-client,client_secret=secret
    htpasswd file=users
    github client_id=xxx,client_secret=yyy
    google client_id=xxx,client_secret=yyy,scope=email
}
```

### Example caddyfile
```
127.0.0.1

root {$PWD}
browse

jwt {
    path /
    allow sub bob
}

login {
    simple bob=secret,alice=secret
}
```

### Example caddyfile with dynamic redirects
```
127.0.0.1

root {$PWD}
browse

jwt {
    path /
    except /favicon.ico
    redirect /login?backTo={rewrite_uri}
    allow sub bob
    allow sub alice
}

login {
    simple bob=secret,alice=secret
    redirect_check_referer false
    redirect_host_file ../redirect_hosts.txt
}
```

### Example caddyfile with Google login

```
127.0.0.1

root {$PWD}
browse

jwt {
    path /
    allow domain example.com
}

login {
    google client_id=xxx,client_secret=yyy
}
```

### Potential issue with a different `cookie-name` in http.login and `token_source cookie cookie_name` in http.jwt

1. If you use `redirect` in http.jwt and you:
   * Are redirected to http.jwt's `redirect` page that is in your caddyfile
   * Are unable to navigate to any page that is protected by http.login
   * Appear to be authenticated when you visit the `redirect` page or `/login`

2. If you don't use `redirect` in http.jwt and you:
   * Are displayed a 401 error for the page you navigate to
   * Appear to be authenticated if you navigate to `/login`

Possible solution:
Confirm that `cookie-name` in http.login and `token_source cookie cookie_name` in http.jwt are identical