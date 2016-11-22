# loginsrv caddy middleware

Login plugin for caddy, based on [tarent/loginsrv](https://github.com/tarent/loginsrv).
The login is checked against a middleware and then returned as JWT token.
This middleware is designed to play together with the [caddy-jwt](https://github.com/BTBurke/caddy-jwt) plugin.

## Configuration
To be compatible with caddy-jwt, the jwt secret is taken from the enviroment variable `JWT_SECRET`
if such a variable is set. Otherwise, a random token is generated and set as enviroment variable JWT_SECRET,
so that caddy-jwt looks up the same shared secret.

### Basic configuration
Providing a login resource unter /login, for user bob with password secret:
```
loginsrv / {
    backend provider=simple,bob=secret
}
```

### Full configuration example
```
loginsrv / {
    success-url /after/login
    cookie-name alternativeName
    cookie-http-only true
    backend provider=simple,bob=secret
    backend provider=osiam,endpoint=http://localhost:8080,clientId=example-client,clientSecret=secret
    backend provider=htpasswd,file=users
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

loginsrv / {
         backend provider=simple,bob=secret,alice=secret
}
```
