# loginsrv caddy middleware

Login plugin for caddy, based on [tarent/loginsrv](https://github.com/tarent/loginsrv).
The login is checked against a backend and then returned as JWT token.
This middleware is designed to play together with the [caddy-jwt](https://github.com/BTBurke/caddy-jwt) plugin.

## Configuration
To be compatible with caddy-jwt, the jwt secret is taken from the enviroment variable `JWT_SECRET`
if such a variable is set. Otherwise, a random token is generated and set as enviroment variable JWT_SECRET,
so that caddy-jwt looks up the same shared secret.

### Basic configuration
Providing a login resource unter /login, for user bob with password secret:
```
login / {
    simple bob=secret
}
```

### Full configuration example
```
login / {
    success-url /after/login
    cookie-name alternativeName
    cookie-http-only true
    simple bob=secret
    osiam endpoint=http://localhost:8080,client_id=example-client,client_secret=secret
    htpasswd file=users
    github client_id=xxx,client_secret=yyy
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

login / {
         simple bob=secret,alice=secret
}
```
