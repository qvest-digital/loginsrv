# loginsrv Caddy middleware

Login plugin for Caddy, based on [tarent/loginsrv](https://github.com/tarent/loginsrv).
The login is checked against a backend and then returned as JWT token.
This middleware is designed to play together with the [caddy-jwt](https://github.com/BTBurke/caddy-jwt) plugin.

For a full documentation of loginsrv configuration and usage, visit the [loginsrv README.md](https://github.com/tarent/loginsrv).

A small demo can also be found in the [./demo](https://github.com/tarent/loginsrv/tree/master/caddy/demo) directory.

## Configuration of `JWT_SECRET`
The jwt secret is taken from the environment variable `JWT_SECRET` if such a variable is set.
If a secret was configured in the directive config, this has higher priority and will be used over the environment variable in the case,
that both are set. This way, it is also possible to configure different secrets for multiple hosts. If no secret was set at all,
a random token is generated and used.

To be compatible with caddy-jwt the secred is also written to the environment variable JWT_SECRET, if this variable was not set before.
This enables caddy-jwt to look up the same shared secret, even in the case of a random token. If the configuration uses differnt tokens
for different server blocks, only the first one will be stored in ent environment variable.

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
    google client_id=xxx,client_secret=yyy,scope=email
}
```
Note: You must enable Google Plus API in Google API Console
